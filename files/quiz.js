window.addEventListener("load", function (evt) {
    let output = document.getElementById("output");
    let ws;
    let waitingForAnswer = true
    let current = 0;
    let total = 1;
    let count = 1;
    let previousQuestion = false;
    let currentServerIndex = 0;

    const currentTheme = localStorage.getItem('theme') ? localStorage.getItem('theme') : null;

    if (currentTheme) {
        document.documentElement.setAttribute('data-theme', currentTheme);

        if (currentTheme === 'dark') {
            document.getElementById('body').setAttribute('data-theme', 'dark');
            document.getElementById('html').setAttribute('data-theme', 'dark');
            document.getElementById('play').setAttribute('data-theme', 'dark');
            document.getElementById('fullscreen').setAttribute('data-theme', 'dark');
        } else {
            document.getElementById('body').setAttribute('data-theme', 'light');
            document.getElementById('html').setAttribute('data-theme', 'light');
            document.getElementById('play').setAttribute('data-theme', 'light');
            document.getElementById('fullscreen').setAttribute('data-theme', 'light');
        }
    }

    function baseName(str)
    {
        let base = new String(str).substring(str.lastIndexOf('/') + 1);
        if(base.lastIndexOf(".") !== -1)
            base = base.substring(0, base.lastIndexOf("."));
        return base;
    }
    let pathArr = window.location.pathname.split('/');
    let course = pathArr[2].toLowerCase();
    let category = baseName(window.location.pathname).toLowerCase();

    console.log("protocol", location.protocol);

    if (location.protocol == "https:") {
        ws = new WebSocket("wss://" + location.host + "/connect?course="+course);
    } else {
        ws = new WebSocket("ws://" + location.host + "/connect?course="+course);
    }

    // load the flagged indices from local storage
    let flagged = JSON.parse(localStorage.getItem(category));
    console.log("initial flagged indices", flagged);

    const urlParams = new URLSearchParams(window.location.search);
    const flaggingEnabled = urlParams.get('flagged');

    console.log("flaggingEnabled", flaggingEnabled);

    if (!flaggingEnabled) {
        console.log("ignore flags")
        flagged = [];
    }

    if (flagged == null) {
        flagged = [];
    }

    ws.onopen = function (evt) {
        if (flagged == null) {
            ws.send(window.location.pathname + ";");
        } else {
            ws.send(window.location.pathname + ";" + flagged.join(","));
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

    let previous = function() {
        if (current === 1) {
            return;
        }
        if (ws) {
            previousQuestion = true;
            ws.send("previous")
        }
    }

    let flag = function() {
        console.log("flagged", flagged);

        if (currentServerIndex === "XXX") {
            return;
        }

        if (!flagged.includes(currentServerIndex)) {

            flagged.push(currentServerIndex);
            localStorage.setItem(category, JSON.stringify(flagged));
            addFlag(currentServerIndex);
            console.log("flagged", currentServerIndex, category);
        }
    }

    let addFlag = function() {
        console.log("addFlag", currentServerIndex)
        let elem = document.getElementById("flag");
        if (elem == null) {
            let d = document.createElement("div");
            d.className = "circle";
            d.id = "flag";
            document.getElementById("index").appendChild(d);
        }
    }

    function remove(array, element) {
        return array.filter(el => el !== element);
    }

    let unflag = function() {
        console.log("flagged before unflag", flagged);

        if (flagged.includes(currentServerIndex)) {

            flagged = remove(flagged, currentServerIndex)

            localStorage.setItem(category, JSON.stringify(flagged));
            clearFlag();
            console.log("unflagged", currentServerIndex);
        }

        console.log("flagged after unflag", flagged);
    }

    let clearFlag = function() {
        console.log("clearFlag", currentServerIndex)
        let elem = document.getElementById("flag");
        if (elem != null) {
            console.log(elem);
            elem.parentNode.removeChild(elem);
        }
    }

    let next = function() {
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
    let elem = document.getElementById("body");

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

    let noSleep = new NoSleep();

    function enableNoSleep() {
        noSleep.enable();
        document.removeEventListener('touchstart', enableNoSleep, false);
    }

    // Enable wake lock.
    // (must be wrapped in a user input event handler e.g. a mouse or touch handler)
    document.addEventListener('touchstart', enableNoSleep, false);

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
            window.location =  "/courses/"+course;
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
        }, 7000);
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

    // TODO: integrate into checkKey() function
    document.addEventListener("keydown", function(event){
        switch(event.keyCode){
            case 9:
                event.stopPropagation();
                event.preventDefault();
                console.log("TAB KEY PRESSED");
                next()
            case 33: //left or previous
                console.log("left");
            case 34: //right or next
                console.log("right");
            case 27: //start or play
                console.log("play");
            case 116: //stop or exit
                console.log("stop");
        }
    });

    let print = function (message) {

        let target = output;

        if (waitingForAnswer) {
            flush();
            target = document.createElement("B");
        }

        let lines = message.split("\n");

        for (var i in lines) {
            let d = document.createElement("p");

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

    let flush = function () {
      output.innerHTML= "";
    };
});