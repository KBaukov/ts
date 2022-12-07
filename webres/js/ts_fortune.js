addSeatForPay = function(node) {
	var id = node.id;
	var sklad = $("div#tForPay");
	if(sklad[0]) {
		sklad[0].classList.add(id);

		var tab = $("table#selSeatsInfo");
		tab.append( $('<tr>', { id: 'tt_'+id}) );
		var dd = $('tr', tab);
		var n = dd.length;
		var tr = $(dd[n-1]);
		tr.append($('<td>', {}));
		dd = $('td', tab);
		n = dd.length;
		var td =$(dd[n-1]);
		td.append($('<div>', { class: 'infoLabelSeat', text: formatSeat(id)}));

		tr.append($('<td>', {}));
		dd = $('td', tab);
		n = dd.length;
		var td =$(dd[n-1]);
		td.append($('<div>', { class: 'infoDataSeat', text: node.getAttribute('tp')+' ₽'}));
	} else
		return null;
};

var formatSeat = function(id) {
	var tt = id.split('r');
	var ww = tt[1].split('m');
	return 'ряд '+ww[0]+' место '+ww[1];
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
	if(sklad[0]) {
		sklad[0].classList.remove(node);
		var tab = $("table#selSeatsInfo");
		$("tr#tt_"+node).remove();
	}

};

updateSeatState = function(seats) {
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

var getSeatForPay = function() {
	var sklad = $("div#tForPay");
	if(sklad[0] && sklad[0].classList.length>0)
		return sklad[0].classList;//.split(' ');
	else
		return [];
};

var opSw1 = function(e) {
	var s = e.target;
	var box = $("div#infoBox");
	if(!s.classList.contains('seatOcup') && !s.classList.contains('seatSel')  && !s.classList.contains('seatBlock')) {
		//var box = $("div#infoBox");
		//var tt = resolveTarif(s.id);
		if(box) {
			$("div#seatZoneName",box).html(s.getAttribute('tn'));
			$("div#seatPrice",box).html(s.getAttribute('tp')+' ₽');
		}
		s.style.fillOpacity = 0.6;
		//box[0].style.display='block';
	}
};
var opSw2 = function(e) {
	var s = e.target;
	var box = $("div#infoBox");
	if(!s.classList.contains('seatOcup') && !s.classList.contains('seatSel')  && !s.classList.contains('seatBlock') ) {
		s.style.fillOpacity = 0.1;
		$("div#seatZoneName",box).html('');
		$("div#seatPrice",box).html('');
		//$("div#infoBox")[0].style.display = 'none';
	}
};
var opSwOn = function(e) {
	e.target.style.fillOpacity = 0.3;
	$("div#tarName").html('Не выбран');
	$("div#tarPrice").html('');

};
var opSwOff = function(e) {
	var z = e.target;
	e.target.style.fillOpacity = 0.75;
	if(z.id) {
		var id = z.id;
		var zNum = id.substring(4);
		for(var i=0; i<zoneTarifs.length; i++) {
			var node = zoneTarifs[i];
			if (''+node.zone_num == zNum ) {
				$("div#tarName").html(node.zone_name);
				$("div#tarPrice").html('от '+node.min+' ₽');
			}
		}

	}

};

var selectSeat = function(e) {
	var s = e.target;

	if(s.classList.contains('seatBlock')) {
		notify('info','Извините это место заблокировано для продажи.');
		return false;
	}
	if(s.classList.contains('seatOcup')) {
		notify('info','Извините это место уже кем-то занято.');
		return false;
	}
	if(s.classList.contains('seatSel')) {
		ws.send('{"action":"unreserve", "seat_id":"'+s.id+'"}');
		s.classList.remove('seatSel');
		removeSeatFromPay(s.id);
	} else {
		ws.send('{"action":"reserve", "seat_id":"'+s.id+'"}');
	}
	//console.log('z2r'+MMXX+s.id+'$'+s.attributes.d.value);

	//return false;
};


var tafiffData = [];
var zoneTarifs = [];

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
		notify('error','Что-то пошло не так...');
	}).always(function(resp,success,tt) {

	});

	$.ajax( "/api/zonetarifs" )
		.done(function(resp,success) {
			if(success) {
				zoneTarifs = resp.data;
			} else {}
		}).fail(function() {
		notify('error','Что-то пощло не так...');
	}).always(function(resp,success,tt) {

	});
}

