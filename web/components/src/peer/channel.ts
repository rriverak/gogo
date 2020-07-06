import { Dispatcher } from '../utils/dispatcher'
export abstract class PeerChannel {

    private isOpen:Promise<void>;
    constructor(
        private channel: RTCDataChannel,
    ) {
        this.isOpen = new Promise((resolve) => {
            this.channel.onopen = () => {
                resolve();
            }
        });

        this.channel.onmessage = (e) => {
            this.OnDataReceived(JSON.parse(e.data))
        };

    }
    SendText(msg:string){
        this.channel.send(msg);
    }
    SendJSON(msg:any){
        this.SendText(JSON.stringify(msg));
    }
    
    WaitForOpen():Promise<any>{
        return this.isOpen
    }

    protected OnDataReceived(_data:any) {}

    GetChannel():RTCDataChannel {
        return this.channel
    }
}

export class SessionChannel extends PeerChannel {
    private state:any;
    private stateDispatcher:Dispatcher;

    constructor(channel: RTCDataChannel){
        super(channel);
        this.stateDispatcher = new Dispatcher();
    }

    getSessionState():any{
        return this.state;
    }

    Open(){
        this.SendText("open")
    }

    Close(){
        this.SendText("close")
    }

    RequestNewState(){
        this.SendText("state")
    }

    AddListnerOnDataReceived(list){
        this.stateDispatcher.addListener("OnDataReceived", list)
    }

    protected OnDataReceived(data:any) {
        this.stateDispatcher.dispatch("OnDataReceived", data);
        this.state = data;
    }

    getDispatcher(){
        return this.stateDispatcher
    }

    setDispatcher(stateDispatcher){
        this.stateDispatcher = stateDispatcher;
    }
}