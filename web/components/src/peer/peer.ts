import { SessionChannel } from './channel'

export class PeerConnection {
    peer: RTCPeerConnection;
    sessionChannel: SessionChannel;

    constructor(private videoTag: HTMLVideoElement) {
        this.peer = new RTCPeerConnection({
            iceServers: [
                {
                    urls: 'stun:stun.l.google.com:19302'
                }
            ]
        });
        this.peer.ontrack = (e) => { this.onMediaTrackHandler(e) }
        this.peer.ondatachannel = (e) => { this.onDataChannelHandler(e) }
        //Add Transceiver
        this.peer.addTransceiver('video')
        this.peer.addTransceiver('audio')

        //DataChannel
        this.sessionChannel = new SessionChannel(this.peer.createDataChannel("session"))

    }

    onMediaTrackHandler(event) {
        if (event && event.streams) {
            const stream = event.streams[0]
            if (stream.id === "video-pipe") {
                console.log("Add Remote Video Stream")
                this.videoTag.srcObject = stream;
            }
            if (stream.id === "audio-pipe") {
                console.log("Add Remote Audio Stream")
                if (this.videoTag.srcObject) {
                    const obj = this.videoTag.srcObject as MediaStream;
                    obj.addTrack(event.track)
                }
            }
        }
    }

    onDataChannelHandler(e) {
        if (e.channel.label == "session") {
            let dispatcher = null
            if (this.sessionChannel) {
                dispatcher = this.sessionChannel.getDispatcher();
            }

            this.sessionChannel = new SessionChannel(e.channel)
            if (dispatcher) {
                this.sessionChannel.setDispatcher(dispatcher)
            }
            this.sessionChannel.WaitForOpen().then(() => {
                this.sessionChannel.Open();
                this.sessionChannel.RequestNewState();
            })
        }
    }

    AddTrack(track, stream): RTCRtpSender {
        return this.peer.addTrack(track, stream)
    }

    ListenOnNewSessionState(listner): any {
        return this.sessionChannel.AddListnerOnDataReceived(listner)
    }


    getSessionState(): any {
        return this.sessionChannel.getSessionState()
    }

    async getSDPOffer() {
        const offer = await this.peer.createOffer();
        await this.peer.setLocalDescription(offer);
        return offer;
    }

    async sendSDPOffer(sessionID, desc) {
        const resp = await fetch('/session/sdp/' + sessionID, {
            method: 'POST',
            body: btoa(JSON.stringify(desc))
        });

        const text = await resp.text()
        return JSON.parse(atob(text));

    }

    async ConnectToSession(sessionID: string) {
        const offer = await this.getSDPOffer();
        const awnser = await this.sendSDPOffer(sessionID, offer);
        await this.peer.setRemoteDescription(awnser);
    }

}