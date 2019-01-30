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
	"hub-video-decoder/input"
)

type Decoder struct {
	ctx *input.StreamContext
}

func (decoder Decoder) decodeFrame(pkt avcodec.Packet) error {
	frame, err := avutil.NewFrame()
	if err != nil {
		glog.Error("Cann't allocate frame")
		return err
	}

	gotFrame, _, err := decoder.ctx.InCodecCtx.DecodeVideo(pkt, frame)
	if err != nil {
		glog.Error("Error while decoding frame")
		return err
	}

	if gotFrame {
		glog.Info("Save Frame")
		C.pgm_save(frame.Data(0), frame.LineSize(0), decoder.ctx.InCodecCtx.Width(), decoder.ctx.InCodecCtx.Height())
	}

	return nil
}
