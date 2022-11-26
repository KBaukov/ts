package ent

import "time"

type AuthorizationExpression struct {
	NAME        string `json:"id"`
	DESCRIPTION string `json:"description"`
	TARGET      string `json:"target"`
	CONDITION   string `json:"condition"`
	ACCESS      string `json:"access"`
	MSG         string `json:"msg"`
}

type User struct {
	ID          int       `json:"id"`
	LOGIN       string    `json:"login"`
	PASS        string    `json:"pass"`
	ACTIVE_FLAG string    `json:"active_flag"`
	USER_TYPE   string    `json:"user_type"`
	LAST_VISIT  time.Time `json:"last_visit"`
}

type Seat struct {
	SEAT_ID     int         `json:"seat_id"`
	TARIFF_ID   int         `json:"tarif_id"`
	SEAT        interface{} `json:"seat"`
	SVG_ID      string      `json:"svg_id"`
	ACTIVE_FLAG int         `json:"active_flag"`
	STATE       int         `json:"state"`
}

type SeatInfo struct {
	SVG_ID      string `json:"svg_id"`
	TARIFF_NAME string `json:"tarif_name"`
	ZONE        string `json:"zone"`
	ROW_NUMBER  int    `json:"row_number"`
	SEAT_NUMBER int    `json:"seat_number"`
	PRICE       int    `json:"price"`
}

type SeatState struct {
	SEAT_ID int    `json:"seat_id"`
	SVG_ID  string `json:"svg_id"`
	STATE   int    `json:"state"`
}

type Tafif struct {
	TARIF_ID       int    `json:"tarif_id"`
	EVENT_ID       int    `json:"event_id"`
	PRICE          int    `json:"price"`
	PRICE_CURRENCY string `json:"price_currency"`
	TARIF_NAME     string `json:"tarif_name"`
	TARIF_ZONE_ID  int    `json:"tarif_zone_id"`
	COUNT          int    `json:"count"`
}

type Order struct {
	ORDER_ID      int         `json:"order_id"`
	ORDER_NUMBER  string      `json:"order_number"`
	ORDER_STATUS  int         `json:"order_status"`
	ORDER_LOG     interface{} `json:"order_log"`
	CUSTOMER_DATA interface{} `json:"customer_data"`
	ORDER_AMOUNT  int         `json:"amount"`
	ORDER_CURR    string      `json:"currency"`
}

type Reserv struct {
	RESERV_ID   int       `json:"reserv_id"`
	CREATE_TIME time.Time `json:"create_time"`
	SEAT_ID     int       `json:"seat_id"`
}

type ReservSeatMsg struct {
	ACTION  string `json:"action"`
	SEAT_ID string `json:"seat_id"`
}

type ApiError struct {
	ERROR_CODE   int    `json:"errorCode"`
	EROR_MESSAGE string `json:"errorMEssage"`
}

type ApiResp struct {
	SUCCESS bool        `json:"success"`
	DATA    interface{} `json:"data"`
	MSG     string      `json:"msg"`
}

type PayData struct {
	PUBLIC_ID   string      `json:"publicId"`
	DESCRIPTION string      `json:"description"`
	AMOUNT      int         `json:"amount"`
	CURR        string      `json:"currency"`
	ACCOUNT_ID  string      `json:"accountId"`
	INVOCE_ID   string      `json:"invoiceId"`
	EMAIL       string      `json:"email"`
	SKIN        string      `json:"skin"`
	AUTO_CLOSE  int         `json:"autoClose"`
	DATA        interface{} `json:"data"`
}

type PayDataExt struct {
	NAME  string `json:"name"`
	EMAIL string `json:"email"`
	PHONE string `json:"phone"`
	SEATS string `json:"seats"`
}

type ActionLog struct {
	LOG_TIME   time.Time `json:"log_time"`
	LOG_ACTION string    `json:"action"`
	ACTOR      string    `json:"actor"`
	PARAMS     string    `json:"params"`
}

// 			{ //options
// 				publicId: 'test_api_00000000000000000000002', //id из личного кабинета
// 				description: 'Fortune 2050 \nОплата билетов', //назначение
// 				amount: amount, //сумма
// 				currency: 'RUB', //валюта
// 				accountId: email, //идентификатор плательщика (необязательно)
// 				invoiceId: '1234567', //номер заказа  (необязательно)
// 				email: email, //email плательщика (необязательно)
// 				skin: "mini", //дизайн виджета (необязательно)
// 				autoClose: 3, //время в секундах до авто-закрытия виджета (необязательный)
// 				data: { seats: data }
// 			},

type ResendMessage struct {
	ACTION    string `json"action"`
	RECIPIENT string `json"recipient"`
	SENDER    string `json"sender"`
	MESSAGE   string `json:"msg"`
}

type WsSendData struct {
	ACTION string      `json"action"`
	TYPE   string      `json"type"`
	DATA   interface{} `json:"data"`
}
