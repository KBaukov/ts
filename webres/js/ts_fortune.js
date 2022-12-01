var addSeatForPay = function(node) {
	var sklad = $("div#tForPay");
	if(sklad[0])
		sklad[0].classList.add(node);
	else
		return null;
};

var emptyStore = function() {
	var sklad = $("div#tForPay");
	if(sklad[0])
		sklad[0].classList.value='';
	else
		return null;
}

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
	var s = e.target;
	var box = $("div#infoBox");
	if(!s.classList.contains('seatOcup') && !s.classList.contains('seatSel')  && !s.classList.contains('seatBlock')) {
		//var box = $("div#infoBox");
		//var tt = resolveTarif(s.id);
		if(box) {
			$("span#tarName",box).html(s.getAttribute('tn'));
			$("span#tarPrice",box).html(s.getAttribute('tp'));
		}
		s.style.fillOpacity = 0.6;
		box[0].style.display='block';
	}
};
var opSw2 = function(e) {
	var s = e.target;
	if(!s.classList.contains('seatOcup') && !s.classList.contains('seatSel')  && !s.classList.contains('seatBlock') ) {
		s.style.fillOpacity = 0.1;
		$("div#infoBox")[0].style.display = 'none';
	}
};
var opSwOn = function(e) { e.target.style.fillOpacity = 0.3;};
var opSwOff = function(e) { e.target.style.fillOpacity = 0.75;};
//MMXX=3;
var selectSeat = function(e) {
	var s = e.target;

	if(s.classList.contains('seatBlock')) {
		notify('warn','Извините это место заблокировано для продажи.');
		return false;
	}
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
		// s.classList.add('seatSel');
		// s.style.fillOpacity = 0.6;
		// addSeatForPay(s.id);
		// //console.log(s.id);
	}
	//console.log('z2r'+MMXX+s.id+'$'+s.attributes.d.value);

	return false;
};

// var getPoints = function(pp) {
// 	var ppt='';
// 	for(var i=0; i<pp.length; i++) {
// 		ppt += pp[i].x + ',' +pp[i].y+' ';
// 	}
// 	return ppt;
// };

//
// var resolveTarif = function(seat) {
// 	// var zz = seat.substring(2,4);
// 	// var rr = eval('tt.'+zz);
// 	// var mm = seat.indexOf('m');
// 	// var dd = seat.substring(4,mm);
// 	// var t = eval('rr.'+dd);
// 	var rr = {}
// 	for(var i=0; i<tafiffData.length; i++) {
// 		var m = tafiffData[i];
// 		if(m.seat_id == seat) {
// 			rr.t = m.t_name;
// 			rr.p = m.t_price;
// 			return rr;
// 		}
// 	}
// 	return rr;
// }

var tafiffData = [];

var UpdateSeatsTarifValues = function() {
	var svgRoot = getSvgRoot();
	for(var i=0; i<tafiffData.length; i++) {
		var m = tafiffData[i];
		var id = m.seat_id;
		var s = $("path#"+id, svgRoot)[0];
		if(s) {
			s.setAttribute('tp', m.t_price);
			s.setAttribute('tn', m.t_name);
		}
	}
}

