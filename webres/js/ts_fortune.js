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
		tr.append($('<td>', { text: sector}));
		tr.append($('<td>', {text: formatSeat(id,1,null)}));
		tr.append($('<td>', {text: formatSeat(id,null, 1)}));
		var pr = node.getAttribute('tp');
		var fpr = number_format(pr,0, '.', ' ');
		tr.append($('<td>', { text: fpr+' ₽' }));
		tr.append($('<td>', { }));
		dd = $('td', tab);
		n = dd.length;
		var td =$(dd[n-1]);
		var aa = $('<a>', { class: 'delete-row', ss: id });
		 td.append(aa);  //class: 'infoDataSeat',

		var pr = parseInt(node.getAttribute('tp'));
		totalAmount += pr;
		var dd = sessionStorage.getItem('discountData');
		var dData = JSON.parse(dd);
		if(dData.amount >0 && isDiscShow) {
			if(dData.discount_type=='procent') {
				discountAmount = Math.round(totalAmount * dData.amount/100);
			} else {
				discountAmount = totalAmount - dData.amount;
			}

		} else
			discountAmount = 0;
		if(discountAmount>0) {
			$("span#totalAmount")[0].innerHTML = '<span class="old-price">'+
				number_format(totalAmount,0, '.', ' ')+' ₽</span>'
				+number_format((totalAmount-discountAmount),0, '.', ' ') + ' ₽';
		} else {
			$("span#totalAmount")[0].innerHTML = number_format(totalAmount,0, '.', ' ') +' ₽';
		}


		aa.click(removeRow);

	} else
		return null;
};

var totalAmount=0;


var formatSeat = function(id, isR, isM) {
	var tt = id.split('r');
	var ww = tt[1].split('m');
	if(isR) {
		return ww[0];
	}
	if(isM) {
		return ww[1];
	}
	return 'Ряд '+ww[0]+', Место '+ww[1];
};

var emptyStore = function() {
	var sklad = $("div#tForPay");
	if(sklad[0])
		sklad[0].classList.value='';
	else
		return null;
}

var removeRow = function(e) {
	var ss = e.target.getAttribute('ss');
	var node = $("path#"+ss,getSvgRoot());
	if(node[0]) {
		selectSeat(node[0]);
	}

}

