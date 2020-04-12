/* eslint-env browser */
var log = msg => {
  if (msg && msg.length > 0) {
    document.getElementById('logs').innerHTML += '<li>' + msg + '</li>'
  }
}
window.loadSessions = () => {
  fetch('/api/sessions/', {
    method: 'GET',
  }).then((resp) => {
    resp.json().then(data => {
      try {
        document.getElementById("tbodySession").innerHTML = ""
        for (let index = 0; index < Object.keys(data).length; index++) {
          const roomKey = Object.keys(data)[index];
          const roomData = data[roomKey].Object;

          var btnJoin = document.createElement("button")
          btnJoin.setAttribute("class", "btn btn-default btn-sm")
          btnJoin.innerText = "Join";
          btnJoin.onclick = () => {
            const userName = document.getElementById('txtUsername').value;
            window.createSession(roomData.ID, userName)
          };

          var btnDelete = document.createElement("button")
          btnDelete.setAttribute("class", "btn btn-default btn-sm")
          btnDelete.innerText = "Delete";
          btnDelete.onclick = () => {
            const userName = document.getElementById('txtUsername').value;
            window.deleteSession(roomData.ID)
          };

          var tr = document.createElement("tr");
          var tdID = document.createElement("td");
          var tdUsers = document.createElement("td");
          var tdActions = document.createElement("td");

          tdID.innerText = roomData.ID
          tdUsers.innerText = roomData.Users.length
          tdActions.appendChild(btnJoin)
          tdActions.appendChild(btnDelete)

          tr.appendChild(tdID)
          tr.appendChild(tdUsers)
          tr.appendChild(tdActions)

          document.getElementById("tbodySession").appendChild(tr)
          //document.getElementById('sessions').appendChild(el)
        }
      } catch (e) {
        alert(e)
      }
    })
  });
};

window.createSession = (roomId, userName) => {
  document.getElementById('roomName').innerText = roomId;
  let dc = null;
  let pc = new RTCPeerConnection({
    iceServers: [
      {
        urls: 'stun:stun.l.google.com:19302'
      }
    ]
  })
  pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
  pc.onicecandidate = event => {
    if (event.candidate === null) {
      window.startSession(pc.localDescription)
    }
  }
  pc.ontrack = function (event) {
    if (event && event.streams) {
      const stream = event.streams[0]
      if (stream.id === "video-pipe") {
        console.log("Add Video Stream")
        var el = document.createElement("video")
        el.srcObject = event.streams[0]
        el.id = "main-video"
        el.autoplay = true
        el.controls = true
        el.muted = true
        el.onerror = (err) => { console.log(err) }
        el.onplaying = (ev) => { console.log(ev) }
        document.getElementById('remoteVideos').appendChild(el)
      }

      if (stream.id === "audio-pipe") {
        console.log("Add Audio Track")
        let el = document.getElementById('main-video')
        if (el.srcObject) {
          el.srcObject.addTrack(event.track)
        }
      }
    }
  }
  /*
  pc.onaddstream = function (event) {
    if (event.stream.id === "mixed") {
      var el = document.createElement("video")
      el.setAttribute("class", "embed-responsive-item")
      el.srcObject = event.stream
      el.autoplay = true
      el.controls = true
      document.getElementById('remoteVideos').appendChild(el)
    }
  }
  */
 
  let sessionSendChannel = pc.createDataChannel('session', {negotiated: true, id:1})
 sessionSendChannel.onmessage = e => log(`Message from '${sessionSendChannel.label}' Channel. Payload => '${e.data}'`)
 sessionSendChannel.onclose = () => log(`${sessionSendChannel.label} has closed`)
 sessionSendChannel.onopen = () => {
  log(`${sessionSendChannel.label} has opened`);
  sessionSendChannel.send("open")
 }

 pc.ondatachannel = e => {
    const chan = e.channel;
    chan.onclose = () => log(`${chan.label} has closed`)
    chan.onopen = () => log(`${chan.label} has opened`)
  
    chan.onmessage = e => log(`Message from DataChannel '${chan.label}' payload '${e.data}'`)
    console.log("DataChannel => " + chan.label)
  }

  navigator.mediaDevices.getUserMedia(
    {
      "audio": true,
      "video": {
        width: 320,
        height: 320,
      },
    }).then(stream => {
      stream.getTracks().forEach(function (track) {
        pc.addTrack(track, stream);
      });
      pc.addTransceiver('video')
      pc.addTransceiver('audio')
      pc.createOffer()
        .then(d => { pc.setLocalDescription(d) && console.log(d.sdp)})
        .catch((msg) => { log; })
    }).catch((msg) => { log; })

  window.leaveSession = () => {
    sessionSendChannel.send("close")
    pc.close();
    window.location.reload()
  }

  window.startSession = () => {
    document.getElementById('logspanel').style = "display: block;";
    fetch('/api/sessions/' + roomId + '/' + userName, {
      method: 'POST',
      body: btoa(JSON.stringify(pc.localDescription))
    }).then((resp) => {
      resp.json().then(data => {
        try {
          pc.setRemoteDescription(new RTCSessionDescription(data))
        } catch (e) {
          alert(e)
        }
      })
    });
  }


  let hideElms = document.getElementsByClassName('hideAfterStart')
  for (let i = 0; i < hideElms.length; i++) {
    hideElms[i].style = 'display: none'
  }
  let showElms = document.getElementsByClassName('showAfterStart')
  for (let i = 0; i < showElms.length; i++) {
    showElms[i].style = 'display: block'
  }

}


window.createRoom = () => {
  const userName = document.getElementById('txtUsername').value;
  const roomName = document.getElementById('txtRoom').value;
  window.createSession(roomName, userName)
}

window.deleteSession = (roomId) =>{
  fetch("/api/sessions/" + roomId.replace("session-")  ,{
    method: 'DELETE',
  }).then(window.loadSessions())
}

window.loadSessions();
