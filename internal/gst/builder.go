package gst

import (
	"fmt"
	"strings"

	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/config"
)

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

type gstBuilder struct {
	Pipe      string
	gstStr    string
	clockRate float32
	config    *config.MediaConfig
}

func newGstBuilder(config *config.MediaConfig) *gstBuilder {
	return &gstBuilder{
		Pipe:   "!",
		config: config,
	}
}

// High Level Builder

func (g *gstBuilder) AddSource(sourceID string, codecName string, overlayText string, boxed bool) *gstBuilder {
	sourceName := getAppSrcString(sourceID)
	g.AddStr(fmt.Sprintf("appsrc format=time is-live=true do-timestamp=true name=%v", sourceName)).AddPipe()
	g.AddStr("application/x-rtp")
	g.AddStr(gstGetDecoder(codecName)).AddPipe()
	//VideoBox
	if boxed {
		fullSize := g.config.Video.GetFullSize()
		g.AddStr(fmt.Sprintf("videobox autocrop=true ! video/x-raw, width=%v, height=%v", fullSize, fullSize)).AddPipe()
	}
	//Overlay
	if len(overlayText) > 0 {
		g.AddStr(gstGetTextOverlay(overlayText)).AddPipe()
	}
	//g.AddStr("queue2").AddPipe()
	g.AddStr("mix.").AddNewLine()
	return g
}

func (g *gstBuilder) AddVideoMixer(codecName string, sinkSettings []string, outputSinkName string) *gstBuilder {
	g.AddStr("compositor name=mix background=black ")
	g.AddStr(strings.Join(sinkSettings, " "))
	encoder, clockRate := gstGetEncoder(codecName)
	g.clockRate = clockRate
	return g.AddStr(encoder).AddPipe().AddStr(fmt.Sprintf("appsink drop=true max-buffers=1 name=%v ", outputSinkName)).AddPipe().AddNewLine()
}
func (g *gstBuilder) AddAudioMixer(codecName string, sinkSettings []string, outputSinkName string) *gstBuilder {
	g.AddStr("audiomixer start-time-selection=first name=mix ")
	g.AddStr(strings.Join(sinkSettings, " "))
	encoder, clockRate := gstGetEncoder(codecName)
	g.clockRate = clockRate
	return g.AddStr(encoder).AddPipe().AddStr(fmt.Sprintf("appsink drop=true max-buffers=1 name=%v ", outputSinkName)).AddPipe().AddNewLine()
}

// LowLevel Funcs
func (g *gstBuilder) AddStr(str string) *gstBuilder {
	g.gstStr += fmt.Sprintf(" %v ", str)
	return g
}

func (g *gstBuilder) AddPipe() *gstBuilder {
	return g.AddStr(fmt.Sprintf(" %v ", g.Pipe))
}

func (g *gstBuilder) AddNewLine() *gstBuilder {
	return g.AddStr(" \n")
}

func (g *gstBuilder) Get() string {
	return g.gstStr
}

func (g *gstBuilder) GetClockRate() float32 {
	return g.clockRate
}

func gstGetTextOverlay(text string) string {
	return fmt.Sprintf("textoverlay text=\"%v\" valignment=bottom halignment=center font-desc=\"Sans, 42\" ", text)
}

func gstGetDecoder(codecName string) string {
	var chansrc string = ""
	switch codecName {
	case webrtc.VP8:
		chansrc += ", encoding-name=VP8-DRAFT-IETF-01 ! rtpvp8depay ! decodebin "
	case webrtc.Opus:
		chansrc += ", payload=96, encoding-name=OPUS ! rtpopusdepay ! decodebin "
	case webrtc.VP9:
		chansrc += " rtpvp9depay ! decodebin  "
	case webrtc.H264:
		chansrc += " ! rtph264depay ! decodebin "
	case webrtc.G722:
		chansrc += " clock-rate=8000 ! rtpg722depay ! decodebin "
	default:
		panic("Unhandled codec " + codecName)
	}
	return chansrc
}

func gstGetEncoder(codecName string) (string, float32) {
	switch codecName {
	case webrtc.VP8:
		return " ! video/x-raw,format=I420 ! vp8enc error-resilient=partitions keyframe-max-dist=30 buffer-size=0 auto-alt-ref=true cpu-used=5 deadline=1 ", videoClockRate
	case webrtc.VP9:
		return " ! vp9enc ", videoClockRate
	case webrtc.H264:
		return " ! video/x-raw,format=I420 ! x264enc bframes=0 speed-preset=veryfast key-int-max=60 ! video/x-h264,stream-format=byte-stream ", videoClockRate
	case webrtc.Opus:
		return " ! opusenc ", audioClockRate
	case webrtc.G722:
		return " ! avenc_g722 ", audioClockRate
	case webrtc.PCMU:
		return " ! audio/x-raw, rate=8000 ! mulawenc ", pcmClockRate
	case webrtc.PCMA:
		return " ! audio/x-raw, rate=8000 ! alawenc ", pcmClockRate
	default:
		fmt.Println("Unhandled codec " + codecName)
		return "", 0
	}
}
