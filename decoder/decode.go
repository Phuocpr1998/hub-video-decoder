package decoder

//#include <stdio.h>
//static void Save_Frame(unsigned char *buf, int wrap, int width, int height, char* filename)
//{
//    FILE *f;
//    int i;
//
//    f=fopen(filename,"w");
//	  fprintf(f,"P6\n%d %d\n%d\n", width, height, 255);
//    for(i = 0; i < height; i++)
//         fwrite(buf + i * wrap, 1, width*3, f);
//	  fclose(f);
//}
import "C"
import (
	"fmt"
	"github.com/baohavan/go-libav/avcodec"
	"github.com/baohavan/go-libav/avutil"
	"github.com/golang/glog"
	"hub-video-decoder/api"
	"hub-video-decoder/swscale"
	"hub-video-decoder/utils"
	"strconv"
	"sync"
	"time"
)

const PathSaveImage = "/tmp/images"

const (
	ControlOutputStop = 0
)

type Decoder struct {
	ctx         *StreamContext
	Sws_Context *swscale.SwsContext
	CamUuid     string
	outputChan  chan int
	wait        sync.WaitGroup
	Width       int
	Height      int
}

func (decoder *Decoder) Init() {
	decoder.wait = sync.WaitGroup{}
}

func (decoder *Decoder) Free() {
	decoder.Sws_Context.Free()
	decoder.Sws_Context = nil
}

func (decoder *Decoder) decodeFrame(pkt *avcodec.Packet) error {
	frame, err := avutil.NewFrame()
	frameRGB, err := avutil.NewFrame()
	if err != nil {
		glog.Error("Cann't allocate frame")
		return err
	}

	gotFrame, _, err := decoder.ctx.InCodecCtx.DecodeVideo(pkt, frame)
	if err != nil {
		glog.Error("Error while decoding frame")
		pkt.Free()
		return err
	}

	if gotFrame {
		go func() {
			// convert image
			buffer := swscale.AllocateBuffer(decoder.Width, decoder.Height)
			swscale.AVPictureFill(frameRGB, buffer, decoder.Width, decoder.Height)
			swscale.Sws_scale(decoder.Sws_Context, frame, frameRGB, decoder.Height)

			filename := fmt.Sprintf("%s/%s/%s", PathSaveImage, decoder.CamUuid, strconv.Itoa((int)(time.Now().UnixNano())))
			C.Save_Frame((*C.uchar)(frameRGB.Data(0)), (C.int)(frameRGB.LineSize(0)),
				(C.int)(decoder.Width), (C.int)(decoder.Height),
				(C.CString)(filename))

			frameRGB.Free()
			frame.Free()
			swscale.FreeBuffer(buffer)

			stren, err := utils.Base64Encoder(filename)
			if err != nil {
				glog.Info(err)
			} else {
				api.PostImage(filename, decoder.CamUuid, ([]byte)(stren))
			}
		}()
	}

	pkt.Free()
	return nil
}

func (decoder *Decoder) Run() {
	defer decoder.wait.Done()
	for {
		select {
		case pkt := <-decoder.ctx.packetChan:
			err := decoder.decodeFrame(pkt)
			if err != nil {
				glog.Error("Failed to decode frame")
			}
		case ctl := <-decoder.outputChan:
			switch ctl {
			case ControlOutputStop:
				glog.Infof("Quit process output => Cleanup %d packet queue", len(decoder.ctx.packetChan))
				if len(decoder.ctx.packetChan) > 0 {
					select {
					case pkt := <-decoder.ctx.packetChan:
						{
							pkt.Free()
							pkt = nil
						}
					default:
						glog.Info("Cleanup packet queue done")
					}
				}
				return
			}
		}
	}
}

func (decoder *Decoder) Stop() {
	glog.Info("Stopping stream decode")
	decoder.wait.Add(1)
	decoder.outputChan <- ControlOutputStop
	decoder.wait.Wait()
	glog.Info("Stopping stream decode done")
}
