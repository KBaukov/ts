var addSeatForPay = function(node) {
	var sklad = $("div#tForPay");
	if(sklad[0])
		sklad[0].classList.add(node);
	else
		return null;
};


var removeSeatFromPay = function(node) {
	var sklad = $("div#tForPay");
	if(sklad[0])
		sklad[0].classList.remove(node);
};

var getSeatForPay = function() {
	var sklad = $("div#tForPay");
	if(sklad[0] && sklad[0].classList.length>0)
		return sklad[0].classList;//.split(' ');
};
var opSw1 = function(e) {
	if(!e.target.classList.contains('seatOcup') && !e.target.classList.contains('seatSel') )
		e.target.style.fillOpacity = 0.6;
};
var opSw2 = function(e) {
	if(!e.target.classList.contains('seatOcup') && !e.target.classList.contains('seatSel') )
		e.target.style.fillOpacity = 0.1;
};
var opSwOn = function(e) { e.target.style.fillOpacity = 0.3;};
var opSwOff = function(e) { e.target.style.fillOpacity = 0.75;};
//MMXX=3;
var selectSeat = function(e) {
	var s = e.target;

	if(s.classList.contains('seatOcup')) {
		notify('warn','Извините это место уже кем-то занято.');
		return false;
	}
	if(s.classList.contains('seatSel')) {
		ws.send('{"action":"unreserve", "seat_id":"'+s.id+'"}');
		s.classList.remove('seatSel');
		removeSeatFromPay(s.id);
	} else {
		ws.send('{"action":"reserve", "seat_id":"'+s.id+'"}');
		s.classList.add('seatSel');
		s.style.fillOpacity = 0.6;
		addSeatForPay(s.id);
		//console.log(s.id);
	}
	//console.log('z2r'+MMXX+s.id+'$'+s.attributes.d.value);

	return false;
};

var getPoints = function(pp) {
	var ppt='';
	for(var i=0; i<pp.length; i++) {
		ppt += pp[i].x + ',' +pp[i].y+' ';
	}
	return ppt;
};

tafiffData = [];

var getTariffs = function() {
	var eventId = 1;
	$.ajax( "/api/eventtarif?event_id="+eventId )
		.done(function(resp,success) {
			if(success) {
				var data = resp.data;
				tafiffData = data;
			} else {}
		}).fail(function() {
			notify('error','Что-то пощло не так...');
		}).always(function(resp,success,tt) {

		});
}

var renew = function(e) {
	var zoneId = 'hh';
	var jqxhr = $.ajax( "/api/seatstates?event_id=1" )
	.done(function(resp,success) {
		if(success) {
			var svgRoot  = getSvgRoot();
			var data = resp.data;
			for(var i=0; i<data.length; i++) {
				var s = $("path#"+data[i].svg_id, svgRoot);
				if(s[0]) {
					if(data[i].state > 0) {
						s[0].classList.add('seatOcup');
					} else {
						s[0].classList.remove('seatOcup');
					}
				}
			}
		} else {}
	}).fail(function() {
			notify('error','Что-то пощло не так...');
	}).always(function(resp,success,tt) {
		
	});
};

var setSeatBuyed = function(seats) {
	//var seats = getSeatForPay();
	$.post( "/api/seatstates", { seat_ids: seats, state: 2 }) // 2 - buyed
	.done(function( resp,success ) {
		renew();
	});
};

cSvg = "crocusHoleAll";

var getSvgRoot = function() {
	var a = document.getElementById(cSvg);
	a.style.display = '';
	var svgDoc = a.contentDocument; //get the inner DOM of alpha.svg
	var svgRoot  = svgDoc.documentElement;
	return svgRoot;
};

var BuySeats = function(payData, seatsIds) {

	var widget = new cp.CloudPayments();
	widget.pay('charge', payData,
		{
			onSuccess: function (options) { // success
				//notify('info','Оплата успешно проведена');
			},
			onFail: function (reason, options) { // fail
				notify('error','Оплата не проведена: '+ reason);
			},
			onComplete: function (paymentResult, options) { //Вызывается как только виджет получает от api.cloudpayments ответ с результатом транзакции.
				notify('info','Оплата успешно проведена');
				setSeatBuyed(seatsIds);
			}
		}
	);
	
};

