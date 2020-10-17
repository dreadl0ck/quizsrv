window.addEventListener("load", function (evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var waitingForAnswer = true
    var current = 0;
    var total = 1;
    var count = 1;
    var previousQuestion = false;
    var currentServerIndex = 0;

    console.log("protocol", location.protocol);

    if (location.protocol == "https:") {
        ws = new WebSocket("wss://" + location.host + "/cia/connect");
    } else {
        ws = new WebSocket("ws://" + location.host + "/cia/connect");
    }

    // load the flagged indices from local storage
    var flagged = JSON.parse(localStorage.getItem(location.pathname.replaceAll("/cia/", "")));
    console.log("initial flagged indices", flagged);

    if (flagged == null) {
        flagged = [];
    }

    ws.onopen = function (evt) {
        if (flagged == null) {
            ws.send(window.location + ";");
        } else {
            ws.send(window.location + ";" + flagged.join(","));
        }
    };

    // hook flag toggling to right mouse click
    document.addEventListener('contextmenu', function(event) {
        event.preventDefault();
        event.stopPropagation();
        if (flagged != null) {
            if (!flagged.includes(currentServerIndex)) {
                flag(currentServerIndex);
            } else {
                unflag(currentServerIndex);
            }
        }
    }, false);

    document.onkeydown = checkKey;

    var previous = function() {
        if (current == 1) {
            return;
        }
        if (ws) {
            previousQuestion = true;
            ws.send("previous")
        }
    }

    var flag = function() {
        console.log("flagged", flagged);

        if (!flagged.includes(currentServerIndex)) {

            flagged.push(currentServerIndex);
            localStorage.setItem(location.pathname.replaceAll("/cia/", ""), JSON.stringify(flagged));
            addFlag(currentServerIndex);
            console.log("flagged", currentServerIndex);
        }
    }

    var addFlag = function() {
        console.log("addFlag", currentServerIndex)
        var elem = document.getElementById("flag");
        if (elem == null) {
            var d = document.createElement("div");
            d.className = "circle";
            d.id = "flag";
            document.getElementById("index").appendChild(d);
        }
    }

    function remove(array, element) {
        return array.filter(el => el !== element);
    }

    var unflag = function() {
        console.log("flagged before unflag", flagged);

        if (flagged.includes(currentServerIndex)) {

            flagged = remove(flagged, currentServerIndex)

            localStorage.setItem(location.pathname.replaceAll("/cia/", ""), JSON.stringify(flagged));
            clearFlag();
            console.log("unflagged", currentServerIndex);
        }

        console.log("flagged after unflag", flagged);
    }

    var clearFlag = function() {
        console.log("clearFlag", currentServerIndex)
        var elem = document.getElementById("flag");
        if (elem != null) {
            console.log(elem);
            elem.parentNode.removeChild(elem);
        }
    }

    var next = function() {
        previousQuestion = false;
        if (waitingForAnswer) {
            if (ws) {
                ws.send("answer")
                waitingForAnswer = false;
            }
        } else {
            if (ws) {
                ws.send("next")
                waitingForAnswer = true;
            }
        }
    }

    /* Get the element you want displayed in fullscreen mode (a video in this example): */
    var elem = document.getElementById("body");

    /* When the openFullscreen() function is executed, open the video in fullscreen.
    Note that we must include prefixes for different browsers, as they don't support the requestFullscreen method yet */
    function openFullscreen() {
        if (elem.requestFullscreen) {
            elem.requestFullscreen();
        } else if (elem.mozRequestFullScreen) { /* Firefox */
            elem.mozRequestFullScreen();
        } else if (elem.webkitRequestFullscreen) { /* Chrome, Safari and Opera */
            elem.webkitRequestFullscreen();
        } else if (elem.msRequestFullscreen) { /* IE/Edge */
            elem.msRequestFullscreen();
        }
    }

    document.getElementById('fullscreen').onclick = openFullscreen;

    function checkKey(e) {

        e = e || window.event;

        console.log(e.key);

        if (e.keyCode == '70') {
            openFullscreen();
            return false;
        }

        if (e.keyCode == '32') {
            // space
            document.getElementById('play').click();
            return false;
        }

        if(e.key === "Escape") {
            console.log("MENU");
            window.location =  "/cia";
            return false;
        }

        if (e.keyCode == '38') {
            // up arrow
            flag(currentServerIndex);
        }
        else if (e.keyCode == '40') {
            // down arrow
            unflag(currentServerIndex);
        }
        else if (e.keyCode == '37') {
           // left arrow
           previous();
        }
        else if (e.keyCode == '39') {
           // right arrow
           next();
        }
    }

    var autoplayActive = false;
    var autoplay = {};

    function invokeNext() {
        console.log("invokeNext");
        autoplay = setTimeout(function() {
            next();
            invokeNext();
        }, 5000);
    }

    var toggleAutoplay = function (ev) {

        console.log("toggleAutoplay");

        if (autoplayActive) {
            document.getElementById("play").className = "button";
        } else {
            document.getElementById("play").className = "button paused";
        }

        ev.preventDefault();
        ev.stopPropagation();

       if (autoplayActive) {
           clearTimeout(autoplay);
           autoplayActive = false;
       } else {
           invokeNext();
           autoplayActive = true;
       }
    }

    document.getElementById("play").onclick = toggleAutoplay;

    ws.onclose = function (evt) {
      ws = null;
    };
    ws.onmessage = function (evt) {
      if (evt.data.startsWith("total=")) {
        total = evt.data.replace("total=", "");
        document.getElementById("index").innerHTML = current + "/" + total
      } else {

          console.log("got", evt.data);

          if (previousQuestion) {

              flush();
              if (waitingForAnswer) {
                  count -= 2;
              } else {
                  count -= 3;
              }

              waitingForAnswer = true;

              if (count % 2 == 0 && current <= total) {
                  current--;
                  document.getElementById("index").innerHTML = current + "/" + total
              }
          } else {
              count++;
              if (count % 2 == 0 && current < total) {
                  current++;
                  document.getElementById("index").innerHTML = current + "/" + total
              }
          }


          // handle server index
          var arr = evt.data.split(":");
          currentServerIndex = arr[0].replace("index=", "");

          if (flagged != null) {
              console.log(flagged);

              // if the current server index is flagged, add the dot
              // if not: remove it
              if (flagged.includes(currentServerIndex)) {
                  addFlag()
              } else {
                  clearFlag();
              }
          }

          // remove the first elem (server index)
          arr.shift();

          // join the remainder
          var msg = arr.join(":");

          console.log("got message with server index", currentServerIndex, "message", msg);

          print(msg);
      }
    };
    ws.onerror = function (evt) {
      print("ERROR: " + evt.data);
    };

    document.addEventListener('swiped-right', previous);
    document.addEventListener('swiped-left', next);
    document.addEventListener('swiped-up', flag);
    document.addEventListener('swiped-down', unflag);

    document.touchmove = function(e)
    {
        e.preventDefault();
    };

    var print = function (message) {

        var target = output;

        if (waitingForAnswer) {
            flush();
            target = document.createElement("B");
        }

        var lines = message.split("\n");

        for (var i in lines) {
          var d = document.createElement("p");

          if (!waitingForAnswer) {
              d.className = "red";
          }

          //d.textContent = lines[i];
          d.innerHTML = lines[i];

          target.appendChild(d);
        }

        if (waitingForAnswer) {
            output.appendChild(target);
        }
    };

    window.addEventListener('click', event => {
        next();
    });

    var flush = function () {
      output.innerHTML= "";
    };
});