var renew = function(e) {
	var zoneId = 'hh';
	var sklad = getSeatForPay();
	$.ajax( "/api/seatstates?event_id=1" )
		.done(function(resp,success) {
			if(success) {
				var svgRoot  = getSvgRoot();
				var data = resp.data;
				for(var i=0; i<data.length; i++) {
					var s = $("path#"+data[i].svg_id, svgRoot);
					if(s[0]) {
						if(sklad.length=0 || sklad.includes(data[i].svg_id)) {
							continue;
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

var orderLog = function(stage, seats, order_num, paymentResult, reson) {
	var code = paymentResult.code;

	$.post( "/api/order/log", { stage: stage, seat_ids: seats, order_number: order_num, code:paymentResult.code,
		message:paymentResult.message, succes: (paymentResult.success ? 'true':'false' ), reason:reson  })
		.done(function( resp,success ) {
			if(code==200) {
				sessionStorage.removeItem('cData');
				emptyStore();
				renew();
			}

		});
};

var BuySeats = function(payData, seatsIds) {
	// var payData = norm(pData);
	var order_num = payData.invoiceId;
	var widget = new cp.CloudPayments();
	widget.pay('charge', payData, {
		onSuccess: function (options) { // success
			var paymentResult = {code: "200", message:'Платеж подтвержден', success: true};
			orderLog('paySuccess',seatsIds, order_num, paymentResult, '');
			window.location = '/paysuccess';
		},
		onFail: function (reason, options) { // fail
			notify('error','Оплата не проведена: '+reason);
			var paymentResult = {code: "-1", message:'Платеж не отправлен', success: false};
			orderLog('payFail', seatsIds, order_num, paymentResult, reason);
		},
		onComplete: function (paymentResult, options) {
			//paymentResult.message = 'Платеж совершен но не подтвержден';
			orderLog('payComplete', seatsIds, order_num, paymentResult, {});
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
	formOpen();
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
		if(inp.name == 'checkbox') customData.checkbox = inp.checked;
	}
	return customData;
}
var formClose = function(e) {
	if( e ) {
		if(e=='ts' || e.target.id == "shadow" || e.target.id =='formCloseButt') {
			$("div#shadow").animate({opacity: .01 }, 500, function() {
				$("div#shadow").hide();
			});
		}
	}

}

var formOpen = function(isClear) {
	$("div#shadow").show();
	$("div#shadow").animate({ opacity: 1}, 500, function() { });
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
	var sesData = JSON.parse(sessionStorage.getItem('cData'));

	if( sesData ) {
		customData.order_number = sesData.invoiceId;
		$.get( "/api/order", customData)
		.done(function( resp,success ) {
			if(success) {
				var cData = resp.data;
				sessionStorage.setItem('cData', JSON.stringify(cData));
				formClose('ts');
				BuySeats(cData, customData.seats);
			} else {
				notify('error', resp.msg);
			}
		});
	} else {
		$.post( "/api/order", customData)
			.done(function( resp,success ) {
				if(success) {
					var cData = resp.data;
					sessionStorage.setItem('cData', JSON.stringify(cData));
					formClose('ts');
					BuySeats(cData, customData.seats);
				} else {
					notify('error', resp.msg);
				}
			});
	}


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
		$("div#" + mapID + "Map").animate({width: '100%', height: '100%', marginLeft: -120}, 1000, function () {
			UpdateSeatsTarifValues();

			// if (isLeft) {
			// 	$("div#infoBox")[0].style.left='';
			// 	$("div#infoBox")[0].style.right='60px';
			// 	$("div#infoBoxSeat")[1].style.left='';
			// 	$("div#infoBoxSeat")[1].style.right='300px';
			// 	// $("div#buyButton")[0].style.right='0px';
			// $("div#buyButton")[0].style.left='';
			// } else {
			// 	$("div#infoBox")[0].style.left='60px';
			// 	$("div#infoBox")[0].style.right='0';
			// 	$("div#infoBoxSeat")[1].style.left='300px';
			// 	$("div#infoBoxSeat")[1].style.right='0';
			// 	// $("div#buyButton")[0].style.left='';
			// 	// $("div#buyButton")[0].style.righr='0px'
			// }

			$("div#toolBar2").show();
			$("div#toolBar1").hide();
			$("div.infoBox").show();
			$("div#exitBackButton").show();
			$("div#buyButton").show();
			//$("div#buyButton").click(getCustomerData);
			renew();
		});
	}
	if(mapID=='exitBackButton') {
		var z = cSvg.substr(14);
		$("div#toolBar2").hide();
		$("div#buyButton").hide();
		$("div#toolBar1").show();
		$("div.infoBox").hide();
		$("div#zone" + z + "Map").animate({width: 0, height: 0}, 500, function () { });
		$("div#bigMap").animate({width: '100%', height: '100%'}, 1000, function () {
		});
	}

}

var defZone = function(a) {
	var svgRoot  = a.target.contentDocument.documentElement;
	$("path.fil0",svgRoot).mouseover(opSw1);
	$("path.fil0",svgRoot).mouseout(opSw2);
	$("path.fil0",svgRoot).click(selectSeat);
}

var defMAin = function(a) {
	var svgRoot  = a.target.contentDocument.documentElement;
	for(var i=1; i<16; i++) {
		$("path#zone" + i, svgRoot).mouseover(opSwOff);
		$("path#zone" + i, svgRoot).mouseout(opSwOn);
		$("path#zone" + i, svgRoot).click(sss);
		if ($("path#zone" + i, svgRoot)[0])
			$("path#zone" + i, svgRoot)[0].style.cursor = 'pointer';
	}
}

var isDefine = [];
var cSvg = "crocusHoleAll";

$(document).ready(function(){

	$("div#shadow").click(formClose);
	$("div#formCloseButt").click(formClose);
	$("div#apruveChBox").click(krClick);
	$("div#exitBackButton").click(sss);
	$("div#buyButton").click(getCustomerData);
	$("div#orderButton").click(orderCreate);
	$("div#exitBackButton").hide();
	//$("div#buyButton")[0].style.display='none';
	$("div.infoBox").hide();

	cSvg = "crocusHoleAll";
	var a = document.getElementById('crocusHoleAll');
	a.addEventListener("load",defMAin,false);
	//var svgDoc = a.contentDocument; //get the inner DOM of alpha.svg
	//var svgRoot  = svgDoc.documentElement;

	for(var i=1; i<16; i++) {
		var b = document.getElementById("crocusHoleZone"+i);
		if(b)
			b.addEventListener("load",defZone,false);
	}

	cSvg = "crocusHoleAll";

	//window.addEventListener("resize", resize);
	//resize();
	getTariffs();
});
var getSvgRoot = function() {
	var a = document.getElementById(cSvg);
	if(a) {
		var svgDoc = a.contentDocument; //get the inner DOM of alpha.svg
		var svgRoot  = svgDoc.documentElement;
		return svgRoot;
	}
	//a.addEventListener("load",defineEvent,false);
	return null;
};

var ws = new WebSocket('wss://'+document.location.host+'/ws');
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

var resize = function(e) {
	var h = window.innerHeight-280;
	var h = window.innerWidth-200;
	//$("div#tForPay")[0].style.height = h+'px';
}



var notify = function(type, msg) {
	var nn = Math.floor(Math.random() * 1000);
	var attr = { 'id':'notif'+nn};

	attr.class = 'notifyBox body'+type;
	//attr.text = '<div class="header"+type></div>'+'<div class="msg"+type>'+msg+'</div>';
	$("body").append( $('<div>', attr) );
	var box = $("div#notif"+nn);

	attr = { class: "header"+type, text: '!!!!!!!!'}
	box.append($('<div>', attr));

	attr = { class: "msg"+type, text: msg}
	box.append($('<div>', attr));

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
	var apr1 = cData.checkbox;

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
		// errBox.style.display = 'none';
		errBox.textContent = ' ';
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
		// errBox.style.display = 'none';
		errBox.textContent = ' ';
		field.classList.remove('inputError');
	}
	if(!validatePhone(phone) || !validatePhoneLen(phone)) {
		var field = $("input[name='phone']",form)[0];
		var errBox = field.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Введен не правильный номер телефона.';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input[name='phone']",form)[0];
		var errBox = field.nextElementSibling;
		// errBox.style.display = 'none';
		errBox.textContent = ' ';
		field.classList.remove('inputError');
	}
	if(!apr1) {
		var field = $("input[name='checkbox']",form)[0];
		var errBox = field.parentElement.nextElementSibling;
		errBox.style.display = 'block';
		errBox.textContent = 'Необходимо принять условия или завершить оформление заказа';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input[name='checkbox']",form)[0];
		var errBox = field.parentElement.nextElementSibling;
		// errBox.style.display = 'none';
		errBox.textContent = ' ';
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

var krClick = function(e) {
	//$("input.t-checkbox")[0].click();
	var tt = $("input.t-checkbox")[0];
	tt.checked= !tt.checked;
	return chboxClick(tt);
}

var chboxClick = function(inp) {

	var kr = $("div#apruveChBox")[0]
	if(kr) {
		if(inp.checked) {
			kr.style.backgroundPosition = 'right';
		} else {
			kr.style.backgroundPosition = 'left';
		}
	}
	return false;
}