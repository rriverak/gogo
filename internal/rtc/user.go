package rtc

import (
	"github.com/pion/webrtc/v2"
	"github.com/rriverak/gogo/internal/utils"
)

//NewUser creates a new User
func NewUser(name string) User {
	return User{ID: utils.RandSeq(5), Name: name}
}

//User can Connect to a Session
type User struct {
	ID         string
	Name       string
	AudioTrack *webrtc.Track
	Peer       *webrtc.PeerConnection
}
