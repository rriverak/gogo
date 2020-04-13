package gst

import (
	"fmt"

	"github.com/rriverak/gogo/internal/config"
)

//Channel in the Pipeline
type Channel struct {
	SourceID string
	Name     string
}

//NewChannel create a new Channel in the Pipeline
func NewChannel(id string, name string) *Channel {
	return &Channel{SourceID: id, Name: name}
}

// CreateVideoMixerPipeline Creating a Pipeline for Composite Video Mixing
func CreateVideoMixerPipeline(codecName string, channels []*Channel, cfg *config.Config) *Pipeline {
	// Definition
	outputSinkName := "appsink"
	mixerSinkSettings := []string{}

	//Build Video Grid Matrix 3 Colums per Row with 350px*350px
	col := 0
	row := 0
	fullSize := cfg.Media.Video.GetFullSize()
	for i := range channels {
		if i > 0 {
			col += fullSize
			if i%3 == 0 {
				row += fullSize
				col = 0
			}
		}
		mixerSinkSettings = append(mixerSinkSettings, fmt.Sprintf(" sink_%v::xpos=%v sink_%v::ypos=%v ", i, col, i, row))
	}

	// Builder
	builder := newGstBuilder(cfg.Media)
	// VideoMixer
	builder.AddVideoMixer(codecName, mixerSinkSettings, outputSinkName)

	// VideoSources
	for _, vChan := range channels {
		builder.AddSource(vChan.SourceID, codecName, vChan.Name, true).AddNewLine()
	}

	// Get GStreamer Pipeline
	gstreamerPipe := builder.Get()
	Logger.Debugf("[GStreamer] Generated VideoMixer Pipeline => \n %v \n", gstreamerPipe)

	// Create the GPipeline
	return CreatePipeline(codecName, gstreamerPipe, builder.GetClockRate())
}

//CreateAudioMixerPipeline Creating a Pipeline for Composite Audio Mixing (n-1)
func CreateAudioMixerPipeline(codecName string, channels []*Channel, cfg *config.Config) *Pipeline {
	// Definition
	outputSinkName := "appsink"
	mixerSinkSettings := []string{}

	// Builder
	builder := newGstBuilder(cfg.Media)
	// VideoMixer
	builder.AddAudioMixer(codecName, mixerSinkSettings, outputSinkName)

	// AudioSources
	for _, vChan := range channels {
		builder.AddSource(vChan.SourceID, codecName, vChan.Name, false).AddNewLine()
	}

	// Get GStreamer Pipeline
	gstreamerPipe := builder.Get()
	Logger.Debugf("[GStreamer] Generated AudioMixer Pipeline => \n %v \n", gstreamerPipe)

	return CreatePipeline(codecName, gstreamerPipe, builder.GetClockRate())
}
