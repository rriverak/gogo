# [WIP] GoGo
A simple Video GroupCall Server.  
Most of it is in Golang.  

Based on the great `pion/web-rtc` and `GStreamer` for Video Processing.  
This GroupCall is use the Idea behinde MCU and Mix Video and Audio Streams together.  

## Actions
- Create a Room
- Join a Room
- Leave a Room
- Delete a Room

## Load & Run
```bash
go mod verify
go run -v ./cmd/server/main.go
```         

## Serve
Open in Browser http://localhost:8080/

## Code-Support
In this Version only Users with the same Codec in a Session are able to Join.  
Currently is the preferred Codec `VP8`. This Codec is a widely used Codec and we have `h264` as a fallback.
### Video
- VP8
- VP9 (Work in Progress)
- H264
### Audio
- Opus
  
## Browsers
Cross-Browsers Calls like `Safari - Chrome - Firefox` are Supported.

### Supported
- Firefox
- Chrome
- Safari 13  
  
## License
[MIT](/LICENSE.md)