var getCustomerData = function() {
	var seats = getSeatForPay();
	if(!seats) {
		notify('warn','Места не выбраны');
		return;
	}
	var input = $("input#selectedSeatsData");
	input.attr('value', seats);
	$("a#nextFormActivate")[0].click();
}

var getDataFromForm = function(form) {
	var customData = {};
	// validate
	var fi = $("input", form);
	for(var i=0; i<fi.length; i++ ) {
		var inp = fi[i];
		if(inp.name == 'name') customData.name = inp.value.trim();;
		if(inp.name == 'email') customData.email = inp.value.trim();;
		if(inp.name == 'phone') customData.phone = customData.phone = inp.value.trim();;
		if(inp.name == 'seats') customData.seats = customData.seats = inp.value.trim();;
	}
	return customData;
}

var orderCreate = function(e) {
	var b = e.target;
	var form = $("form#userData")[0];
	var customData = getDataFromForm(form);
	var isValid = validateData(customData, form);
	if(!isValid) {
		return;
	}
	customData.event_id = 1;

	$.post( "/api/order", customData)
		.done(function( resp,success ) {
		if(success) {
			var data = resp.data;
			$("div#closeOrderForm")[0].click();
			BuySeats(data, customData.seats);
		} else {
			notify('error', resp.msg);
		}
	});
}

var sss = function(e) {
	var mapID = e.target.id ? e.target.id : e.target.parentElement.id;
	var isLeft = false;

	if(mapID.includes('zone')) {
		$("div#bigMap").animate({width: 0, height: 0}, 500, function () {});
		if (mapID == 'zone1') { cSvg = 'crocusHoleZone1'; isLeft = false; }
		if (mapID == 'zone2') { cSvg = 'crocusHoleZone2'; isLeft = true; }
		if (mapID == 'zone3') { cSvg = 'crocusHoleZone3'; isLeft = false; }
		if (mapID == 'zone4') { cSvg = 'crocusHoleZone4'; isLeft = true; }
		if (mapID == 'zone5') { cSvg = 'crocusHoleZone5'; isLeft = false; }
		if (mapID == 'zone6') { cSvg = 'crocusHoleZone6'; isLeft = true; }
		if (mapID == 'zone7') { cSvg = 'crocusHoleZone7'; isLeft = false; }
		if (mapID == 'zone8') { cSvg = 'crocusHoleZone8'; isLeft = true; }
		if (mapID == 'zone9') { cSvg = 'crocusHoleZone9'; isLeft = false; }
		if (mapID == 'zone10') { cSvg = 'crocusHoleZone10'; isLeft = true; }
		if (mapID == 'zone11') { cSvg = 'crocusHoleZone11'; isLeft = false; }
		if (mapID == 'zone12') { cSvg = 'crocusHoleZone12'; isLeft = true; }
		if (mapID == 'zone13') { cSvg = 'crocusHoleZone13'; isLeft = true; }
		if (mapID == 'zone14') { cSvg = 'crocusHoleZone14'; isLeft = true; }
		if (mapID == 'zone15') { cSvg = 'crocusHoleZone15'; isLeft = false; }
		$("div#" + mapID + "Map").animate({width: '100%', height: '100%'}, 1000, function () {
			defineEvent();

			if (isLeft) {
				$("div#exitBackButton")[0].style.left='10px';
				$("div#exitBackButton")[0].style.right='';
				$("div#buyButton")[0].style.right='10px';
				$("div#buyButton")[0].style.left='';
			} else {
				$("div#exitBackButton")[0].style.left='';
				$("div#exitBackButton")[0].style.right='10px';
				$("div#buyButton")[0].style.right='';
				$("div#buyButton")[0].style.left='10px'
			}
			$("div#exitBackButton")[0].style.display='block';
			$("div#buyButton")[0].style.display='block';
		});
	}
	if(mapID=='exitBackButton') {
		var z = cSvg.substr(14);
		$("div#exitBackButton")[0].style.display='none';
		$("div#buyButton")[0].style.display='none';
		$("div#zone" + z + "Map").animate({width: 0, height: 0}, 500, function () { });
		$("div#bigMap").animate({width: '100%', height: '100%'}, 1000, function () {
			defineEvent();
		});
	}

}

