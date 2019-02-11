package decoder

//#include <stdio.h>
//static void pgm_save(unsigned char *buf, int wrap, int xsize, int ysize)
//{
//    FILE *f;
//    int i;
//
//    f=fopen("ImageTest","w");
//     fprintf(f,"P5\n%d %d\n%d\n",xsize,ysize,255);
//    for(i=0;i<ysize;i++)
//         fwrite(buf + i * wrap,1,xsize,f);
//    fclose(f);
//}
import "C"
import (
	"github.com/baohavan/go-libav/avcodec"
	"github.com/baohavan/go-libav/avutil"
	"github.com/golang/glog"
	"sync"
)

const (
	ControlOutputStop = 0
)

type Decoder struct {
	ctx 			*StreamContext
	CamUuid 		string
	outputChan 		chan int
	wait       		sync.WaitGroup
}

func (decoder *Decoder) Init(){
	decoder.wait = sync.WaitGroup{}
}


func (decoder *Decoder) decodeFrame(pkt *avcodec.Packet) error {
	frame, err := avutil.NewFrame()
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
		glog.Info("Save Frame")
		C.pgm_save((*C.uchar)(frame.Data(0)), (C.int)(frame.LineSize(0)),
			(C.int)(decoder.ctx.InCodecCtx.Width()), (C.int)(decoder.ctx.InCodecCtx.Height()))
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

func (decoder *Decoder) Stop()  {
	glog.Info("Stopping stream decode")
	decoder.wait.Add(1)
	decoder.outputChan <- ControlOutputStop
	decoder.wait.Wait()
	glog.Info("Stopping stream decode done")
}