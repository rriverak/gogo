/* eslint-env browser */
var log = msg => {
  if (msg && msg.length > 0 ){
    document.getElementById('logs').innerHTML += '<li>' + msg + '</li>'
  }
}
var loadSessions = () => {
  fetch('/api/sessions/', {
    method: 'GET',
  }).then((resp) => {
    resp.json().then(data => {
      try {
        for (let index = 0; index < Object.keys(data).length; index++) {
          const roomKey = Object.keys(data)[index];
          const roomData = data[roomKey].Object;

          var el = document.createElement("button")
          el.setAttribute("class","btn btn-default hideAfterStart")
          el.innerHTML = roomData.ID + " <span class=\"badge\">" + roomData.Users.length + "</span>";
          el.onclick = () => {
            const userName = document.getElementById('txtUsername').value;
            window.createSession(roomData.ID, userName)
          };
          document.getElementById('sessions').appendChild(el)
        }
      } catch (e) {
        alert(e)
      }
    })
  });
};

window.createSession = (roomId, userName) => {
  document.getElementById('roomName').innerText = roomId;

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

  pc.onaddstream = function (event) {
    console.log(event)
    if (event.stream.id === "mixed") {
      var el = document.createElement("video")
      el.setAttribute("class", "embed-responsive-item")
      el.srcObject = event.stream
      el.autoplay = true
      el.controls = true
      document.getElementById('remoteVideos').appendChild(el)
    }
  }
  
  navigator.mediaDevices.getUserMedia(
    {
      video: {
        width: 320,
        height: 320,
      },
      audio: false
    })
    .then(stream => {
      pc.addStream(stream)
      pc.addTransceiver('video')
      pc.createOffer()
        .then(d => pc.setLocalDescription(d))
        .catch(log)
    }).catch(log)

  window.leaveSession = () => {
    pc.close();
  }

  window.startSession = () => {
    document.getElementById('logspanel').style = "display: block;";
    fetch('/api/sessions/'+roomId+'/' + userName, {
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


window.createRoom = ()=>{
  const userName = document.getElementById('txtUsername').value;
  const roomName = document.getElementById('txtRoom').value;
  window.createSession(roomName, userName)
}

loadSessions();
