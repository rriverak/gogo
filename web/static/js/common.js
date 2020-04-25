window.common = (function(){
    function newModal() {
        const model = {
            closeEventHandler: null,
            init: function () {
                var closemodal = document.querySelectorAll('.modal-close')
                for (var i = 0; i < closemodal.length; i++) {
                    closemodal[i].addEventListener('click', () => { this.close() })
                }
            },
            setField: function (text, query) {
                const field = document.querySelector(query)
                if (field) {
                    field.textContent = text;
                }
            },
            close: function () {
                if(this.closeEventHandler){
                    this.closeEventHandler()
                }
                const body = document.querySelector('body')
                const modal = document.querySelector('.modal')
                if (modal && body) {
                    modal.classList.add('opacity-0')
                    modal.classList.add('pointer-events-none')
                    body.classList.remove('modal-active')
                }

            },
            open: function () {
                const body = document.querySelector('body')
                const modal = document.querySelector('.modal')
                if (modal && body) {
                    modal.classList.remove('opacity-0')
                    modal.classList.remove('pointer-events-none')
                    body.classList.add('modal-active')
                }
            },
            onClose: function(handler) {
                this.closeEventHandler = handler;
            }
        }
        model.init();
        return model;
    }

    function getDevicesByKind(kind){
        const result = [];
        return new Promise(
            (resolve)=>{
                navigator.mediaDevices.enumerateDevices().then((devices)=>{
                    devices.forEach(function (device) {
                        if (device.kind == kind) {
                            result.push(device)
                        }
                    });
                    resolve(result)
                })
            }
        );
    }
    function getAudioDevices(){
        return getDevicesByKind("audioinput");
    }
    function getVideoDevices(){
        return getDevicesByKind("videoinput");
    }

    function getMediaStream(audioID, videoID){
        return navigator.mediaDevices.getUserMedia(
            {
                "audio": { deviceId: audioID },
                "video": {
                    deviceId: videoID,
                    width: 320,
                    height: 320,
                },
            })    
    }

    function createPeerConnection(){
        return new RTCPeerConnection({
            iceServers: [
                {
                    urls: 'stun:stun.l.google.com:19302'
                }
            ]
        })
    }

    function sendSDPOffer(sessionID, desc){
        return new Promise(
            (resolve)=>{
                fetch('/session/sdp/' + sessionID, {
                    method: 'POST',
                    body: btoa(JSON.stringify(desc))
                }).then((resp) => {
                    try {
                        resp.text().then((text) => {
                            const answer = JSON.parse(atob(text))
                            resolve(answer)
                        })
                    } catch (e) {
                        alert(e)
                    }
                });
            }
        );
    }

    return {
        NewModal: newModal,
        GetVideoDevices: getVideoDevices,
        GetAudioDevices: getAudioDevices,
        CreatePeerConnection: createPeerConnection,
        GetMediaStream: getMediaStream,
        SendSDPOffer: sendSDPOffer
    }
})();