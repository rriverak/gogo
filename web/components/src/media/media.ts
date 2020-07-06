export class MediaManager {
    constructor() { }

    async getDevicesByKindAsync(kind: string): Promise<MediaDeviceInfo[]> {
        const devices = await navigator.mediaDevices.enumerateDevices();
        return devices.filter(d => d.kind == kind);
    }
    async getAudioDevicesAsync(): Promise<MediaDeviceInfo[]> {
        return await this.getDevicesByKindAsync("audioinput");
    }
    async getVideoDevicesAsync(): Promise<MediaDeviceInfo[]> {
        return await this.getDevicesByKindAsync("videoinput");
    }

    async getMediaStream(audioID, videoID): Promise<MediaStream> {
        const constrain = {
            "audio": { deviceId: audioID },
            "video": {
                deviceId: videoID,
                width: 320,
                height: 320,
            },
        };
        return await navigator.mediaDevices.getUserMedia(constrain);
    }
}