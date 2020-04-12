package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/rriverak/gogo/internal/utils"
	"github.com/sirupsen/logrus"
)

// Logger is the Gst Logger Instance
var Logger *logrus.Logger

var pipelines = make(map[string]*Pipeline)
var pipelinesLock sync.Mutex

// Pipeline is a wrapper for a GStreamer  Pipeline
type Pipeline struct {
	Pipeline     *C.GstElement
	outputTracks []*webrtc.Track
	id           string
	codecName    string
	clockRate    float32
}

//AddOutputTrack to the Track Stream
func (p *Pipeline) AddOutputTrack(newTrack *webrtc.Track) {
	p.outputTracks = append(p.outputTracks, newTrack)
}

//GetOutputTracks get the Tracks
func (p *Pipeline) GetOutputTracks() []*webrtc.Track {
	return p.outputTracks
}

//SettingOutputTracks set the Tracks
func (p *Pipeline) SettingOutputTracks(tracks []*webrtc.Track) {
	p.outputTracks = tracks
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	C.gstreamer_start_pipeline(p.Pipeline, C.CString(p.id))
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {
	// Lock Pipelines
	C.gstreamer_stop_pipeline(p.Pipeline)
}

//WriteSampleToOutputTrack to the OutputTrack
func (p *Pipeline) WriteSampleToOutputTrack(buffer []byte, samples uint32) error {
	for _, track := range p.outputTracks {
		err := track.WriteSample(media.Sample{Data: buffer, Samples: samples})
		if err != nil {
			Logger.Errorf("WriteSampleToOutputTrack => %v", track.ID(), err)
		}
	}
	return nil
}

// WriteSampleToInputSource writes a Buffer to a appsrc of the GStreamer Pipeline
func (p *Pipeline) WriteSampleToInputSource(buffer []byte, sourceID string) {
	// Generate AppSource
	appSource := getAppSrcString(sourceID)
	// App Source as CString
	appSourceStrUnsafe := C.CString(appSource)
	defer C.free(unsafe.Pointer(appSourceStrUnsafe))
	// Buffer as CBytes
	b := C.CBytes(buffer)
	defer C.free(b)

	// Push Buffer to Pipeline
	// #TODO: Not good that p could be empty in case of GC Collect! Leak in Stop func Possible
	if p != nil && p.Pipeline != nil {
		C.gstreamer_push_buffer(p.Pipeline, b, C.int(len(buffer)), appSourceStrUnsafe)
	}
}

//StartMainLoop for GStreamer
func StartMainLoop() {
	C.gstreamer_start_mainloop()
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(codecName string, pipelineStr string, clockRate float32) *Pipeline {
	// Generate C String from Input
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	// Lock Pipelines
	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	// Create new Pipeline
	pipeline := &Pipeline{
		Pipeline:  C.gstreamer_create_pipeline(pipelineStrUnsafe),
		id:        utils.RandSeq(5),
		codecName: codecName,
		clockRate: clockRate,
	}
	pipeline.outputTracks = []*webrtc.Track{}
	// Add new Pipeline
	pipelines[pipeline.id] = pipeline
	return pipeline
}

//export goHandlePipelineOutputBuffer
func goHandlePipelineOutputBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int, pipelineID *C.char) {
	// Lock Pipelines
	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()
	// Get Running Pipeline
	pipelineIDstr := C.GoString(pipelineID)
	pipeline, ok := pipelines[pipelineIDstr]
	if ok {
		// Create Samples
		samples := uint32(pipeline.clockRate * (float32(duration) / 1000000000))
		// Write Samples to Pipeline Output Track
		if err := pipeline.WriteSampleToOutputTrack(C.GoBytes(buffer, bufferLen), samples); err != nil && err != io.ErrClosedPipe {
			Logger.Error(err)
		}
	} else {
		Logger.Errorf("Discarding buffer! No pipeline with ID => '%s' found... \n", pipelineIDstr)
	}
	// Free old Buffer
	C.free(buffer)
}

func getAppSrcString(str string) string {
	return fmt.Sprintf("src-%v", str)
}
