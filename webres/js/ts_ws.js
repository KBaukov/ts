ws = new WebSocket('wss://localhost:8443/ws');
ws.onerror = function(event) {
    notify('error','WS соединение с сервером не установленно.');
    console.log('ws: Error');
};

ws.onopen = function(e) {
    e.target.send("WS: connection success/");
};

ws.onmessage = function(ws) {
    var svgRoot  = getSvgRoot();
    var msg = ws.data;
    console.log('ws:' + msg);
    var cmd = jQuery.parseJSON(msg);
    if(cmd) {
        if(cmd.action == 'seatStateUpdate') {
            updateSeatState(cmd.data);
        }

        if(cmd.action == 'sysreserve' || cmd.action == 'sysunreserve') {
            //cmd.seat
            var s = $("path#"+cmd.seat, svgRoot);
            if(s[0]) {
                if(cmd.state==0) {
                    s[0].classList.remove('seatOcup');
                    s[0].style.fillOpacity = 0.1;
                }
                if(cmd.state==1) {
                    s[0].classList.add('seatOcup');
                    s[0].style.fillOpacity = 0.9;
                }
            }
        }
        if(cmd.action == 'sysunreserves') {
            if(cmd.seats) {
                var seats = cmd.seats;
                for(var i=0; i<seats.length; i++) {
                    var ss = seats[i];
                    var s = $("path#"+ss.seat, svgRoot);
                    if(s[0]) {
                        if(ss.state==0) {
                            s[0].classList.remove('seatOcup');
                            s[0].classList.remove('seatSel');
                            s[0].style.fillOpacity = 0.1;
                        }
                        // if(ss.state==1) {
                        // 	s[0].classList.add('seatOcup');
                        // 	s[0].style.fillOpacity = 0.9;
                        // }
                    }
                }
            }
        }
        if(cmd.action == 'reserve') {
            if(cmd.success) {
                var s = $("path#"+cmd.seat, svgRoot);
                if(s[0]) {
                    var ss = s[0];
                    ss.classList.add('seatSel');
                    ss.style.fillOpacity = 0.6;
                    addSeatForPay(ss);
                    //console.log(s.id);
                }
            }
        }
    }
};

