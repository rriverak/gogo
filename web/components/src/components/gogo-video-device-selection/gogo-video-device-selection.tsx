import { Component, Host, h, EventEmitter, Event, State } from '@stencil/core';
import { MediaManager } from "../../media/media";

@Component({
  tag: 'gogo-video-device-selection',
  styleUrl: 'gogo-video-device-selection.css',
  shadow: true,
})
export class GogoVideoDeviceSelection {

  MediaManager: MediaManager = new MediaManager();
  @State() AudioDevices:MediaDeviceInfo[] = [];
  @State() VideoDevices:MediaDeviceInfo[] = [];

  constructor(){
  }
  devices: any = { video: "", audio: "" };
  componentDidLoad() {
    this.loadDevices()
  }

  async loadDevices(){
    this.AudioDevices = await this.MediaManager.getAudioDevicesAsync()
    this.VideoDevices = await this.MediaManager.getVideoDevicesAsync()
  }

  setVideoDevice(deviceId){
    this.devices["video"] = deviceId
  }
  setAudioDevice(deviceId){
    this.devices["audio"] = deviceId
  }
  @Event() close: EventEmitter;
  closeHandler() {
    this.close.emit(this.devices);
  }

  OnModalClose(){
    this.closeHandler()
  }

  render() {
    return (
      <Host>
        <gogo-video-modal onClose={() => {this.OnModalClose()}} header={"Devices"}>
          <div>
            <label>Video Device</label>
            <select onInput={(event:any) => this.setVideoDevice(event.target.value)}>
              {this.VideoDevices.map(d => {
                return <option selected={this.devices.video === d.deviceId} value={d.deviceId}>{d.label || d.deviceId || 'Unknown'}</option>
              })}
            </select>
          </div>
          <div>
            <label>Audio Device</label>
            <select onInput={(event:any) => this.setAudioDevice(event.target.value)}>
              {this.AudioDevices.map(d => {
                return <option selected={this.devices.audio === d.deviceId} value={d.deviceId}>{d.label || d.deviceId || 'Unknown'}</option>
              })}
            </select>
          </div>
        </gogo-video-modal>
      </Host>
    );
  }

}
