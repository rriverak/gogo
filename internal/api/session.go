package api

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/pion/rtcp"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/rtc"
	"github.com/rriverak/gogo/internal/signal"
)

var peerConnectionConfig webrtc.Configuration = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

const (
	rtcpPLIInterval = time.Second * 1
	VP8             = true
	VP9             = false
	H246            = true
)

//SessionHandler handles API Requests for Sessions
type SessionHandler struct {
	SessionRegister *cache.Cache
	MediaEngine     *webrtc.MediaEngine
}

//RegisterSessionRoutes apply all Routes to the Router
func (s *SessionHandler) RegisterSessionRoutes(r *mux.Router) {
	s.SessionRegister = cache.New(cache.NoExpiration, cache.NoExpiration)
	sub := r.PathPrefix("/api/sessions").Subrouter()
	sub.HandleFunc("/", s.ListSessions).Methods("GET")
	sub.HandleFunc("/{id}/{user}", s.JoiningSessions).Methods("POST")
}

//ListSessions Handles a HTTP Get to List all Sessions with Users
func (s *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, s.SessionRegister.Items())
}

//GetSessionKey for Cache Repository
func (s *SessionHandler) GetSessionKey(sessID string) string {
	return fmt.Sprintf("session-%v", sessID)
}

func (s *SessionHandler) setSDPOfferCodec(session *rtc.Session, mediaEngine *webrtc.MediaEngine, sdpStr string) uint8 {
	var customPayloadType uint8 = 0
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
			Logger.Info(err)
			customPayloadType = 0
		} else {
			customPayloadType = customVP9PayloadType
			Logger.Info("Found VP9 Codec in SDP")
			sdpCodec, _ := parsed.GetCodecForPayloadType(customVP9PayloadType)
			session.Codec = sdpCodec.Name
			codec := webrtc.NewRTPVP9Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)
			return customPayloadType
		}
	}
	if customPayloadType == 0 && VP8 {
		codecStr := sdp.Codec{
			Name: "VP8",
		}
		customVP8PayloadType, err := parsed.GetPayloadTypeForCodec(codecStr)
		if err != nil {
			Logger.Info(err)
			customPayloadType = 0
		} else {
			customPayloadType = customVP8PayloadType
			Logger.Info("Found VP8 Codec in SDP")
			sdpCodec, _ := parsed.GetCodecForPayloadType(customVP8PayloadType)
			session.Codec = sdpCodec.Name
			codec := webrtc.NewRTPVP8Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)
			return customPayloadType
		}
	}
	if customPayloadType == 0 && H246 {
		codecStr := sdp.Codec{
			Name: "H264",
		}
		customH246PayloadType, err := parsed.GetPayloadTypeForCodec(codecStr)
		if err != nil {
			Logger.Info(err)
			customPayloadType = 0
		} else {
			customPayloadType = customH246PayloadType
			Logger.Info("Found H264 Codec in SDP")
			sdpCodec, _ := parsed.GetCodecForPayloadType(customH246PayloadType)
			session.Codec = sdpCodec.Name
			codec := webrtc.NewRTPH264Codec(customPayloadType, 90000)
			codec.SDPFmtpLine = sdpCodec.Fmtp
			mediaEngine.RegisterCodec(codec)
			return customPayloadType
		}
	}
	return customPayloadType
}

