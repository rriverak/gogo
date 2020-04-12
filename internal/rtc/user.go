package rtc

import (
	"io"
	"math/rand"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/signal"
	"github.com/rriverak/gogo/internal/utils"
)

//NewUser creates a new User
func NewUser(name string, peerConnectionConfig webrtc.Configuration, offer webrtc.SessionDescription) (*User, error) {
	media := webrtc.MediaEngine{}
	media.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	customPayloadType, codec := signal.RegisterCodecFromSDPOffer(&media, offer.SDP)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(media))

	for _, cdec := range media.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		Logger.Infof("User => %v offer => Video Codec: %v PayloadType: %v Clock: %v", name, cdec.Name, cdec.PayloadType, cdec.ClockRate)
	}
	for _, cdec := range media.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
		Logger.Infof("User => %v offer => Audio Codec: %v PayloadType: %v Clock: %v", name, cdec.Name, cdec.PayloadType, cdec.ClockRate)
	}

	// Create a PeerConnection
	pc, err := api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		return nil, err
	}

	newUser := User{
		ID:           utils.RandSeq(5),
		Name:         name,
		Peer:         pc,
		API:          api,
		MediaEngine:  &media,
		PayloadType:  customPayloadType,
		Codec:        codec,
		DataChannels: map[string]*webrtc.DataChannel{},
	}

	// Allow the Peer to send a Video Stream
	if _, err = newUser.Peer.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	// Allow the Peer to send a Audio Stream
	if _, err = newUser.Peer.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	// Add VideoMixed Track to Peer
	if _, err = newUser.Peer.AddTrack(newUser.VideOutput()); err != nil {
		panic(err)
	}
	// Add AudioMixed Track to Peer
	if _, err = newUser.Peer.AddTrack(newUser.AudioOutput()); err != nil {
		panic(err)
	}
	return &newUser, nil
}

//User can Connect to a Session
type User struct {
	ID            string
	Name          string
	MediaEngine   *webrtc.MediaEngine
	API           *webrtc.API
	outVideoTrack *webrtc.Track
	outAudioTrack *webrtc.Track
	Codec         string
	PayloadType   uint8
	Peer          *webrtc.PeerConnection
	DataChannels  map[string]*webrtc.DataChannel
}

//VideOutput is the Video Pipeline Output Track
func (u *User) VideOutput() *webrtc.Track {
	if u.outVideoTrack == nil {
		// Create a new Mixed Video Track if not exists
		mixedVideoTrack, newTrackErr := u.Peer.NewTrack(u.PayloadType, rand.Uint32(), "video", "video-pipe")
		if newTrackErr != nil {
			panic(newTrackErr)
		}
		Logger.Infof("User => %v create output VideoTrack: Code: %v Payload: %v", u.Name, mixedVideoTrack.Codec().Name, mixedVideoTrack.Codec().PayloadType)
		u.outVideoTrack = mixedVideoTrack
	}
	return u.outVideoTrack
}

//AudioOutput is the Audio Pipeline Output Track
func (u *User) AudioOutput() *webrtc.Track {
	if u.outAudioTrack == nil {
		// Create a new Mixed Video Track if not exists
		mixedAudioTrack, newTrackErr := u.Peer.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "audio-pipe")
		if newTrackErr != nil {
			panic(newTrackErr)
		}
		Logger.Infof("User => %v create output AudioTrack: Code: %v Payload: %v", u.Name, mixedAudioTrack.Codec().Name, mixedAudioTrack.Codec().PayloadType)
		u.outAudioTrack = mixedAudioTrack
	}
	return u.outAudioTrack
}

//Anwser generates the Anwser for the SDP Handshake
func (u *User) Anwser(offer webrtc.SessionDescription) webrtc.SessionDescription {
	// Set the remote SessionDescription for User
	err := u.Peer.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer for User
	answer, err := u.Peer.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners for User
	err = u.Peer.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}
	return answer
}

//OnUserSessionMessage attach all known DataChannels
func (u *User) OnUserSessionMessage(session *Session) func(m webrtc.DataChannelMessage) {
	return func(message webrtc.DataChannelMessage) {
		msg := string(message.Data)
		Logger.Infof("User => %v sends to Session => '%v'", u.Name, msg)
		switch msg {
		case "open":
			break
		case "close":
			session.DisconnectUser(u) // Remove from Session
			break
		}
	}
}

//OnUserConnectionStateChangedHandler handles user Timeout
func (u *User) OnUserConnectionStateChangedHandler(session *Session) func(f webrtc.PeerConnectionState) {
	return func(f webrtc.PeerConnectionState) {
		if f == webrtc.PeerConnectionStateDisconnected || f == webrtc.PeerConnectionStateFailed {
			Logger.Infof("User => %v has a Timeout!", u.Name)
			session.RemoveUser(u.ID)
		}
	}
}

//OnRemoteTrackHandler dasdas
func (u *User) OnRemoteTrackHandler(session *Session) func(*webrtc.Track, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		Logger.Infof("User => %v send a Track with Codec: %v Payloadtyp: %v", u.Name, remoteTrack.Codec().Name, remoteTrack.PayloadType())
		if remoteTrack.PayloadType() == u.VideOutput().PayloadType() {
			// Video Track
			// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
			// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
			go func() {
				ticker := time.NewTicker(rtcpPLIInterval)
				for range ticker.C {
					if rtcpSendErr := u.Peer.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
						if rtcpSendErr == io.ErrClosedPipe {
							ticker.Stop()
						} else {
							Logger.Errorf("rtcp PLI Error: %v", rtcpSendErr)
						}
					}
				}
			}()

			// Create a Buffer Loop
			rtpBuf := make([]byte, 1400)
			for {
				// Read remote Buffer
				i, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					if readErr == io.EOF {
						break
					}
					Logger.Errorf("Read on RemoteTrack Error: %v", readErr)
				} else {
					// Push RTP Samples to GStreamer Pipeline with specific appsrc (user_id)
					session.VideoPipeline.WriteSampleToInputSource(rtpBuf[:i], u.ID)
				}
			}
		} else if remoteTrack.PayloadType() == u.AudioOutput().PayloadType() {
			// Audio Track
			// Create a Buffer Loop
			rtpBuf := make([]byte, 1400)
			for {
				// Read remote Buffer
				i, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					if readErr == io.EOF {
						break
					}
					Logger.Errorf("Read on RemoteTrack Error: %v", readErr)
				} else {
					// Push RTP Samples to GStreamer Pipeline with specific appsrc (user_id)
					session.AudioPipeline.WriteSampleToInputSource(rtpBuf[:i], u.ID)
				}
			}
		} else {
			Logger.Error("OnTrack Codec not match...!")
			Logger.Errorf("	RemoteTrack=> Codec %v::%v", remoteTrack.PayloadType(), remoteTrack.Codec().Name)
			Logger.Errorf("	VideoTrack => Codec %v::%v", u.VideOutput().PayloadType(), u.VideOutput().Codec().Name)
			Logger.Errorf("	AudioTrack => Codec %v::%v", u.AudioOutput().PayloadType(), u.AudioOutput().Codec().Name)
		}
	}
}
