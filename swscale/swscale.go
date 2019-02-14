package swscale

//#include <libavformat/avformat.h>
//#include <libswscale/swscale.h>
//#include <libavutil/avutil.h>
//#include <libavcodec/avcodec.h>
//#cgo pkg-config: libavformat libswscale libavutil libavcodec
import "C"
import (
	"github.com/baohavan/go-libav/avutil"
	"unsafe"
)

const CODE_PIX_FMT_RGB24 = C.AV_PIX_FMT_RGB24

type SwsContext struct {
	CSwsContext *C.struct_SwsContext
}

func (swsctx *SwsContext) Free() {
	if swsctx.CSwsContext != nil {
		C.sws_freeContext(swsctx.CSwsContext)
	}
}

func GetSwsContext(width int, height int, pix_fmt avutil.PixelFormat) *SwsContext {
	var sContext *C.struct_SwsContext
	sContext = C.sws_getContext((C.int)(width), (C.int)(height), (C.enum_AVPixelFormat)(pix_fmt),
		(C.int)(width), (C.int)(height), CODE_PIX_FMT_RGB24, C.SWS_BILINEAR,
		(*C.struct_SwsFilter)(C.NULL), (*C.struct_SwsFilter)(C.NULL), (*C.double)(C.NULL))
	return &SwsContext{sContext}
}

func AVPictureFill(pFrameRGB *avutil.Frame, buffer unsafe.Pointer, width int, height int) {
	frame := unsafe.Pointer(pFrameRGB.CAVFrame)
	C.avpicture_fill((*C.AVPicture)(frame), (*C.uint8_t)(buffer), CODE_PIX_FMT_RGB24, (C.int)(width), (C.int)(height))
}

func AllocateBuffer(width int, height int) (buffer unsafe.Pointer) {
	var numBytes C.int
	numBytes = C.avpicture_get_size(CODE_PIX_FMT_RGB24, (C.int)(width), (C.int)(height))
	size := (C.int)(numBytes) * C.sizeof_uint8_t
	buffer = unsafe.Pointer((*C.uint8_t)(C.malloc((C.uint)(size))))
	return
}

func FreeBuffer(buffer unsafe.Pointer) {
	C.free(buffer)
}

func Sws_scale(swsContext *SwsContext, pFrame *avutil.Frame, pFrameRGB *avutil.Frame, height int) {
	var frame, frameRGB *C.AVFrame
	frame = (*C.AVFrame)(unsafe.Pointer(pFrame.CAVFrame))
	frameRGB = (*C.AVFrame)(unsafe.Pointer(pFrameRGB.CAVFrame))
	C.sws_scale(swsContext.CSwsContext, (**C.uint8_t)(unsafe.Pointer(&frame.data[0])), (*C.int)(unsafe.Pointer(&frame.linesize[0])),
		0, (C.int)(height), (**C.uint8_t)(unsafe.Pointer(&frameRGB.data[0])), (*C.int)(unsafe.Pointer(&frameRGB.linesize[0])))
}