var removeSeatFromPay = function(node) {
	var id = node.id;
	var sklad = $("div#tForPay");
	if(sklad[0]) {
		sklad[0].classList.remove(id);
		var tab = $("table#selSeatsInfo");
		$("tr#tt_"+id).remove();

		var pr = parseInt(node.getAttribute('tp'));
		totalAmount -= pr;
		$("span#totalAmount")[0].innerHTML = totalAmount+' ₽';
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

var isHoverLock = false;

var opSw1 = function(e) {
	if(isHoverLock) return;
	var s = e.target;
	//var x = e.x; var y =e.y;

	if(!s.classList.contains('seatOcup') && !s.classList.contains('seatSel')  && !s.classList.contains('seatBlock')) {
		var box = $("div#infoBox");
		var svgBox = $("#"+cSvg)[0].getBoundingClientRect();
		if(box) {
			$("div#seatZoneName",box).html(sector); //s.getAttribute('tn'));
			$("div#seatName",box).html(formatSeat(e.target.id, null, null));
			$("div#seatPrice",box).html(s.getAttribute('tp')+' ₽');
			if(!isTooltipBlock) box[0].style.display = 'block';
			box[0].style.top = (e.originalEvent.clientY+svgBox.top-120)+'px';
			box[0].style.left = (e.originalEvent.clientX+svgBox.left-70)+'px';
		}
		s.style.fillOpacity = 0.6;
		//box[0].style.display='block';
	}
};
var opSw2 = function(e) {
	if(isHoverLock) return;
	var s = e.target;
	var box = $("div#infoBox");
	box[0].style.display = 'none';
	if(!s.classList.contains('seatOcup') && !s.classList.contains('seatSel')  && !s.classList.contains('seatBlock') ) {
		s.style.fillOpacity = 0.1;
		$("div#seatZoneName",box).html('');
		$("div#seatPrice",box).html('');
	}
};
var opSwMove = function(e) {
	var s = e.target;
	if(!s.classList.contains('seatOcup') && !s.classList.contains('seatSel')  && !s.classList.contains('seatBlock') ) {
		var box = $("div#infoBox");
		var svgBox = $("#"+cSvg)[0].getBoundingClientRect();
		var zBox = $("#tForPay")[0].getBoundingClientRect();
		if(box) {
			if(!isTooltipBlock) box[0].style.display = 'block';
			box[0].style.top = (e.originalEvent.clientY+svgBox.top-120)+'px';
			box[0].style.left = (e.originalEvent.clientX+svgBox.left-70)+'px';
		}
	}
};
var opSwOn = function(e) {
	e.target.style.fillOpacity = 0.3;
	var box = $("div#zoneInfoBox");
	if(box) {
		box[0].style.display = 'none';
	}

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
	var box = $("div#zoneInfoBox");
	var svgBox = $("#"+cSvg)[0].getBoundingClientRect();
	if(box) {
		if(!isTooltipBlock) box[0].style.display = 'block';
		box[0].style.top = (e.originalEvent.clientY+svgBox.top-120)+'px';
		box[0].style.left = (e.originalEvent.clientX+svgBox.left-70)+'px';
	}
};

var opSWmove = function(e) {
	var box = $("div#zoneInfoBox");
	var svgBox = $("#"+cSvg)[0].getBoundingClientRect();
	if(box) {
		if(!isTooltipBlock)  box[0].style.display = 'block';
		box[0].style.top = (e.originalEvent.clientY+svgBox.top-120)+'px';
		box[0].style.left = (e.originalEvent.clientX+svgBox.left-70)+'px';
	}
};

// function oMousePos(svg, evt) {
// 	var ClientRect = svg.getBoundingClientRect();
// 	return { //objeto
// 		x: Math.round(evt.clientX - ClientRect.left),
// 		y: Math.round(evt.clientY - ClientRect.top)
// 	}
// };
//
// function getCursorPosition(event, svgElement) {
// 	//var svg = getSvgRoot();
// 	var svgPoint = svgElement.createSVGPoint();
// 	svgPoint.x = event.clientX;
// 	svgPoint.y = event.clientY;
// 	return svgPoint.matrixTransform(svgElement.getScreenCTM().inverse());
// };

var selectSeat = function(e) {
	var s;

	if(e.type) {
		s = e.target;
	} else {
		s = e;
	}

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
		removeSeatFromPay(s);
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
					//$("div#infoBox")[0].style.left = "-200px;"
					//$("div#infoBox").show();

				}
			} else {}
		}).fail(function() {
			notify('error','Что-то пощло не так...');
		}).always(function(resp,success,tt) {

		});
};

