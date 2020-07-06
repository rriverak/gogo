import { Component, Host, h, State, Prop } from '@stencil/core';
import { MediaManager } from "../../media/media";
import { PeerConnection } from "../../peer/peer";
@Component({
  tag: 'gogo-video',
  styleUrl: 'gogo-video.css',
  shadow: true,
})
export class GogoVideo {
  @Prop() SessionId: string;
  
  @State() SessionState: any;
  @State() VideoID: string = "";
  @State() AudioID: string = "";
  @State() IsSidebar: boolean = true;
  @State() IsDeviceSelection: boolean = true;
  @State() IsDebug: boolean = true;

  private videoElement: HTMLVideoElement;

  MediaManager: MediaManager = new MediaManager();

  componentDidLoad() {
  }

  openDeviceSelection(){
    this.IsDeviceSelection = true;
  }
  closeDeviceSelection(event:CustomEvent){
    if(event.detail){
      event.preventDefault()
      this.AudioID = event.detail.audio;
      this.VideoID = event.detail.video;
      this.IsDeviceSelection = false;
      this.StartAsync()
    }
  }


  toggleSidebar() {
    this.IsSidebar = !this.IsSidebar;
  }

  toggleDebug() {
    this.IsDebug = !this.IsDebug;
  }

  async StartAsync() {
    if(this.SessionId){
      const peer: PeerConnection = new PeerConnection(this.videoElement);
      peer.ListenOnNewSessionState((state)=>{
        this.SessionState = state;
      })

      const streams = await this.MediaManager.getMediaStream(this.AudioID, this.VideoID);
      streams.getTracks().forEach((track) => peer.AddTrack(track, streams));

      await peer.ConnectToSession(this.SessionId);
    }
  }

  render() {
    return (
      <Host>
        {this.IsDeviceSelection ? <gogo-video-device-selection onClose={(e)=>{ this.closeDeviceSelection(e) }}></gogo-video-device-selection> : null}
        <div class="video-box">
          <div class="video-wrapper">
            <div class="overlay top">
              <div class="toolbar">
                <button type="button" onClick={() => this.toggleDebug()} class="green">Debug</button>
                <button type="button" onClick={() => this.toggleSidebar()} class="blue">Info</button>
              </div>
            </div>
            {this.IsDebug ? <gogo-video-debug></gogo-video-debug> : null}
            <div class="video-viewer">
              <video ref={(el) => this.videoElement = el as HTMLVideoElement} autoplay playsinline></video>
            </div>
            {this.IsSidebar && this.SessionState ? <gogo-video-sidebar state={this.SessionState}></gogo-video-sidebar> : null}
            <div class="overlay bottom">
              <div class="toolbar">
                <button type="button" class="red">Close</button>
              </div>
            </div>
          </div>
        </div>
      </Host>
    );
  }

}
