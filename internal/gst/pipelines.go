package gst

import (
	"fmt"
)

// CreateVideoMixerPipeline Creating a Pipeline for Composite Video Mixing
func CreateVideoMixerPipeline(codecName string, channels []string) *Pipeline {
	// Definition
	outputSinkName := "appsink"
	mixerSinkSettings := []string{}

	//Build Video Grid Matrix 3 Colums per Row with 350px*350px
	col := 0
	row := 0
	for i := range channels {
		if i > 0 {
			col += 350
			if i%3 == 0 {
				row += 350
				col = 0
			}
		}
		mixerSinkSettings = append(mixerSinkSettings, fmt.Sprintf(" sink_%v::xpos=%v sink_%v::ypos=%v ", i, col, i, row))
	}

	// Builder
	builder := newGstBuilder()
	// VideoMixer
	builder.AddVideoMixer(codecName, mixerSinkSettings, outputSinkName)

	// VideoSources
	for _, vChan := range channels {
		builder.AddSource(vChan, codecName, vChan, true).AddNewLine()
	}

	// Get GStreamer Pipeline
	gstreamerPipe := builder.Get()
	Logger.Infof("[GStreamer] Generated VideoMixer Pipeline => \n %v \n", gstreamerPipe)

	// Create the GPipeline
	return CreatePipeline(codecName, gstreamerPipe, builder.GetClockRate())
}

//CreateAudioMixerPipeline Creating a Pipeline for Composite Audio Mixing (n-1)
func CreateAudioMixerPipeline(codecName string, channels []string) *Pipeline {
	// Definition
	outputSinkName := "appsink"
	mixerSinkSettings := []string{}

	// Builder
	builder := newGstBuilder()
	// VideoMixer
	builder.AddAudioMixer(codecName, mixerSinkSettings, outputSinkName)

	// AudioSources
	for _, vChan := range channels {
		builder.AddSource(vChan, codecName, "", false).AddNewLine()
	}

	// Get GStreamer Pipeline
	gstreamerPipe := builder.Get()
	Logger.Infof("[GStreamer] Generated AudioMixer Pipeline => \n %v \n", gstreamerPipe)

	return CreatePipeline(codecName, gstreamerPipe, builder.GetClockRate())
}
