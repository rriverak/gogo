
(function(){
    const log = function(text){
        const logsBody = document.getElementById('logs').querySelector("tbody");
        if(logsBody){
            let row = document.createElement("tr");
            let col = document.createElement("td");
            col.textContent = text
            row.append(col);
            logsBody.append(row)
        }
    }
    const sessionPage = {
        startSession(vDeviceID, aDeviceID){
            log("Create a PeerConnection")
            const peer = window.common.CreatePeerConnection()
            peer.ontrack = function (event) {
                if (event && event.streams) {
                    const stream = event.streams[0]
                    if (stream.id === "video-pipe") {
                        log("Add Remote Video Stream")
                        let el = document.getElementById('main-video')
                        el.srcObject = event.streams[0]
                        el.autoplay = true
                        el.controls = false
                        el.muted = true
                        el.setAttribute("playsinline", "")
                        //el.setAttribute("class", "w-full rounded-lg")
                        document.getElementById('video-body').appendChild(el)
                    }
                    if (stream.id === "audio-pipe") {
                        log("Add Remote Audio Stream")
                        let el = document.getElementById('main-video')
                        if (el.srcObject) {
                            el.srcObject.addTrack(event.track)
                        }
                    }
                }
            }

            window.common.GetMediaStream().then((stream)=>{
                stream.getTracks().forEach(function (track) {
                    peer.addTrack(track, stream);
                });
                peer.addTransceiver('video')
                peer.addTransceiver('audio')
                peer.createOffer().then(desc => {
                        peer.setLocalDescription(desc).then(() => {
                            log("Prepeare a Connection")
                            window.common.SendSDPOffer(window.session_id, desc).then((answer)=>{
                                log("Setting Remote Connection and wait for Track...")
                                peer.setRemoteDescription(new RTCSessionDescription(answer))
                            })
                        })
                    })
                    .catch((msg) => { 
                        log(msg); 
                    })
            })
        },
        onLoad: function(){
            // Init Modal
            window.common.GetVideoDevices().then((devices)=>{
                devices.map(function(device){
                    const option = document.createElement("option");
                    option.title = device.label
                    option.value = device.deviceId
                    option.textContent = "Camera: " + option.title
                    document.getElementById('video_devices').append(option)
                })
            })
            window.common.GetAudioDevices().then((devices)=>{
                devices.map(function(device){
                    const option = document.createElement("option");
                    option.title = device.label
                    option.value = device.deviceId
                    option.textContent = "Mic: " + option.title
                    document.getElementById('audio_devices').append(option)
                });
            })

            const btnLog = document.getElementById("btnLogs")
            btnLog.onclick = function(){
                    sessionPage.toggleLog();
            }
        

            modal.open();
    
        },
        toggleLog: function(){
            const logs = document.getElementById("logs");
            if (logs){
                logs.classList.toggle('hidden')
            }
        },
        toggleSidebar: function(){
            const logs = document.getElementById("logs");
            if (logs){
                logs.classList.toggle('hidden')
            }
        }     
    }
    const modal = window.common.NewModal()
    modal.onClose(function(){
        const aDevices = document.getElementById('audio_devices')
        const vDevices = document.getElementById('audio_devices')
        const aDeviceID = aDevices.options[aDevices.selectedIndex].value;
        const vDeviceID = vDevices.options[vDevices.selectedIndex].value;
        sessionPage.startSession(vDeviceID, aDeviceID);
    });

    window.addEventListener('load', sessionPage.onLoad)
})();