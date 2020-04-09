#ifndef GST_H
#define GST_H

#include <glib.h>
#include <gst/gst.h>
#include <stdint.h>
#include <stdlib.h>

extern void goHandlePipelineOutputBuffer(void *buffer, int bufferLen, int samples, char *pipelineId);

GstElement *gstreamer_create_pipeline(char *pipeline);
void gstreamer_start_pipeline(GstElement *pipeline, char *pipelineId);
void gstreamer_stop_pipeline(GstElement *pipeline);
void gstreamer_push_buffer(GstElement *pipeline, void *buffer, int len, char *appSource);
void gstreamer_start_mainloop(void);

#endif