var orderLog = function(stage, seats, order_num, paymentResult, reson) {
	var code = paymentResult.code;
	var sesData = JSON.parse(sessionStorage.getItem('cData'));
	var leadId =  sesData.amoData.lead_id;
	var contactId =  sesData.amoData.contact_id;

	$.post( "/api/order/log", { stage: stage, seat_ids: seats, order_number: order_num, code:paymentResult.code,
		message:paymentResult.message, succes: (paymentResult.success ? 'true':'false' ), reason:reson , lead_id: leadId, contact_id: contactId })
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
			var sesData = JSON.parse(sessionStorage.getItem('cData'))
			window.location = 'https://fortune2050.com/success/'+window.location.search;
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
	if(!seats || seats.length==0) {
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
		if(inp.name == 'phone') customData.phone = $("input#customDataPhone").intlTelInput("getNumber");
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
	//if() {
		$("div#shadow").show();
		$("div#shadow").animate({ opacity: 1}, 500, function() { });
	//}

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
		customData.utm = sessionStorage.utm;
		customData.ref_id = refId;
		customData.link = link;
		customData.contact_id = sesData.amoData.contact_id;
		customData.lead_id = sesData.amoData.lead_id;
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
		customData.utm = sessionStorage.utm;
		customData.ref_id = refId;
		customData.link = link;
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

var sector = '';
var isLeft = false;
var isDiscShow = true;

var sss = function(e) {
	var mapID = e.target.id ? e.target.id : e.target.parentElement.id;
	isLeft = false;
	if(mapID.includes('zone')) {
		$("div#bigMap").animate({width: 0, height: 0}, 500, function () {});
		if (mapID == 'zone1') { cSvg = 'crocusHoleZone1'; isLeft = false; sector='VIP-ПАРТЕР'; isDiscShow = false;}
		if (mapID == 'zone2') { cSvg = 'crocusHoleZone2'; isLeft = true;  sector='VIP-ПАРТЕР'; isDiscShow = false;}
		if (mapID == 'zone3') { cSvg = 'crocusHoleZone3'; isLeft = false; sector='VIP-ПАРТЕР';}
		if (mapID == 'zone4') { cSvg = 'crocusHoleZone4'; isLeft = true; sector='VIP-ПАРТЕР';}
		if (mapID == 'zone5') { cSvg = 'crocusHoleZone5'; isLeft = false; sector='ПАРТЕР';}
		if (mapID == 'zone6') { cSvg = 'crocusHoleZone6'; isLeft = true; sector='ПАРТЕР';}
		if (mapID == 'zone7') { cSvg = 'crocusHoleZone7'; isLeft = false; sector='ПАРТЕР';}
		if (mapID == 'zone8') { cSvg = 'crocusHoleZone8'; isLeft = true; sector='ПАРТЕР';}
		if (mapID == 'zone9') { cSvg = 'crocusHoleZone9'; isLeft = false; sector='АМФИТЕАТР';}
		if (mapID == 'zone10') { cSvg = 'crocusHoleZone10'; isLeft = true; sector='АМФИТЕАТР';}
		if (mapID == 'zone11') { cSvg = 'crocusHoleZone11'; isLeft = true; sector='БЕЛЬЭТАЖ'; isDiscShow = false;}
		if (mapID == 'zone12') { cSvg = 'crocusHoleZone12'; isLeft = false; sector='БЕЛЬЭТАЖ'; isDiscShow = false;}
		if (mapID == 'zone13') { cSvg = 'crocusHoleZone13'; isLeft = true; sector='БЕЛЬЭТАЖ'; }
		if (mapID == 'zone14') { cSvg = 'crocusHoleZone14'; isLeft = true; sector='БЕЛЬЭТАЖ'; isDiscShow = false;}
		if (mapID == 'zone15') { cSvg = 'crocusHoleZone15'; isLeft = false; sector='БЕЛЬЭТАЖ'; isDiscShow = false;}
		isHoverLock=true;
		$("div#zoneInfoBox").hide();
		$("div#" + mapID + "Map").animate({width: '100%', height: '100%'}, 1000, function () { //, marginLeft: -120
			isHoverLock=false;
			UpdateSeatsTarifValues();

			// var cc = $("#"+cSvg)[0].getBoundingClientRect();
			// if (isLeft) {
			// 	$("div#bannerBox")[0].style.left='';
			// 	$("div#bannerBox")[0].style.right=  (window.innerWidth-cc.width-cc.x)+'px';
			//
			// } else {
			// 	$("div#bannerBox")[0].style.left= (cc.x+30)+'px';
			// 	$("div#bannerBox")[0].style.right=''
			//
			// }


			$("div#toolBar2").show();
			$("div#toolBar1").hide();
			$("a#exitBackButton").show();
			$("div#totalBlock").show();
			// $("div#infoBoxSeat").show();

			// $("div#bannerBox").show();




			renew();
		});
		$("div#infoBoxSeat").animate( { opacity: 1, height: 281 }, 500,function(ee) {
			this.style.height = 'inherit';

		} );
		$("div#banerSector")[0].innerHTML= sector;
		var dd = sessionStorage.getItem('discountData');
		var dData = JSON.parse(dd);
		if(dData.amount && dData.amount>0 && isDiscShow) {
			$("div#discontBanner").show();
			$("div#discontBanner")[0].innerHTML= 'вам скидка '+discountVal;
		}
		$("div#bannerBox").animate( { opacity: 1, height: 87 }, 500, function() {} );
	}
	if(mapID=='exitBackButton') {
		var cont = $("div#zalContainer")[0].getBoundingClientRect();
		var z = cSvg.substr(14);
		$("div#toolBar2").hide();
		$("div#totalBlock").hide();
		$("div#toolBar1").show();
		$("div.infoBox").hide();
		// $("div#infoBoxSeat").hide();
		$("div#infoBoxSeat").animate( { opacity: 0, height: 0 }, 500, function() {} );
		$("div#bannerBox").animate( { opacity: 0, height: 0 }, 500, function() {} );
		// $("div#bannerBox").hide();
		$("div#zone" + z + "Map").animate({width: 0, height: 0}, 500, function () { });
		$("div#bigMap").animate({width: cont.width, height: cont.height}, 1000, function () {

		});
		cSvg = 'crocusHoleAll';
	}

}

var defZone = function(a) {
	var svgRoot  = a.target.contentDocument.documentElement;
	$("path.fil0",svgRoot).mouseover(opSw1);
	$("path.fil0",svgRoot).mouseout(opSw2);
	$("path.fil0",svgRoot).mousemove(opSwMove);
	$("path.fil0",svgRoot).click(selectSeat);
	//
	// document.querySelector("div#zalContainer").onmouseenter=infoBoxEnt;
	// document.querySelector("#crocusHoleZone1").onmousemove=infoBoxPos;
	// document.querySelector("div#zalContainer").onmouseleave=infoBoxLeav;


}

var defMAin = function(a) {
	var svgRoot  = a.target.contentDocument.documentElement;
	for(var i=1; i<16; i++) {
		$("path#zone" + i, svgRoot).mouseover(opSwOff);
		$("path#zone" + i, svgRoot).mouseout(opSwOn);
		$("path#zone" + i, svgRoot).mousemove(opSWmove);
		$("path#zone" + i, svgRoot).click(sss);
		if ($("path#zone" + i, svgRoot)[0])
			$("path#zone" + i, svgRoot)[0].style.cursor = 'pointer';
	};
	// svgRoot.style.cursor = 'move';
	// svgRoot.addEventListener('mousemove', mapDrag, {ddd: "fff"});
};

var svgX = 0;
var cX = 0;
var dX = 0;
var sX = 0;
var cSvgX = 0;
var mapDrag = function(e) {
	var svg = e.target;
	var butt = e.buttons;
	var cursor = {x: e.x }; //getCursorPosition(e, svg);
	//console.log(cursor);

	cX = cursor.x;
	if(butt) {
		dX = (sX - cX);
		var ss = svg.viewBox.baseVal;
		var max = ss.width;
		//if(dX <0) dX =0;
		//if(dX>max) dX =max;
		// cSvgX=dX
		var ww = cSvgX+dX;
		ss.x = ( ww > 1300 ? 1300 : ( ww<0 ? 0 : ww) );
	} else {
		sX = cursor.x;
		var ww = svg.viewBox.baseVal.x;
		cSvgX = ( ww > 1300 ? 1300 : ( ww<0 ? 0 : ww) );
	}

	//console.log('cX:'+cX+' sX:'+sX+' dX:'+dX+' svgX:'+cSvgX );
};

function getCursorPosition(event, svgElement) {
	var svgPoint = svgElement.createSVGPoint();
	svgPoint.x = event.clientX;
	svgPoint.y = event.clientY;
	return svgPoint.matrixTransform(svgElement.getScreenCTM().inverse());
};

var isDefine = [];
var cSvg = "crocusHoleAll";

$(document).ready(function(){

	sessionStorage.removeItem('utm');
	sessionStorage.removeItem('cData');
	sessionStorage.removeItem('discountData');

	$("div#shadow").click(formClose);
	$("div#formCloseButt").click(formClose);
	$("div#apruveChBox").click(krClick);
	$("a#exitBackButton").click(sss);
	$("a#buyButton").click(getCustomerData);
	$("div#orderButton").click(orderCreate);
	$("a#exitBackButton").hide();
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
	window.addEventListener("resize", resize);
	resize();
	getTariffs();
	getUtm();

	// var input = document.querySelector("#customDataPhone");
	//window.intlTelInput(input, {defaulteCountry: 'RU'});
	$("input#customDataPhone").intlTelInput( {defaultCountry: "RU", preferredCountries: [ "RU" ], initialCountry: "RU",utilsScript: "/js/utils.js", autoFormat: true, autoPlaceholder: true } );

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

var refId='';
var link='';
var discountVal='';

var checkDiscount = function() {
	$.get( "/api/getdiscount", { ref_id: refId})
		.done(function( resp,success ) {
			if(success) {
				var dData = resp.data;
				sessionStorage.setItem('discountData', JSON.stringify(dData));
				if(dData.discount_type=='procent') {
					discountVal=dData.amount + '%'
				} else {
					discountVal=dData.amount+' ₽'
				}
			} else {
				notify('error', resp.msg);
			}
		});
}

var getUtm = function() {
	link = window.location.href;
	var stt = window.location.search.substring(1);
	var utm = '';
	if(stt=='') {
		utm = '{}';
	} else {
		var tt = stt.split('&');
		tt.forEach(function(e) {
			var t = e.split('=');
			if(t[0]=='referrer' || t[0]=='utm_referrer') {
				refId = t[1];
			}
			utm += ',"'+t[0]+'":"'+t[1]+'"'
		});
		utm = '{'+utm.substring(1)+'}';
	}
	sessionStorage.setItem('utm', utm);

	checkDiscount();
}

var ws = new WebSocket('wss://'+document.location.host+'/ws');
ws.onerror = function(event) {
	notify('error','WS соединение с сервером не установленно.');
	console.log('ws: Error');
};

ws.onopen = function(e) {
	e.target.send("WS: connection success/");
};

isWsLog = false;

ws.onmessage = function(ws) {
	var svgRoot  = getSvgRoot();
	var msg = ws.data;
	if (isWsLog) console.log('ws:' + msg);
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

var isTooltipBlock = false;

var resize = function(e) {
	isTooltipBlock = (window.innerWidth<601);
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
	if(!validatePhone(phone) ) { //|| !validatePhoneLen(phone)
		var field = $("input[name='phone']",form)[0];
		var errBox = $("div#phoneErr")[0];
		errBox.style.display = 'block';
		errBox.textContent = 'Введен не правильный номер телефона.';
		field.classList.add('inputError');
		isValid = false;
	} else {
		var field = $("input[name='phone']",form)[0];
		var errBox = $("div#phoneErr")[0];
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
	// phone = phone.replaceAll(' ', '');
	// return phone.match(
	// 	/^[+7]*[(]{0,1}[0-9]{1,3}[)]{0,1}[-\s\./0-9]*$/g
	// );
	return $("input#customDataPhone").intlTelInput("isValidNumber");
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
};

function number_format( number, decimals, dec_point, thousands_sep ) {  // Format a number with grouped thousands
	var i, j, kw, kd, km;

	if( isNaN(decimals = Math.abs(decimals)) ){
		decimals = 2;
	}
	if( dec_point == undefined ){
		dec_point = ",";
	}
	if( thousands_sep == undefined ){
		thousands_sep = ".";
	}

	i = parseInt(number = (+number || 0).toFixed(decimals)) + "";

	if( (j = i.length) > 3 ){
		j = j % 3;
	} else{
		j = 0;
	}

	km = (j ? i.substr(0, j) + thousands_sep : "");
	kw = i.substr(j).replace(/(\d{3})(?=\d)/g, "$1" + thousands_sep);
	kd = (decimals ? dec_point + Math.abs(number - i).toFixed(decimals).replace(/-/, 0).slice(2) : "");

	return km + kw + kd;
};