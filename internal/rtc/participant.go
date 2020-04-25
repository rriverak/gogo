package rtc

import (
	"io"
	"math/rand"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/utils"
)

//NewParticipant creates a new Participant
func NewParticipant(name string, peerConnectionConfig webrtc.Configuration, media *webrtc.MediaEngine, customPayloadType uint8, codec string) (*Participant, error) {

	api := webrtc.NewAPI(webrtc.WithMediaEngine(*media))

	// Create a PeerConnection
	pc, err := api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		return nil, err
	}

	part := Participant{
		ID:           utils.RandSeq(5),
		Name:         name,
		Peer:         pc,
		API:          api,
		MediaEngine:  media,
		PayloadType:  customPayloadType,
		Codec:        codec,
		DataChannels: map[string]*webrtc.DataChannel{},
	}

	return &part, nil
}

//Participant can connect to a Session
type Participant struct {
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

//VideoOutput is the Video Pipeline Output Track
func (p *Participant) VideoOutput() *webrtc.Track {
	if p.outVideoTrack == nil {
		// Create a new Mixed Video Track if not exists
		mixedVideoTrack, newTrackErr := p.Peer.NewTrack(p.PayloadType, rand.Uint32(), "video", "video-pipe")
		if newTrackErr != nil {
			Logger.Errorf("Error: %v PayloadType: %v", newTrackErr, p.PayloadType)
		}
		Logger.Infof("Participant => %v create output VideoTrack: Code: %v Payload: %v", p.Name, mixedVideoTrack.Codec().Name, mixedVideoTrack.Codec().PayloadType)
		p.outVideoTrack = mixedVideoTrack
	}
	return p.outVideoTrack
}

//AudioOutput is the Audio Pipeline Output Track
func (p *Participant) AudioOutput() *webrtc.Track {
	if p.outAudioTrack == nil {
		// Create a new Mixed Video Track if not exists
		mixedAudioTrack, newTrackErr := p.Peer.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "audio-pipe")
		if newTrackErr != nil {
			Logger.Error(newTrackErr)
		}
		Logger.Infof("Participant => %v create output AudioTrack: Code: %v Payload: %v", p.Name, mixedAudioTrack.Codec().Name, mixedAudioTrack.Codec().PayloadType)
		p.outAudioTrack = mixedAudioTrack
	}
	return p.outAudioTrack
}

//Anwser generates the Anwser for the SDP Handshake
func (p *Participant) Anwser(offer webrtc.SessionDescription) webrtc.SessionDescription {
	// Set the remote SessionDescription for Participant
	err := p.Peer.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer for Participant
	answer, err := p.Peer.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners for Participant
	err = p.Peer.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}
	return answer
}

//OnParticipantSessionMessage attach Session DataChannels
func (p *Participant) OnParticipantSessionMessage(session *Session) func(m webrtc.DataChannelMessage) {
	return func(message webrtc.DataChannelMessage) {
		msg := string(message.Data)
		Logger.Infof("Participant => %v sends to Session => '%v'", p.Name, msg)
		switch msg {
		case "open":
			break
		case "close":
			session.DisconnectParticipant(p) // Remove from Session
			break
		}
	}
}

//OnParticipantConnectionStateChangedHandler handles Participant Timeout
func (p *Participant) OnParticipantConnectionStateChangedHandler(session *Session) func(f webrtc.PeerConnectionState) {
	return func(f webrtc.PeerConnectionState) {
		if f == webrtc.PeerConnectionStateDisconnected || f == webrtc.PeerConnectionStateFailed {
			Logger.Infof("Participant => %v has a Timeout!", p.Name)
			session.RemoveParticipant(p.ID)
		}
	}
}

//OnRemoteTrackHandler dasdas
func (p *Participant) OnRemoteTrackHandler(session *Session) func(*webrtc.Track, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		Logger.Infof("Participant => %v send a Track with Codec: %v Payloadtyp: %v", p.Name, remoteTrack.Codec().Name, remoteTrack.PayloadType())
		if remoteTrack.PayloadType() == p.VideoOutput().PayloadType() {
			// Video Track
			// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
			// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
			go func() {
				ticker := time.NewTicker(rtcpPLIInterval)
				for range ticker.C {
					if rtcpSendErr := p.Peer.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
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
					// Push RTP Samples to GStreamer Pipeline with specific appsrc (participant_id)
					session.VideoPipeline.WriteSampleToInputSource(rtpBuf[:i], p.ID)
				}
			}
		} else if remoteTrack.PayloadType() == p.AudioOutput().PayloadType() {
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
					// Push RTP Samples to GStreamer Pipeline with specific appsrc (participant_id)
					session.AudioPipeline.WriteSampleToInputSource(rtpBuf[:i], p.ID)
				}
			}
		} else {
			Logger.Error("OnTrack Codec not match...!")
			Logger.Errorf("	RemoteTrack=> Codec %v::%v", remoteTrack.PayloadType(), remoteTrack.Codec().Name)
			Logger.Errorf("	VideoTrack => Codec %v::%v", p.VideoOutput().PayloadType(), p.VideoOutput().Codec().Name)
			Logger.Errorf("	AudioTrack => Codec %v::%v", p.AudioOutput().PayloadType(), p.AudioOutput().Codec().Name)
		}
	}
}
