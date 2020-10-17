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

    // hook flag toggling to flag button
    document.getElementById("flagButton").onclick = function(event) {
        event.preventDefault();
        event.stopPropagation();
        if (flagged != null) {
            if (!flagged.includes(currentServerIndex)) {
                flag(currentServerIndex);
            } else {
                unflag(currentServerIndex);
            }
        }
    };

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

    var flag = function(index) {
        console.log("flagged", flagged);

        if (!flagged.includes(index)) {

            flagged.push(index);
            localStorage.setItem(location.pathname.replaceAll("/cia/", ""), JSON.stringify(flagged));
            addFlag(index);
            console.log("flagged", index);
        }
    }

    var addFlag = function(index) {
        console.log("addFlag", index)
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

    var unflag = function(index) {
        console.log("flagged before unflag", flagged);

        if (flagged.includes(index)) {

            flagged = remove(flagged, index)

            localStorage.setItem(location.pathname.replaceAll("/cia/", ""), JSON.stringify(flagged));
            clearFlag(index);
            console.log("unflagged", index);
        }

        console.log("flagged after unflag", flagged);
    }

    var clearFlag = function(index) {
        console.log("clearFlag", index)
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

    function checkKey(e) {

        e = e || window.event;

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
                  addFlag(currentServerIndex)
              } else {
                  clearFlag(currentServerIndex);
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

    document.addEventListener("tap", next, false);

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