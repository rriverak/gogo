package rtc

import (
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/config"
)

//GetMediaEngineForSDPOffer for an Offer
func GetMediaEngineForSDPOffer(offer webrtc.SessionDescription, mediaCfg *config.MediaConfig) (uint8, string, *webrtc.MediaEngine) {
	media := webrtc.MediaEngine{}

	// Register Video Codec with dynamic
	customPayloadType, codec := registerCodecFromSDPOffer(&media, offer.SDP, mediaCfg)

	// Default Audio Codec
	media.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))

	return customPayloadType, codec, &media
}

//RegisterCodecFromSDPOffer to ensure the right PayloadType
func registerCodecFromSDPOffer(mediaEngine *webrtc.MediaEngine, sdpStr string, mediaCfg *config.MediaConfig) (uint8, string) {
	var customPayloadType uint8 = 0
	var customCodec string = ""
	parsed := sdp.SessionDescription{}
	if err := parsed.Unmarshal([]byte(sdpStr)); err != nil {
		panic(err)
	}

	if customPayloadType == 0 && mediaCfg.Video.Codecs.VP9 {
		customPayloadType, sdpCodec := findSDPCodec(&parsed, "VP9")
		if customPayloadType != 0 {
			// Register PayloadType with Codec
			codec := webrtc.NewRTPVP9Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)

		}
	}
	if customPayloadType == 0 && mediaCfg.Video.Codecs.VP8 {
		customPayloadType, sdpCodec := findSDPCodec(&parsed, "VP8")
		if customPayloadType != 0 {
			// Register PayloadType with Codec
			codec := webrtc.NewRTPVP8Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)

		}
	}
	if customPayloadType == 0 && mediaCfg.Video.Codecs.H264 {

		customPayloadType, sdpCodec := findSDPCodec(&parsed, "H264")
		if customPayloadType != 0 {
			// Register PayloadType with Codec
			codec := webrtc.NewRTPH264Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)

		}
	}
	return customPayloadType, customCodec
}

func findSDPCodec(desc *sdp.SessionDescription, codecName string) (uint8, *sdp.Codec) {
	// Get PayloadType from SDP for Codec
	pt, err := desc.GetPayloadTypeForCodec(sdp.Codec{
		Name: codecName,
	})
	if err != nil {
		return 0, nil
	}
	// Get Codec from SDP for PayloadType
	sdpCodec, err := desc.GetCodecForPayloadType(pt)
	if err != nil {
		return 0, nil
	}
	return pt, &sdpCodec
}
