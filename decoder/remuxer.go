package decoder

import (
	"errors"
	"fmt"
	"github.com/baohavan/go-libav/avcodec"
	"github.com/baohavan/go-libav/avformat"
	"github.com/golang/glog"
	"hub-video-decoder/config"
	"hub-video-decoder/models"
	"hub-video-decoder/swscale"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	ControlQuit = 0
)

type Remuxer struct {
	Name            string
	CamUuid         string
	RequestStop     bool
	Running         bool
	openMessageChan bool
	WaitInput       sync.WaitGroup
	CfgHash         string
	quit            chan int
	control         chan int
	outputChan      chan int

	Width  int
	Height int

	Ctx         StreamContext
	input       StreamInput
	decoder     Decoder
	requestStop bool

	// For watch dog
	notOKCount int
	hangCount  int
}

func Initialize() error {
	avformat.RegisterAll()
	avcodec.RegisterAll()
	os.RemoveAll(PathSaveImage)
	return nil
}

func (r *Remuxer) Run() error {
	r.quit = make(chan int)
	r.control = make(chan int)
	r.outputChan = make(chan int)
	r.requestStop = false

	// For watch dog
	r.notOKCount = 0
	go func() {
		r.publish()
	}()

	return nil
}

func (r *Remuxer) publish() {
	glog.Info("Start publish")

	r.WaitInput = sync.WaitGroup{}

	r.input = StreamInput{
		ctx:  &r.Ctx,
		idle: false,
	}

	r.input.Init()

	defer r.input.Free()

	r.decoder = Decoder{
		ctx:              &r.Ctx,
		CamUuid:          r.CamUuid,
		outputChan:       r.outputChan,
		hasFirstKeyFrame: false,
		isInit:           false,
	}
	r.decoder.Init()
	defer r.decoder.Free()

	r.processInput()

	glog.Info("Wait for stream finish")

	select {
	case control := <-r.control:
		switch control {
		case ControlQuit:
			break
		}
	}

	glog.Info("Done wait stream finish")
}

func (r *Remuxer) processInput() {
	go func() {
		storagePath := fmt.Sprintf("%s/%s", PathSaveImage, r.CamUuid)
		if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
			glog.Infof("Failed to create folder %s err: %v ", storagePath, err)
			return
		}

		glog.Info("R+: Start process input")
		defer r.WaitInput.Done()
		for !r.requestStop {
			glog.Infof("Input setup %s", r.Ctx.InputFileName)
			r.Ctx.InputFileName = r.Ctx.InputFileNameNew

			if err := r.input.SafeInit(); err == nil {
				glog.Infof("Input ready %s => start decode", r.Ctx.InputFileName)

				r.Width = r.Ctx.inFmtCtx.StreamAt(r.Ctx.Index).CodecContext().Width()
				r.Height = r.Ctx.inFmtCtx.StreamAt(r.Ctx.Index).CodecContext().Height()
				r.input.openPacketChan()

				err = r.processOutput()
				if err != nil {
					continue
				}

				glog.Info("Input read frame loop")
				for !r.requestStop {
					if err := r.input.Run(); err != nil {
						break
					}
				}

				glog.Infof("Input finish stop output %s", r.Ctx.InputFileName)
				r.decoder.Stop()
				glog.Infof("Input finish stop output done %s", r.Ctx.InputFileName)
				r.input.closePacketChan()
				r.input.Idle()

				runtime.GC()
			} else {
				if r.input.Idle() != nil {
					glog.Infof("Input idle sleep")
					time.Sleep(config.IdleSleep)
				}
			}
		}
		glog.Infof("R-: Done process input %s", r.Ctx.InputFileName)
	}()
}

func (r *Remuxer) processOutput() error {
	if r.Width <= 0 || r.Height <= 0 {
		glog.Info("Resolution not available", r.Width, "x", r.Height)
		return errors.New("Resolution not available")
	} else {
		r.decoder.Height = r.Height
		r.decoder.Width = r.Width
		glog.Info("Resolution: ", r.Width, "x", r.Height)
	}

	r.decoder.Sws_Context = swscale.GetSwsContext(r.Width, r.Height, r.Ctx.inFmtCtx.StreamAt(r.Ctx.Index).CodecContext().PixelFormat())

	if r.decoder.Sws_Context.CSwsContext == nil {
		glog.Info("Cannot get sws context")
		return errors.New("Cannot get sws context")
	} else {
		r.decoder.isInit = true
		glog.Info("Get sws context done")
	}

	go func() {
		glog.Infof("R+: Start process output decode %s", r.Ctx.InputFileName)
		r.decoder.Run()
		glog.Infof("R-: Done process output decode %s", r.Ctx.InputFileName)
	}()
	return nil
}

func (r *Remuxer) Restart() error {
	r.Stop()
	r.Run()
	return nil
}

func (r *Remuxer) Stop() error {
	glog.Infof("Stop remuxer %s", r.Ctx.InputFileName)

	if !r.requestStop {
		r.WaitInput.Add(1)
		r.requestStop = true
		r.WaitInput.Wait()
	}

	glog.Infof("Stop remuxer done %s", r.Ctx.InputFileName)

	r.control <- ControlQuit
	glog.Infof("Quit remuxer %s", r.Ctx.InputFileName)
	close(r.control)
	return nil
}

func (r *Remuxer) dumpInfo() {
	glog.Infof("Stream config name %s, running %t\n", r.Name, r.Running)
}

func (r *Remuxer) Update(info models.BaseConfigInfo) {
	// Check open/close input
	r.input.ctx.InputFileNameNew = info.Input
}

func (r *Remuxer) InputStatus() int {
	if r.input.init {
		return 0
	} else {
		return 1
	}
}

func (r *Remuxer) StreamWidth() int {
	return r.Width
}

func (r *Remuxer) StreamHeight() int {
	return r.Height
}