// RRR = 1;
// ZZZ = 1;
// MMM = 1;
var defineEvent = function() {
	var svgRoot  = getSvgRoot();
		if(cSvg == 'crocusHoleAll') {
			$("div#exitBackButton")[0].style.display='none';
			$("div#buyButton")[0].style.display='none';
			for(var i=1; i<16; i++)
				$("path#zone"+i,svgRoot).mouseover(opSwOff);
			for(var i=1; i<16; i++)
				$("path#zone"+i,svgRoot).mouseout(opSwOn);
			for(var i=1; i<16; i++)
				$("path#zone"+i,svgRoot).click(sss);
		}
		if(cSvg.includes('crocusHoleZone')  ) {
			$("path",svgRoot).mouseover(opSw1);
			$("path",svgRoot).mouseout(opSw2);
			$("path",svgRoot).click(selectSeat);
			// $("path",svgRoot).click( function(e) {
			// 	var s = e.target;
			// 	console.log('id="e1z'+ZZZ+'r'+RRR+'m'+MMM+'"-'+s.attributes.d.value);
			// 	MMM++;
			// });
			renew();
		}
}

$(document).ready(function(){
	var a = document.getElementById(cSvg);
	a.addEventListener("load",defineEvent,false);

	$("div#exitBackButton").click(sss);
	$("div#buyButton").click(getCustomerData);
	$("div#orderButton").click(orderCreate);
	// window.addEventListener("resize", resize);
	getTariffs();
});

var ws = new WebSocket('ws://localhost:8081/ws');
ws.onerror = function(event) {
	notify('error','WS соединение с сервером не установленно.');
	console.log('ws: Error');
};

ws.onopen = function(e) {
	e.target.send("WS: connection success/");
};

ws.onmessage = function(ws) {
	var msg = ws.data;
	console.log('ws:' + msg);
	var cmd = jQuery.parseJSON(msg);
	if(cmd) {
		if(cmd.action == 'seatStateUpdate') {
			updateSeatState(cmd.data);
		}
	}
};

var updateSeatState = function(seats) {
	var svgRoot  = getSvgRoot();
	for(var i=0; i<seats.length-1; i++) {
		var s = $("path#"+seats[i],svgRoot);
		if(s[0]) {
			s[0].classList.remove('seatSel');
			s[0].style.fillOpacity = 0.9;
			s[0].classList.add('seatOcup');
		}
	}
};

var notify = function(type, msg) {
	var nn = Math.floor(Math.random() * 1000);
	var attr = {'class': 'notifyBox', text: msg, 'id':'notif'+nn};
	if(type) {
		if(type=='info')
			attr.style = 'border-color:#00ff00';
		if(type=='warn')
			attr.style = 'border-color:#ff6600';
		if(type=='error')
			attr.style = 'border-color:#ff0000';
	}
	$("body").append( $('<div>', attr) );

	var box = $("div#notif"+nn);

	box.animate({top: '20px' }, 500, function () {
		setTimeout(function() {
			box.animate({right: '-350px' }, 1000, function () {
				box.remove();
			});
		}, 2000);
	});
}

var validateData = function(cData, form) {

	var isValid = true;
	var name = cData.name;
	var email = cData.email;
	var phone = cData.phone;

	if(!validateName(name)) {
		var field = $("input#customDataName", form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Введите фамилию и Имя через пробел';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input#customDataName", form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'none';
		field.classList.remove('inputError');
	}
	if(!validateEmail(email)) {
		var field = $("input#customDataEmail", form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Введен не правильный Email адрес.';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input#customDataEmail", form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'none';
		field.classList.remove('inputError');
	}
	if(!validatePhone(phone) || !validatePhoneLen(phone)) {
		var field = $("input#customDataPhone", form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Введен не правильный номер телефона.';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input#customDataPhone", form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'none';
		field.classList.remove('inputError');
	}

    return isValid;
}

const validateEmail = (email) => {
	return email.match(
		/^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
	);
};

const validatePhone = (phone) => {
	return phone.match(
		/^[+7]*[(]{0,1}[0-9]{1,3}[)]{0,1}[-\s\./0-9]*$/g
	);
};

const validateName = (name) => {
	return name.match(
		/[A-Za-zА-Яа-яЁё]+(\s+[A-Za-zА-Яа-яЁё]+)/g
	);
};

const validatePhoneLen = (phone) => {
	if(phone.startsWith('+7') || phone.startsWith('8')) {
		return phone.match(/\d/g).length===11
	} else
		return phone.match(/\d/g).length===10;
};