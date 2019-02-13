package decoder

import (
	"errors"
	"github.com/baohavan/go-libav/avcodec"
	"github.com/baohavan/go-libav/avformat"
	"github.com/baohavan/go-libav/avutil"
	"github.com/golang/glog"
	"hub-video-decoder/config"
	"time"
)

const (
	ControlInputQuit = 0
)

/*
Read frame add to queue
*/
type StreamInput struct {
	control   chan int
	ctx       *StreamContext
	init      bool
	idle      bool
	nextRetry time.Time
}

func (si *StreamInput) Init() error {
	si.control = make(chan int)
	return nil
}

func (si *StreamInput) initialize() error {
	return si.setupInput()
}

func (si *StreamInput) SafeInit() error {
	if !si.init {
		if !si.idle {
			return si.initialize()
		} else if time.Now().After(si.nextRetry) {
			glog.Info("Input init after idle")
			return si.initialize()
		} else {
			time.Sleep(config.IdleSleep)
			return errors.New("input.error.notinit")
		}
	}

	return nil
}

func (si *StreamInput) Idle() error {
	if si.idle {
		glog.Warning("Input already idled")
		return errors.New("input.idle.already")
	}

	si.nextRetry = time.Now().Add(config.IdleRetryOpenInput)
	si.Uninitialize()
	si.idle = true

	glog.Infof("Input idled %s", si.ctx.InputFileName)

	return nil
}

func (si *StreamInput) Uninitialize() {
	glog.Infof("Uninitialize %s", si.ctx.InputFileName)
	if si.init {
		si.ctx.inFmtCtx.CloseInput()
		si.ctx.inFmtCtx.Free()
		si.ctx.InCodecCtx.Close()
		si.ctx.InCodecCtx.Free()
		si.ctx.InCodecCtx = nil
		si.ctx.inFmtCtx = nil
		si.init = false
	}
}

func (si *StreamInput) Free() {
	si.Uninitialize()
}

func (si *StreamInput) Stop() {
	glog.Infof("Input stop %s close channel", si.ctx.InputFileName)
	si.control <- ControlInputQuit
}

func (si *StreamInput) setupInput() error {
	options := avutil.NewDictionary()
	defer options.Free()

	options.Set("rtsp_flags", "prefer_tcp")
	options.Set("max_delay", "1000000")
	options.Set("stimeout", "1000000")
	options.Set("fflags", "nobuffer")
	options.Set("analyzeduration", "1000000")
	options.Set("probesize", "1000000")
	options.Set("-c:v", "cedrus264")

	si.ctx.inFmtCtx, _ = avformat.NewContextForInput()
	//r.inFmtCtx,_ = avformat.NewContextForInput()
	// open file for decoding
	glog.Infof("Input opening: %s", si.ctx.InputFileName)
	if err := si.ctx.inFmtCtx.OpenInput(si.ctx.InputFileName, nil, options); err != nil {
		si.ctx.inFmtCtx.Free()
		si.ctx.inFmtCtx = nil
		glog.Errorf("Failed to open input file: %s, err %v ", si.ctx.InputFileName, err)
		return errors.New("error open input " + si.ctx.InputFileName)
	}

	// initialize context with stream information
	if err := si.ctx.inFmtCtx.FindStreamInfo(nil); err != nil {
		glog.Errorf("Failed to find stream info: %s\n", err)
		return errors.New("error find stream")
	}
	// dump streams to standard output

	glog.Info("Done input open")

	for i := uint(0); i < si.ctx.inFmtCtx.NumberOfStreams(); i++ {
		inStream := si.ctx.inFmtCtx.Streams()[i]
		// Skip all others frame type
		if inStream.CodecContext().CodecType() != avutil.MediaTypeVideo {
			continue
		}
		si.ctx.Index = i
	}

	// find video decoder
	codec := avcodec.FindDecoderByID(si.ctx.inFmtCtx.Streams()[si.ctx.Index].CodecContext().CodecID())
	if codec == nil {
		glog.Info("Failed to find codec")
		return errors.New("Failed to find codec")
	} else {
		codeCtx, err := avcodec.NewContextWithCodec(codec)
		if err != nil {
			glog.Info("Failed to allocate video codec")
			return errors.New("Failed to allocate video codec")
		} else {
			err := codeCtx.OpenWithCodec(codec, nil)
			if err != nil {
				glog.Info("Cannot open codec")
				return errors.New("Cannot open codec")
			} else {
				si.ctx.InCodecCtx = codeCtx
				glog.Info("Done to open decoder")
			}
		}
	}

	si.init = true
	si.idle = false

	return nil
}

func (si *StreamInput) Run() error {
	//glog.Info("Input begin process")
	packet, err := avcodec.NewPacket()
	if err != nil {
		glog.Errorf("Failed to alloc packet: %v", err)
		return errors.New("input.error.alloc.packet")
	}

	reading, err := si.ctx.inFmtCtx.ReadFrame(packet)
	if err != nil {
		glog.Errorf("Failed to read packet: %v", err)
		packet.Free()
		packet = nil
		return errors.New("input.error.read.frame")
	}

	if !reading {
		glog.Errorf("Input error while reading => Switch to idle")
		packet.Free()
		packet = nil
		return errors.New("input.error.reading")
	}

	index := uint(packet.StreamIndex())
	//r.index ignore index of stream not type video,
	//Sometime cause panic if not check index (stream with sound)
	if si.ctx.Index == index {
		select {
		case si.ctx.packetChan <- packet:
			{

			}
		default:
			{
				packet.Free()
				packet = nil
			}
		}
	} else {
		packet.Free()
		packet = nil
	}

	return nil
}

func (si *StreamInput) openPacketChan() {
	if !si.ctx.openChan {
		si.ctx.packetChan = make(chan *avcodec.Packet, 64)
		si.ctx.openChan = true
	}
}

func (si *StreamInput) closePacketChan() {
	if si.ctx.openChan {
		si.ctx.openChan = false
		close(si.ctx.packetChan)
		si.ctx.packetChan = nil
	}
}

func IsKeyFrame(packet *avcodec.Packet) bool {
	return (packet.Flags() & avcodec.PacketFlagKey) != 0
}