//JoiningSessions Handles the Joining Offer
func (s *SessionHandler) JoiningSessions(w http.ResponseWriter, r *http.Request) {
	// HTTP VARs
	sessID := mux.Vars(r)["id"]
	userName := mux.Vars(r)["user"]

	// Get or Create a Session
	var session *rtc.Session
	sessKey := s.GetSessionKey(sessID)
	sess, sessionFound := s.SessionRegister.Get(sessKey)
	if sessionFound {
		Logger.Infof("Join Session with Key => '%v' \n", sessKey)
		session = sess.(*rtc.Session)
	} else {
		Logger.Infof("Create new Session with Key => '%v' \n", sessKey)
		session = &rtc.Session{ID: sessID}
	}

	m := webrtc.MediaEngine{}
	m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	// m.RegisterDefaultCodecs()
	// m.PopulateFromSDP(offer)

	// Get the offer
	offer := webrtc.SessionDescription{}
	body, _ := ioutil.ReadAll(r.Body)
	signal.Decode(string(body), &offer)

	customPayloadType := s.setSDPOfferCodec(session, &m, offer.SDP)
	/*
		switch codec {
		case webrtc.VP9:
			m.RegisterCodec(webrtc.NewRTPVP9Codec(customPayloadType, 90000))
			session.Codec = webrtc.VP9
			break
		case webrtc.VP8:
			m.RegisterCodec(webrtc.NewRTPVP8Codec(customPayloadType, 90000))
			session.Codec = webrtc.VP8
			break
		case webrtc.H264:
			m.RegisterCodec(webrtc.NewRTPH264Codec(customPayloadType, 90000))
			session.Codec = webrtc.H264
			break
		default:
			//m.RegisterDefaultCodecs()
			break
		}
	*/
	Logger.Info("CustomPayloadType =>", customPayloadType)

	// Create a API
	rtcAPI := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	/*
		if session.API == nil {
			session.API = webrtc.NewAPI(webrtc.WithMediaEngine(m))
		}*/

	// Create a Peer
	peerConnection, err := rtcAPI.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		panic(err)
	}

	// Add User to List
	newUser := rtc.NewUser(userName)
	newUser.Peer = peerConnection
	session.AddUser(newUser)
	Logger.Infof("Users in Session => %v", session.Users)

	for _, cdec := range m.GetCodecsByKind(webrtc.RTPCodecTypeVideo) {
		Logger.Infof("User: %v Video Codec: %v PayloadType: %v Clock: %v", userName, cdec.Name, cdec.PayloadType, cdec.ClockRate)
	}
	for _, cdec := range m.GetCodecsByKind(webrtc.RTPCodecTypeAudio) {
		Logger.Infof("User: %v Audio Codec: %v PayloadType: %v Clock: %v", userName, cdec.Name, cdec.PayloadType, cdec.ClockRate)
	}

	// Allow the Peer to send a Video Stream
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	// Allow the Peer to send a Audio Stream
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	// Create a new Mixed Video Track if not exists
	mixedVideoTrack, newTrackErr := peerConnection.NewTrack(customPayloadType, rand.Uint32(), "video", "video-mixed")
	if newTrackErr != nil {
		Logger.Errorf("Err: %v PayloadType: %v ", newTrackErr, customPayloadType)
		panic(newTrackErr)
	}

	// Create a new Mixed Audio Track if not exists
	mixedAudioTrack, newTrackErr := peerConnection.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "audio-mixed")
	if newTrackErr != nil {
		Logger.Errorf("Err: %v PayloadType: %v ", newTrackErr, customPayloadType)
		panic(newTrackErr)
	}

	// Add VideoMixed Track to Peer
	if _, err = peerConnection.AddTrack(mixedVideoTrack); err != nil {
		panic(err)
	}
	// Add AudioMixed Track to Peer
	if _, err = peerConnection.AddTrack(mixedAudioTrack); err != nil {
		panic(err)
	}

	session.Restart()
	session.VideoPipeline.AddOutputTrack(mixedVideoTrack)
	session.AudioPipeline.AddOutputTrack(mixedAudioTrack)

	Logger.Infof("Create Output VideoMix: Code: %v Payload: %v", mixedVideoTrack.Codec().Name, mixedVideoTrack.Codec().PayloadType)
	Logger.Infof("Create Output AudioMix: Code: %v Payload: %v", mixedAudioTrack.Codec().Name, mixedAudioTrack.Codec().PayloadType)

	// On Peer Conncetion Timeout or Disconnected
	peerConnection.OnConnectionStateChange(func(f webrtc.PeerConnectionState) {
		Logger.Infof("User '%v' State Changed => %v", userName, f.String())
		if f == webrtc.PeerConnectionStateDisconnected {
			Logger.Infof("User => %v Timeout or Disconnected", userName)
			session.RemoveUser(userName)
		}
	})

	// DataChannel
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		Logger.Infof("New DataChannel %s %d", d.Label(), d.ID())
		// Register channel opening handling
		d.OnOpen(func() {
			// Send Open Messages
			sendErr := d.SendText("open")
			if sendErr != nil {
				panic(sendErr)
			}
		})
		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			message := string(msg.Data)
			Logger.Infof("Message from DataChannel '%s': '%s'", d.Label(), message)
			if d.Label() == "data" && message == "close" {
				Logger.Infof("User => %v close the Session", userName)
				session.RemoveUser(userName)
				err := peerConnection.Close()
				if err != nil {
					Logger.Error(err)
				}
			}
		})
	})

	// Set a handler for when a new remote track starts by our Peer
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		rTrackPayloadType := remoteTrack.PayloadType()
		rTrackCodec := remoteTrack.Codec()
		Logger.Infof("Track from %v with Codec: %v Payloadtype: %v", userName, rTrackCodec.Name, rTrackPayloadType)
		if rTrackPayloadType == customPayloadType {
			// Video Track
			// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
			// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
			go func() {
				ticker := time.NewTicker(rtcpPLIInterval)
				for range ticker.C {
					if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
						Logger.Error(rtcpSendErr)
					}
				}
			}()

			// Create a Buffer Loop
			rtpBuf := make([]byte, 1400)
			for {
				// Read remote Buffer
				i, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					Logger.Error(readErr)
				} else {
					// Push RTP Samples to GStreamer Pipeline with specific appsrc (user_id)
					session.VideoPipeline.WriteSampleToInputSource(rtpBuf[:i], newUser.ID)
				}
			}
		} else if remoteTrack.PayloadType() == mixedAudioTrack.PayloadType() {
			// Audio Track
			// Create a Buffer Loop
			rtpBuf := make([]byte, 1400)
			for {
				// Read remote Buffer
				i, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					Logger.Error(readErr)
					break
				} else {
					// Push RTP Samples to GStreamer Pipeline with specific appsrc (user_id)
					session.AudioPipeline.WriteSampleToInputSource(rtpBuf[:i], newUser.ID)
				}
			}
		} else {
			Logger.Error("OnTrack Codec not match...!")
			Logger.Errorf("	RemoteTrack=> Codec %v::%v", rTrackPayloadType, rTrackCodec.Name)
			Logger.Errorf("	VideoTrack => Codec %v::%v", mixedVideoTrack.PayloadType(), mixedVideoTrack.Codec().Name)
			Logger.Errorf("	AudioTrack => Codec %v::%v", mixedAudioTrack.PayloadType(), mixedAudioTrack.Codec().Name)
		}

	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Add the Session in Cache or Replace it.
	if _, found := s.SessionRegister.Get(s.GetSessionKey(sessID)); !found {
		err = s.SessionRegister.Add(s.GetSessionKey(sessID), session, cache.NoExpiration)
	} else {
		err = s.SessionRegister.Replace(s.GetSessionKey(sessID), session, cache.NoExpiration)
	}
	if err != nil {
		panic(err)
	}

	// Write Awnser to Client
	WriteJSON(w, answer)
}
