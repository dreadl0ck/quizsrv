window.addEventListener("load", function (evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var waitingForAnswer = true
    var current = 0;
    var total = 1;
    var count = 1;
    var previousQuestion = false;

    console.log("protocol", location.protocol);

    if (location.protocol == "https:") {
        ws = new WebSocket("wss://" + location.host + "/cia/connect");
    } else {
        ws = new WebSocket("ws://" + location.host + "/cia/connect");
    }

    ws.onopen = function (evt) {
      ws.send(window.location);
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
        }
        else if (e.keyCode == '40') {
            // down arrow
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
          print(evt.data);
      }
    };
    ws.onerror = function (evt) {
      print("ERROR: " + evt.data);
    };

    document.addEventListener("touchstart", next, false);

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