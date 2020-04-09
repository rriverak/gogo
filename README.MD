# [WIP] GoGo
A simple Video GroupCall Server.  
Most of it is written in Golang.  

Based on the great `pion/web-rtc` and `GStreamer` for Video Processing.

## Load & Run
```bash
go mod verify
go run -v ./cmd/server/main.go
```         

## Serve
Open in Browser http://localhost:8080/

## Code-Support
### Video
- VP8
- VP9
- H264
### Audio
- Opus

## Todo
### App
- [ ] Extend Config 

### UI
- [x] Vanilla Test UI
- [ ] Anular WebUI
- [ ] Stenciljs / Ionic PWA
- [ ] Health / Status Page

### VideoCall
- [x] VideoMixer
- [x] AudioMixer 
- [ ] User´s name instead of ID in Video Overlay
- [ ] User´s AudioTrack not on Server. 

### WebAPI
- [ ] Authentication
