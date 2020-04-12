// Package signal contains helpers to exchange the SDP session
// description between examples.
package signal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v2"
)

// Allows compressing offer/answer to bypass terminal input limits.
const compress = false

// MustReadStdin blocks until input is received from stdin
func MustReadStdin() string {
	r := bufio.NewReader(os.Stdin)

	var in string
	for {
		var err error
		in, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				panic(err)
			}
		}
		in = strings.TrimSpace(in)
		if len(in) > 0 {
			break
		}
	}

	fmt.Println("")

	return in
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	if compress {
		b = zip(b)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func Decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if compress {
		b = unzip(b)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}

func zip(in []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(in)
	if err != nil {
		panic(err)
	}
	err = gz.Flush()
	if err != nil {
		panic(err)
	}
	err = gz.Close()
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func unzip(in []byte) []byte {
	var b bytes.Buffer
	_, err := b.Write(in)
	if err != nil {
		panic(err)
	}
	r, err := gzip.NewReader(&b)
	if err != nil {
		panic(err)
	}
	res, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return res
}

var (
	// VP8 Enabled
	VP8 = true
	// VP9 Enabled
	VP9 = false
	// H264 Enabled
	H264 = true
)

//RegisterCodecFromSDPOffer to ensure the right PayloadType
func RegisterCodecFromSDPOffer(mediaEngine *webrtc.MediaEngine, sdpStr string) (uint8, string) {
	var customPayloadType uint8 = 0
	var customCodec string = ""
	parsed := sdp.SessionDescription{}
	if err := parsed.Unmarshal([]byte(sdpStr)); err != nil {
		panic(err)
	}

	if customPayloadType == 0 && VP9 {
		codecStr := sdp.Codec{
			Name: "VP9",
		}
		customVP9PayloadType, err := parsed.GetPayloadTypeForCodec(codecStr)
		if err != nil {
			customPayloadType = 0
		} else {
			customPayloadType = customVP9PayloadType
			sdpCodec, _ := parsed.GetCodecForPayloadType(customVP9PayloadType)
			customCodec = sdpCodec.Name
			codec := webrtc.NewRTPVP9Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)
		}
	}
	if customPayloadType == 0 && VP8 {
		codecStr := sdp.Codec{
			Name: "VP8",
		}
		customVP8PayloadType, err := parsed.GetPayloadTypeForCodec(codecStr)
		if err != nil {
			customPayloadType = 0
		} else {
			customPayloadType = customVP8PayloadType
			sdpCodec, _ := parsed.GetCodecForPayloadType(customVP8PayloadType)
			customCodec = sdpCodec.Name
			codec := webrtc.NewRTPVP8Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)
		}
	}
	if customPayloadType == 0 && H264 {
		codecStr := sdp.Codec{
			Name: "H264",
		}
		customH264PayloadType, err := parsed.GetPayloadTypeForCodec(codecStr)
		if err != nil {
			customPayloadType = 0
		} else {
			customPayloadType = customH264PayloadType
			sdpCodec, _ := parsed.GetCodecForPayloadType(customH264PayloadType)
			customCodec = sdpCodec.Name
			codec := webrtc.NewRTPH264Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)
		}
	}
	return customPayloadType, customCodec
}
