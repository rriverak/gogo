#include "gst.h"

#include <gst/app/gstappsrc.h>

GMainLoop *gstreamer_pipeline_mainloop = NULL;

typedef struct SampleHandlerUserData {
  char *pipelineId;
} SampleHandlerUserData;

// Starting the MainLoop for GStreamer
void gstreamer_start_mainloop(void) {
  gstreamer_pipeline_mainloop = g_main_loop_new(NULL, FALSE);
  g_main_loop_run(gstreamer_pipeline_mainloop);
}

// Create Pipeline
GstElement *gstreamer_create_pipeline(char *pipeline) {
  gst_init(NULL, NULL);
  GError *error = NULL;
  return gst_parse_launch(pipeline, &error);
}

// Pipeline Bus
static gboolean gstreamer_pipeline_bus_call(GstBus *bus, GstMessage *msg, gpointer data) {
  switch (GST_MESSAGE_TYPE(msg)) {

  case GST_MESSAGE_EOS:
    g_print("End of stream\n");
    exit(1);
    break;

  case GST_MESSAGE_ERROR: {
    gchar *debug;
    GError *error;

    gst_message_parse_error(msg, &error, &debug);
    g_free(debug);

    g_printerr("Error: %s\n", error->message);
    g_error_free(error);
    exit(1);
  }
  default:
    break;
  }

  return TRUE;
}

// Pipeline Bus
static gboolean gstreamer_pipeline_send_bus_call(GstBus *bus, GstMessage *msg, gpointer data) {
  switch (GST_MESSAGE_TYPE(msg)) {

  case GST_MESSAGE_EOS:
    g_print("End of stream\n");
    exit(1);
    break;

  case GST_MESSAGE_ERROR: {
    gchar *debug;
    GError *error;

    gst_message_parse_error(msg, &error, &debug);
    g_free(debug);

    g_printerr("Error: %s\n", error->message);
    g_error_free(error);
    exit(1);
  }
  default:
    break;
  }

  return TRUE;
}

// Handles the Output Samples of the Pipeline
GstFlowReturn gstreamer_pipeline_output_sample_handler(GstElement *object, gpointer user_data) {
  GstSample *sample = NULL;
  GstBuffer *buffer = NULL;
  gpointer copy = NULL;
  gsize copy_size = 0;
  SampleHandlerUserData *s = (SampleHandlerUserData *)user_data;

  g_signal_emit_by_name (object, "pull-sample", &sample);
  if (sample) {
    buffer = gst_sample_get_buffer(sample);
    if (buffer) {
      gst_buffer_extract_dup(buffer, 0, gst_buffer_get_size(buffer), &copy, &copy_size);
      goHandlePipelineOutputBuffer(copy, copy_size, GST_BUFFER_DURATION(buffer), s->pipelineId);
    }
    gst_sample_unref (sample);
  }

  return GST_FLOW_OK;
}

//Start Pipeline
void gstreamer_start_pipeline(GstElement *pipeline, char *pipelineId) {
  SampleHandlerUserData *s = calloc(1, sizeof(SampleHandlerUserData));
  s->pipelineId = pipelineId;

  GstBus *bus = gst_pipeline_get_bus(GST_PIPELINE(pipeline));
  gst_bus_add_watch(bus, gstreamer_pipeline_bus_call, NULL);
  gst_bus_add_watch(bus, gstreamer_pipeline_send_bus_call, NULL);
  gst_object_unref(bus);

  GstElement *appsink = gst_bin_get_by_name(GST_BIN(pipeline), "appsink");
  g_object_set(appsink, "emit-signals", TRUE, NULL);
  g_signal_connect(appsink, "new-sample", G_CALLBACK(gstreamer_pipeline_output_sample_handler), s);
  gst_object_unref(appsink);

  gst_element_set_state(pipeline, GST_STATE_PLAYING);
}

//Stop Pipeline
void gstreamer_stop_pipeline(GstElement *pipeline) { 
  gst_element_set_state(pipeline, GST_STATE_NULL); 
}

//Push Buffer to Pipeline AppSource
void gstreamer_push_buffer(GstElement *pipeline, void *buffer, int len, char *appSource) {
  GstElement *src = gst_bin_get_by_name(GST_BIN(pipeline), appSource);
  if (src != NULL) {
    gpointer p = g_memdup(buffer, len);
    GstBuffer *buffer = gst_buffer_new_wrapped(p, len);
    gst_app_src_push_buffer(GST_APP_SRC(src), buffer);
    gst_object_unref(src);
  }
}

