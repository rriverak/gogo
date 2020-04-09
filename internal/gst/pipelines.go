package gst

import "fmt"

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
		builder.AddSource(vChan, codecName, vChan).AddNewLine()
	}

	// Get GStreamer Pipeline
	gstreamerPipe := builder.Get()
	Logger.Debugf("Generated GStreamer Pipeline => \n %v \n", gstreamerPipe)

	// Create the GPipeline
	return CreatePipeline(codecName, gstreamerPipe, builder.GetClockRate())
}
