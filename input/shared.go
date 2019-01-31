package input

import (
	"github.com/baohavan/go-libav/avcodec"
	"github.com/baohavan/go-libav/avformat"
)

type StreamContext struct {
	InputFileName    string
	InputFileNameNew string

	inFmtCtx   *avformat.Context
	InCodecCtx *avcodec.Context

	Index      uint //Save index of stream video, using for ignore packet from audio stream, which can make crash app( Not open output for audio)
	PacketChan chan *avcodec.Packet
	openChan   bool
}