var getTariffs = function() {
	var eventId = 1;
	$.ajax( "/api/seattarifs?event_id="+eventId )
		.done(function(resp,success) {
			if(success) {
				tafiffData = resp.data;
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
					if(s[0].classList.contains(data[i].svg_id)) {
						return;
					}

					if(data[i].state > 0) {
						if(data[i].state ==10) {
							s[0].classList.add('seatBlock');
						} else
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

var cSvg = "crocusHoleAll";

var getSvgRoot = function() {
	var a = document.getElementById(cSvg);
	a.style.display = '';
	var svgDoc = a.contentDocument; //get the inner DOM of alpha.svg
	var svgRoot  = svgDoc.documentElement;
	return svgRoot;
};

var orderLog = function(stage, seats, order_num, paymentResult, reson) {
	$.post( "/api/payaction", { stage: stage, seat_ids: seats, order_number: order_num, code:paymentResult.code,
		message:paymentResult.message, succes: (paymentResult.success ? 'true':'false' ), reason:reson  })
		.done(function( resp,success ) {
			renew();
		});
};

var BuySeats = function(payData, seatsIds) {
	var order_num = payData.invoiceId;;
	var widget = new cp.CloudPayments();
	widget.pay('charge', payData, {
		onSuccess: function (options) { // success
			var paymentResult = {code: "0", message:'Платеж отправлен', success: true};
			orderLog('paySuccess',seatsIds, order_num, paymentResult, '');
			emptyStore();
			window.location = '/paysuccess';
		},
		onFail: function (reason, options) { // fail
			notify('error','Оплата не проведена: ');
			var paymentResult = {code: "-1", message:'Платеж не отправлен', success: false};
			orderLog('payFail', seatsIds, order_num, 'Платеж не отправлен', reason);
		},
		onComplete: function (paymentResult, options) { //Вызывается как только виджет получает от api.cloudpayments ответ с результатом транзакции.
			//console.log('payResult:' + paymentResult);
			//notify('info','По заказу '+order_num+' - оплата успешно проведена');
			paymentResult.message =+ ': Платеж подтвержден';
			orderLog('payComplete', seatsIds,order_num, paymentResult, '');

		}
	});
	
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
		if(inp.name == 'name') customData.name = inp.value.trim();
		if(inp.name == 'email') customData.email = inp.value.trim();
		if(inp.name == 'phone') customData.phone = inp.value.trim();
		if(inp.name == 'seats') customData.seats = inp.value.trim();
		if(inp.name == 'checkbox') customData.checkbox = inp.value.trim();
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

	var seats = getSeatForPay();
	customData.event_id = 1;
	customData.seats = seats.toString();

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
			UpdateSeatsTarifValues();

			if (isLeft) {
				$("div#infoBox")[0].style.left='';
				$("div#infoBox")[0].style.right='50px';
				// $("div#buyButton")[0].style.right='0px';
				// $("div#buyButton")[0].style.left='';
			} else {
				$("div#infoBox")[0].style.left='50px';
				$("div#infoBox")[0].style.right='5';
				// $("div#buyButton")[0].style.left='';
				// $("div#buyButton")[0].style.righr='0px'
			}
			$("div#exitBackButton")[0].style.display='block';
			$("div#buyButton")[0].style.display='block';
			// $("div#infoBox")[0].style.display='block';

		});
	}
	if(mapID=='exitBackButton') {
		var z = cSvg.substr(14);
		$("div#exitBackButton")[0].style.display='none';
		$("div#buyButton")[0].style.display='none';
		$("div#infoBox")[0].style.display='none';
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
		//getTariffs();
		$("path",svgRoot).mouseover(opSw1);
		$("path",svgRoot).mouseout(opSw2);
		$("path",svgRoot).click(selectSeat);
		renew();
	}
	resize();
}

$(document).ready(function(){
	var a = document.getElementById(cSvg);
	a.addEventListener("load",defineEvent,false);

	$("div#exitBackButton").click(sss);
	$("div#buyButton").click(getCustomerData);
	$("div#orderButton").click(orderCreate);
	$("div#exitBackButton")[0].style.display='none';
	$("div#buyButton")[0].style.display='none';
	$("div#infoBox")[0].style.display='none';

	window.addEventListener("resize", resize);
	resize();
	getTariffs();
});

var resize = function(e) {
  var h = window.innerHeight-280;
  $("div#tForPay")[0].style.height = h+'px';
}

var ws = new WebSocket('ws://localhost:8443/ws');
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
					 s[0].classList.add('seatSel');
					 s[0].style.fillOpacity = 0.6;
					 addSeatForPay(s[0].id);
					 //console.log(s.id);
				}
			}
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
	//var apr1 = cData.checkbox;

	if(!validateName(name)) {
		var field = $("input[name='name']",form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Введите фамилию и Имя через пробел';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input[name='name']",form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'none';
		field.classList.remove('inputError');
	}
	if(!validateEmail(email)) {
		var field = $("input[name='email']",form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Введен не правильный Email адрес.';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input[name='email']",form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'none';
		field.classList.remove('inputError');
	}
	if(!validatePhone(phone) || !validatePhoneLen(phone)) {
		var field = $("input[name='phone']",form)[0];
		var errBox = field.parentElement.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Введен не правильный номер телефона.';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input[name='phone']",form)[0];
		var errBox = field.parentElement.nextElementSibling;
		errBox.style.display = 'none';
		field.classList.remove('inputError');
	}
	// if(!apr1) {
	// 	var field = $("input[name='phone']",form)[0];
	// 	var errBox = field.parentElement.nextElementSibling;
	// 	errBox.style.display = 'block';
	// 	errBox.textContent = 'Введен не правильный номер телефона.';
	// 	field.classList.add('inputError');
	// 	isValid = false;
	// } else {
	// 	var field = $("input[name='checkbox']",form)[0];
	// 	var errBox = field.parentElement.nextElementSibling;
	// 	errBox.style.display = 'none';
	// 	field.classList.remove('inputError');
	// }

    return isValid;
}

const validateEmail = (email) => {
	return email.match(
		/^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
	);
};

const validatePhone = (phone) => {
	phone = phone.replaceAll(' ', '');
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