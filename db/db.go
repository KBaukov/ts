package db

import (
	_ "encoding/json"
	"fmt"
	"github.com/KBaukov/ts/amo"
	"github.com/KBaukov/ts/config"
	"github.com/KBaukov/ts/ent"
	"github.com/KBaukov/ts/esend"
	"math"
	_ "math"
	_ "reflect"
	_ "sort"
	"strconv"
	"strings"
	_ "strings"
	"time"

	//"errors"
	"github.com/jmoiron/sqlx"
	"log"

	_ "github.com/lib/pq"
)

var (
	OfdVat      int
	OfdMathod   int
	OfdOobject  int
	TaxSyst     int
	amoTickSend string
)

const (
	getSeatsStatusesQuery = `select * from seats where tariff_id i= (select tariff_id from tariff where event_id= $1 and tariff_name=$2) order by seat_id;`

	getSeatStatesQuery = `select seat_id, svg_id, state from seats where tariff_id in (select tariff_id from tariff where event_id= $1 ) order by seat_id;`
	setSeatStateQuery  = `update seats set state = $1 where svg_id = $2;`
	checkSeaState      = `select state from seats where svg_id = $1;`

	getSeatsInfoQuery = `select s.svg_id, t.tariff_name, s.seat->>'zone' as zone, s.seat->>'row' as row_number, s.seat->>'place' as seat_number,  
t.price::numeric::int as price from seats s, tariff t  
where s.svg_id  = $1
and s.tariff_id =t.tariff_id;`

	addCrmDataInOrder = `update orders set customer_data = customer_data || ($1)::jsonb where order_number = $2;`

	getSeatInfoBySvgId = `select t.tariff_name as zone, seat->>'row' as row, seat->>'place' as seat from seats s, tariff t  
    where svg_id in ( select * from UNNEST(regexp_split_to_array($1, ' '))) and s.tariff_id =t.tariff_id;`

	getEventTarifsQuery = `select tariff_id, event_id, price::numeric::int as price, price_currency, tariff_name, tariff_zone_id, count 
							from tariff where event_id = $1 order by tariff_id;`

	reserveSeatQuery    = `insert into reserv (seat_id) values ($1);`
	unreserveSeatQuery  = `delete from reserv where seat_id = (select seat_id from seats where svg_id = $1);`
	checkSeatForReserve = `select seat_id from seats where svg_id = $1;`

	getLastOrderByEvent = `select order_number from orders where order_id = (select max(order_id) as order_id from orders where event_id = $1 );`
	createNewOrder      = `insert into orders (order_number, order_status, order_log, customer_data, event_id, amount, amount_currency) 
values ($1, $2, ( $3 )::jsonb, ( $4 )::jsonb, $5, $6, $7);`

	assignSeatToOrder = `update seats set order_id = $1 where svg_id in ( select unnest(regexp_split_to_array($2,' ')));`

	updateOrder = `update orders set order_log= order_log|| ($1)::jsonb, customer_data =$2, amount=$3 where order_number = $4;`

	orderLog = `update orders set order_log = order_log || ($1)::jsonb , order_status = $2 where order_number = $3;`

	getOrderIdByNumber = `select order_id from orders where order_number = $1;`

	getAmountBySeats = `select seat->>'zone' as zone, seat->>'row' as row, seat->>'place' as seat, 
t.price::numeric::float as price, t.price_currency  from seats s, tariff t  
where s.svg_id in (select unnest(string_to_array($1, ' ')) as dd)
and s.tariff_id = t.tariff_id;`

	getDiscountByRefId = `select * from discount_ref where ref_id = $1;`

	getSeatTarif = `select tt.price::numeric::int as p, tt.tariff_name as t, s.svg_id as id from seats s, tariff tt where tt.tariff_id =s.tariff_id and tt.event_id = $1;`

	getSeatsExpierReserv     = `select seat_id, svg_id from seats where state=1 and seat_id in ( select seat_id from reserv where create_time < ( now() - time '00:10') );`
	clearExpiredReserves     = `delete from reserv where create_time < ( now() - time '00:10');`
	clearStatusForUnreserved = `update seats set state = 0 where state=1 and seat_id not in ( select seat_id from reserv );`

	getTicketsForSend = `select o.order_status, o.order_log#>'{0}'->>'log_time' dt, o.order_number, o.customer_data->>'name' as name,
o.customer_data->>'email' as email, o.customer_data->>'phone' as pone, s.seat ->>'zone' as sector,  (s.seat ->>'row') as row, 
s.seat ->>'place' as seat, s.seat_id, o.customer_data ->'crm'->>'lead_id' as lead_id
from orders o, seats s CROSS JOIN UNNEST(regexp_split_to_array(o.customer_data->>'seats', ' ') ) as ss
where o.order_status in ('4','6') and s.svg_id = ss;`

	getLastTicketNumber     = `select ticket_number  from ticket where ticket_id = ( select min(ticket_id) from ticket t where state = 0 );`
	updateTicketStatus      = `update ticket set state = 1, seat_id = $1 where ticket_number = $2;`
	getTicketNumberBySeatId = `select ticket_number from ticket where seat_id = $1;`

	selectZoneMaxMin = `select zone_number::int, z_name, max(price), min(price) from (
	select ( regexp_split_to_array((regexp_split_to_array(s.svg_id, 'r'))[1], 'z'))[2] as zone_number,
    seat->>'zone' as z_name, t.price::numeric::float
	from seats s, tariff t 
	where s.tariff_id = t.tariff_id 
) as dd group by zone_number, z_name order by zone_number`
)

// database структура подключения к базе данных
type Database struct {
	Conn *sqlx.DB
}

func init() {
	cfg := config.LoadConfig("config.json")
	OfdVat = cfg.OfdData.Vat
	OfdMathod = cfg.OfdData.OfdMathod
	OfdOobject = cfg.OfdData.OfdOobject
	TaxSyst = cfg.OfdData.TaxSyst
	amoTickSend = cfg.PipelineStages.TicketSend
}

// dbService представляет интерфейс взаимодействия с базой данных
type DbService interface {
	GetLastId(table string) (int, error)
	GetSeats() ([]ent.Seat, error)
	GetSeatStates() ([]ent.SeatState, error)
	GetSeatsInfo() ([]ent.SeatInfo, error)
	SetSeatStates() error
	GetEventTarifs() ([]ent.Tafif, error)
	CheckSeatStatess() (bool, error)
	ReserveSeat() (bool, error)
	UnReserveSeat() (bool, error)
	CreateOrder() (string, string, float32, []*ent.Item, error)
	GetNewOrderNumber() (string, error)
	CalculateOrderAmount() (float32, []*ent.Item, error)
	ClearExpiredReserves() ([]string, error)

	OrderLog() (bool, error)

	SendTickets() (bool, error)
	GetLastTicketNumber() (string, error)
}

// newDB открывает соединение с базой данных
func NewDB(connectionString string) (Database, error) {
	dbConn, err := sqlx.Open("postgres", connectionString)
	log.Println(connectionString)
	return Database{Conn: dbConn}, err
}

// #################################################################
func (db Database) GetSeats(event_id int, tarif string) ([]*ent.Seat, error) {

	seats := make([]*ent.Seat, 0)
	//err := db.Conn.Select(&users, authQuery, login, password)

	stmt, err := db.Conn.Prepare(getSeatsStatusesQuery)
	if err != nil {
		return seats, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(event_id, tarif)

	for rows.Next() {
		var seat_id int
		var tariff_id int
		var seat string
		var svg_id string
		var active_flag int
		var state int
		err = rows.Scan(&seat_id, &tariff_id, &seat, &svg_id, &active_flag, &state)
		if err != nil {
			return seats, err
		}
		s := ent.Seat{seat_id, tariff_id, seat, svg_id, active_flag, state}
		seats = append(seats, &s)
	}

	return seats, err
}

func (db Database) GetSeatStatess(event_id int) ([]*ent.SeatState, error) {

	states := make([]*ent.SeatState, 0)
	//err := db.Conn.Select(&users, authQuery, login, password)

	stmt, err := db.Conn.Prepare(getSeatStatesQuery)
	if err != nil {
		return states, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(event_id)
	if err != nil {
		return states, err
	}
	for rows.Next() {
		var seat_id int
		var svg_id string
		var state int
		err = rows.Scan(&seat_id, &svg_id, &state)
		if err != nil {
			return states, err
		}
		st := ent.SeatState{seat_id, svg_id, state}
		states = append(states, &st)
	}

	return states, err
}

func (db Database) SetSeatStatess(seatIds string, state int) (bool, error) {

	stmt, err := db.Conn.Prepare(setSeatStateQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	iDs := strings.Split(seatIds, " ")

	for _, v := range iDs {
		_, err = stmt.Exec(state, v)
		if err != nil {
			log.Println("Error while updateng seat state: svg_id = %v", err)
			break
		}
		log.Printf("Update seat state: svg_id = %v to 1 - succes", v)
	}

	return true, err
}

func (db Database) CheckSeatStatess(seatIds string, expectState int) (bool, error) {

	var state int

	stmt, err := db.Conn.Prepare(checkSeaState)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(seatIds)
	if err != nil {
		log.Println("Error while updateng seat state: svg_id = %v", err)
		return false, err
	}

	if rows.Next() {
		err = rows.Scan(&state)
		if err != nil {
			return false, err
		}
	}
	rows.Close()

	return state == expectState, err
}

func (db Database) GetSeatsInfo(seatIds string) ([]*ent.SeatInfo, error) {

	sInfo := make([]*ent.SeatInfo, 0)

	stmt, err := db.Conn.Prepare(getSeatsInfoQuery)
	if err != nil {
		return sInfo, err
	}
	defer stmt.Close()

	iDs := strings.Split(seatIds, " ")
	for _, v := range iDs {

		rows, err := stmt.Query(v)
		if err != nil {
			log.Println("Error while select seat info:  = %v", err)
		}
		info := ent.SeatInfo{}
		for rows.Next() {
			var svg_id string
			var tarif_name string
			var zone string
			var row_number int
			var seat_number int
			var price int
			err = rows.Scan(&svg_id, &tarif_name, &zone, &row_number, &seat_number, &price)
			if err != nil {
				return sInfo, err
			}
			info = ent.SeatInfo{svg_id, tarif_name, zone, row_number, seat_number, price}
		}
		sInfo = append(sInfo, &info)
	}

	return sInfo, err
}

func (db Database) GetEventTarifs(evenId int) ([]*ent.Tafif, error) {

	et := make([]*ent.Tafif, 0)

	stmt, err := db.Conn.Prepare(getEventTarifsQuery)
	if err != nil {
		return et, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(evenId)
	if err != nil {
		log.Println("Error while select event tarifs:  = %v", err)
	}
	t := ent.Tafif{}
	for rows.Next() {
		var (
			tarifId     int
			eventId     int
			price       int
			priceCur    string
			tarifName   string
			tarifZoneId int
			count       int
		)
		err = rows.Scan(&tarifId, &eventId, &price, &priceCur, &tarifName, &tarifZoneId, &count)
		if err != nil {
			return et, err
		}
		t = ent.Tafif{tarifId, eventId, price, priceCur, tarifName, tarifZoneId, count}
		et = append(et, &t)
	}

	return et, err
}

func (db Database) ReserveSeat(svgId string) (bool, error) {
	var seatId int
	//var seat_state int

	stmt, err := db.Conn.Prepare(checkSeatForReserve)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(svgId)
	if err != nil {
		log.Println("Error while check seat stae for reserve:  = %v", err)
	}
	for rows.Next() {
		err = rows.Scan(&seatId)
		if err != nil {
			return false, err
		}
	}

	stmt, err = db.Conn.Prepare(reserveSeatQuery)
	if err != nil {
		return false, err
	}
	_, err = stmt.Exec(seatId)
	if err != nil {
		return false, err
	}

	return true, err
}

func (db Database) UnReserveSeat(svgId string) (bool, error) {

	stmt, err := db.Conn.Prepare(unreserveSeatQuery)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(svgId)
	if err != nil {
		return false, err
	}

	return true, err
}

func (db Database) OrderLog(act string, stage string, orderNum string, code string, message string, success string, reson string) (bool, error) {

	pp := "{ \"stage\":\"" + stage + "\", \"code\":\"" + code + "\", \"message\":\"" + message + "\", \"success\":\"" + success + "\", \"reason\":\"" + reson + "\" }"
	orderLogData := "[{\"log_time\":\"" + time.Now().Format("2006-01-02T15:04:05") + "\", \"action\":\"" + act + "\",\"actor\":\"user\",\"params\":" + pp + " }]"

	log.Printf("pp: %v, orderLogData: %v ", pp, orderLogData)

	stmt, err := db.Conn.Prepare(orderLog)
	if err != nil {
		log.Println("Error while Log order:  = %v", err)
		return false, err
	}
	defer stmt.Close()

	var status int
	if stage == "orderCreate" {
		status = 0
	}
	if stage == "payComplete" { //Платеж совершен но не подтвержден
		status = 1
	}
	if stage == "payFail" { //Платеж откланен или ошибка
		status = 3
	}
	if stage == "paySuccess" { //Платеж совершен и подтвержден
		status = 4
	}
	if stage == "ticketAsignedForSeat" { // Данные для билета отправлены получен 200 ОК
		status = 5
	}
	if stage == "ticketSendNoSucces" { // Данные для билета не отправлены ошибка отправки
		status = 6
	}
	if stage == "ticketSendSucces" { // Данные для билета отправлены получен 200 ОК
		status = 7
	}
	//if stage == "ticketSendNoSucces" { // Данные для билета не отправлены ошибка отправки
	//	status = 8
	//}

	_, err = stmt.Exec(orderLogData, status, orderNum)
	if err != nil {
		log.Println("Error while Log order:  = %v", err)
		return false, err
	}
	return true, err
}

func (db Database) UpdateOrder(orderNumber string, name string, email string, phone string, svgIds string, eventId int, utm string, amoData ent.AmoInfoData, refId string) (string, float32, float32, []*ent.Item, error) {

	//определение суммы заказа
	amount, dAmount, items, err := db.CalculateOrderAmount(svgIds, refId)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, 0, nil, err
	}

	stmt, err := db.Conn.Prepare(updateOrder)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, 0, nil, err
	}
	defer stmt.Close()

	crmData := "{\"crm\":{\"contact_id\":" + amoData.CONTACTID + ",\"lead_id\":" + amoData.LEADID + "}}"

	pp := name + "|" + email + "|" + name + "|" + phone + "|" + svgIds
	orderLog := "[{\"log_time\":\"" + time.Now().Format("2006-01-02T15:04:05") + "\", \"action\":\"updateCreate\",\"actor\":\"system\",\"params\":\"" + pp + "\"}]"
	cData := "{\"name\":\"" + name + "\", \"email\":\"" + email + "\", \"phone\":\"" + phone + "\",\"seats\":\"" + svgIds + "\", \"utm\": " + utm + ", \"crm\": " + crmData + " }"

	_, err = stmt.Exec(orderLog, cData, amount, orderNumber)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, 0, nil, err
	}
	return orderNumber, amount, dAmount, items, err

}

func (db Database) CreateOrder(name string, email string, phone string, svgIds string, eventId int, utm string, refId string) (string, float32, float32, []*ent.Item, error) {
	//определение следующего номера заказа
	orderNumber, err := db.GetNewOrderNumber(eventId)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, 0, nil, err
	}

	//определение суммы заказа
	amount, dAmount, items, err := db.CalculateOrderAmount(svgIds, refId)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, 0, nil, err
	}

	// insert into orders (order_number, order_status, order_log, customer_data, event_id) values ($1, $2. $3, $4, $5);
	stmt, err := db.Conn.Prepare(createNewOrder)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, 0, nil, err
	}
	defer stmt.Close()

	pp := name + "|" + email + "|" + name + "|" + phone + "|" + svgIds
	orderLog := "[{\"log_time\":\"" + time.Now().Format("2006-01-02T15:04:05") + "\", \"action\":\"orderCreate\",\"actor\":\"system\",\"params\":\"" + pp + "\"}]"

	cData := "{\"name\":\"" + name + "\", \"email\":\"" + email + "\", \"phone\":\"" + phone + "\",\"seats\":\"" + svgIds + "\", \"utm\":" + utm + ", \"crm\": {} }"

	_, err = stmt.Exec(orderNumber, 0, orderLog, cData, eventId, amount, "RUB")
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, 0, nil, err
	}

	stmt, err = db.Conn.Prepare(getOrderIdByNumber)
	if err != nil {
		log.Println("Error while get orderId:  = %v", err)
		return "", 0, 0, nil, err
	}
	rows, err := stmt.Query(orderNumber)
	if err != nil {
		log.Println("Error while get orderId:  = %v", err)
		return "", 0, 0, nil, err
	}
	var orderID int
	for rows.Next() {
		err = rows.Scan(&orderID)
		if err != nil {
			log.Println("Error while get orderId:  = %v", err)
		}
	}

	stmt, err = db.Conn.Prepare(assignSeatToOrder)
	if err != nil {
		log.Println("Error while assign seat to order order:  = %v", err)
		return "", 0, 0, nil, err
	}
	_, err = stmt.Exec(orderID, svgIds)
	if err != nil {
		log.Println("Error while assign seat to order order:  = %v", err)
		return "", 0, 0, nil, err
	}

	return orderNumber, amount, dAmount, items, err
}

func (db Database) AddCrmData(contactId string, leadId string, orderNum string) (bool, error) {

	stmt, err := db.Conn.Prepare(addCrmDataInOrder)
	if err != nil {
		log.Println("Error while add crm data in order:  = %v", err)
		return false, err
	}
	crmData := "{\"crm\":{\"contact_id\":" + contactId + ",\"lead_id\":" + leadId + "}}"
	_, err = stmt.Exec(crmData, orderNum)
	if err != nil {
		log.Println("Error while add crm data in order:  = %v", err)
		return false, err
	}

	return true, nil
}

func (db Database) GetDiscountByRefID(refId string) (ent.DiscountRef, error) {
	var discount ent.DiscountRef
	stmt, err := db.Conn.Prepare(getDiscountByRefId)
	if err != nil {
		log.Println("Error while calculete amount:  = %v", err)
		return discount, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(refId)
	if err != nil {
		log.Println("Error while calculete discount:  = %v", err)
		return discount, err
	}

	for rows.Next() {
		var (
			id     int
			rId    string
			dType  string
			amount int
			count  int
		)
		err = rows.Scan(&id, &rId, &dType, &amount, &count)
		if err != nil {
			log.Println("Error while calculete amount:  = %v", err)
			return discount, err
		}
		discount = ent.DiscountRef{id, rId, dType, amount, count}
	}

	return discount, nil
}

func (db Database) CalculateOrderAmount(svgIds string, refId string) (float32, float32, []*ent.Item, error) {

	items := make([]*ent.Item, 0)

	//вычисление скидки
	stmt, err := db.Conn.Prepare(getDiscountByRefId)
	if err != nil {
		log.Println("Error while calculete amount:  = %v", err)
		return 0, 0, nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(refId)
	if err != nil {
		log.Println("Error while calculete discount:  = %v", err)
		return 0, 0, items, err
	}
	var discount ent.DiscountRef
	for rows.Next() {
		var (
			id     int
			rId    string
			dType  string
			amount int
			count  int
		)
		err = rows.Scan(&id, &rId, &dType, &amount, &count)
		if err != nil {
			log.Println("Error while calculete amount:  = %v", err)
			return 0, 0, items, err
		}
		discount = ent.DiscountRef{id, rId, dType, amount, count}
	}
	//вычисление сумм по каждому месту с применением скидки
	stmt, err = db.Conn.Prepare(getAmountBySeats)
	if err != nil {
		log.Println("Error while calculete amount:  = %v", err)
		return 0, 0, nil, err
	}
	rows, err = stmt.Query(svgIds)
	if err != nil {
		log.Println("Error while calculete amount:  = %v", err)
		return 0, 0, items, err
	}
	var amount float64
	var dAmount float64
	for rows.Next() {
		var (
			cZone string
			cRow  string
			cSeat string
			cPice int
			cCurr string
		)
		err = rows.Scan(&cZone, &cRow, &cSeat, &cPice, &cCurr)
		if err != nil {
			log.Println("Error while calculete amount:  = %v", err)
			return 0, 0, items, err
		}
		var dPrice float64
		if discount.AMOUNT > 0 && !(strings.Contains(svgIds, "e1z1r") ||
			strings.Contains(svgIds, "e1z2r") ||
			strings.Contains(svgIds, "e1z11r") ||
			strings.Contains(svgIds, "e1z12r")) {

			if discount.DISC_TYPE == "procent" {
				dPrice = roundFloat(float64(cPice)-float64(cPice)*(float64(discount.AMOUNT)/100), 2)
			}
			if discount.DISC_TYPE == "summ" {
				dPrice = roundFloat(float64(cPice)-float64(discount.AMOUNT), 2)
			}
		} else {
			dPrice = float64(cPice)
		}

		amount += dPrice
		dAmount += (float64(cPice) - dPrice)
		item := ent.Item{"Билет на мероприятие Fortune2050: " + cZone + " ряд " + cRow + " место " + cSeat,
			dPrice, float32(1.00), dPrice, OfdVat, OfdMathod, OfdOobject, "шт"}
		items = append(items, &item)
	}

	return float32(amount), float32(dAmount), items, nil
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func (db Database) GetNewOrderNumber(eventId int) (string, error) {
	stmt, err := db.Conn.Prepare(getLastOrderByEvent)
	if err != nil {
		log.Println("Error check last order for event:  = %v", err)
		return "", err
	}
	defer stmt.Close()
	rows, err := stmt.Query(eventId)
	if err != nil {
		log.Println("Error check last order for event:  = %v", err)
		return "", err
	}
	var orderNumber string
	for rows.Next() {
		var orderNum string
		err = rows.Scan(&orderNum)
		if err != nil {
			log.Println("Error check last order for event:  = %v", err)
			return "", err
		}
		orderNumber = orderNum
	}

	part := strings.Split(orderNumber, "_")
	num, err := strconv.Atoi(part[1])
	if err != nil {
		log.Println("err:", err.Error())
		return "", err
		num = 0
	}
	num++
	nn := strconv.Itoa(num)
	if err != nil {
		log.Println("err:", err.Error())
		return "", err
	}
	orderNumber = part[0] + "_" + nn
	return orderNumber, err
}

func (db Database) ClearExpiredReserves() ([]string, error) {
	var stt []string

	stmt, err := db.Conn.Prepare(getSeatsExpierReserv)
	if err != nil {
		log.Println("Error while clear expier reserves:  = %v", err)
		return stt, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Println("Error while clear expier reserves:  = %v", err)
		return stt, err
	}
	for rows.Next() {
		var (
			svgId string
			sId   int
		)
		err = rows.Scan(&sId, &svgId)
		if err != nil {
			log.Println("Error scan Data:  = %v", err)
			return stt, err
		}
		//st := ent.SeatState{sId, svgId, 0}
		stt = append(stt, svgId)
	}

	stmt, err = db.Conn.Prepare(clearExpiredReserves)
	if err != nil {
		log.Println("Error while clear expier reserves:  = %v", err)
		return nil, err
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Println("Error while clear expier reserves:  = %v", err)
		return nil, err
	}

	stmt, err = db.Conn.Prepare(clearStatusForUnreserved)
	if err != nil {
		log.Println("Error while clear status for unresed seats:  = %v", err)
		return nil, err
	}
	_, err = stmt.Exec()
	if err != nil {
		log.Println("Error while clear status for unresed seats:  = %v", err)
		return nil, err
	}

	return stt, nil
}

func (db Database) GetSeatTarif(event_id int) ([]*ent.SeatTarif, error) {
	st := make([]*ent.SeatTarif, 0)
	stmt, err := db.Conn.Prepare(getSeatTarif)
	if err != nil {
		log.Println("Error while get tarifs:  = %v", err)
		return st, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(event_id)
	if err != nil {
		log.Println("Error while get tarifs:  = %v", err)
		return st, err
	}
	for rows.Next() {
		var (
			tPice int
			tName string
			sId   string
		)
		err = rows.Scan(&tPice, &tName, &sId)
		if err != nil {
			log.Println("Error while get tarifs:  = %v", err)
			return st, err
		}
		stt := ent.SeatTarif{tPice, tName, sId}
		st = append(st, &stt)
	}

	return st, nil
}

func (db Database) GetSeatInfoBySvgId(svgIds string) (string, string, string, error) {

	stmt, err := db.Conn.Prepare(getSeatInfoBySvgId)
	if err != nil {
		return "", "", "", err
	}
	defer stmt.Close()
	rows, err := stmt.Query(svgIds)
	if err != nil {
		return "", "", "", err
	}
	rZone := ""
	rRow := ""
	rSeat := ""

	for rows.Next() {
		var (
			zone string
			row  string
			seat string
		)
		err = rows.Scan(&zone, &row, &seat)
		if err != nil {
			log.Println("Error while get seats info:  = %v", err)
			return "", "", "", err
		}
		if !strings.Contains(rZone, zone) {
			rZone += "," + zone
		}
		if !strings.Contains(rRow, row) {
			rRow += "," + row
		}

		rSeat += "," + seat
	}

	return rZone[1:], rRow[1:], rSeat[1:], err
}

func (db Database) SendTickets() (bool, error) {
	stt := make([]*ent.TicketForSend, 0)
	stmt, err := db.Conn.Prepare(getTicketsForSend)
	if err != nil {
		log.Println("Error while get tikets fo send:  = %v", err)
		return false, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Error while get tikets fo send:  = %v", err)
		return false, err
	}
	for rows.Next() {
		var (
			oStatus int
			oDate   string
			oNumb   string
			name    string
			email   string
			phone   string
			sect    string
			row     string
			seat    string
			seat_id int
			leadId  string
		)
		err = rows.Scan(&oStatus, &oDate, &oNumb, &name, &email, &phone, &sect, &row, &seat, &seat_id, &leadId)
		if err != nil {
			if strings.Contains(err.Error(), "lead_id") {
				leadId = ""
				log.Println("Error lead_id = null", err)
			} else {
				log.Println("Error while get tikets fo send:  = %v", err)
				return false, err
			}
		}
		var tNumb string
		if oStatus == 4 {
			tNumb, err = db.GetLastTicketNumber(seat_id)
			if err != nil {
				fmt.Println(err)
				return false, err
			}
		} else {
			tNumb, err = db.GetTicketNumber(seat_id)
			if err != nil {
				fmt.Println(err)
				return false, err
			}
		}

		st := ent.TicketForSend{oDate, oNumb, tNumb, name, email, phone, sect, row, seat, leadId}
		stt = append(stt, &st)
	}

	sts, err := esend.SendTickets(stt)
	if err != nil {
		fmt.Println(err)
		//return false, err
	}

	tMap := make(map[string]string)
	var tickets string
	var ord string
	for i, sst := range sts {

		//Log  Order
		var stage string
		if sst.STATUS == "OK" {
			stage = "ticketSendSucces"
			db.OrderLog("ticketSend", stage, sst.ORDER_NUMBER, "0", "Билет успешно отправлен: "+sst.TICKET_NUMBER, "true", sst.REASON)
			if ord == sst.ORDER_NUMBER || i == 0 {
				tickets += "," + sst.TICKET_NUMBER
			} else {
				tickets = "," + sst.TICKET_NUMBER
			}
			tMap[sst.LAED] = tickets
			ord = sst.ORDER_NUMBER
		} else {
			stage = "ticketSendNoSucces"
			db.OrderLog("ticketSend", stage, sst.ORDER_NUMBER, "0", "Билет небыл отправлен: "+sst.TICKET_NUMBER, "false", sst.REASON)
		}
	}

	if len(tMap) > 0 {
		//Update CRM lead status
		for k, v := range tMap {
			if k != "" {
				status, _ := amo.LeadStatusTicketsUpdate(k, amoTickSend, v[1:]) //"48793132")
				if status != "OK" {
					log.Printf("Ошибка смены статуса сделки в Амо CRM : ): %v", status)
				} else {
					log.Printf("Сделка %v переведена в статус %v в Амо CRM : ): ", k, amoTickSend)
				}

			} else {
				log.Printf("Ошибка смены статуса сделки в Амо CRM : ): lead_id = null")
			}

		}

	}

	return true, nil
}

func (db Database) GetTicketNumber(seat_id int) (string, error) {
	tNumber := ""
	stmt, err := db.Conn.Prepare(getTicketNumberBySeatId)
	if err != nil {
		log.Println("Error while get tikets number:  = %v", err)
		return tNumber, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(seat_id)
	if err != nil {
		log.Println("Error while get tikets number:  = %v", err)
		return tNumber, err
	}
	for rows.Next() {
		err = rows.Scan(&tNumber)
		if err != nil {
			log.Println("Error while get tikets number:  = %v", err)
			return tNumber, err
		}
	}

	return tNumber, nil
}

func (db Database) GetLastTicketNumber(seat_id int) (string, error) {
	tNumber := ""
	stmt, err := db.Conn.Prepare(getLastTicketNumber)
	if err != nil {
		log.Println("Error while get tikets number:  = %v", err)
		return tNumber, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Error while get tikets number:  = %v", err)
		return tNumber, err
	}
	for rows.Next() {
		err = rows.Scan(&tNumber)
		if err != nil {
			log.Println("Error while get tikets number:  = %v", err)
			return tNumber, err
		}
		stmt, err = db.Conn.Prepare(updateTicketStatus)
		if err != nil {
			log.Println("Error while update tikets status:  = %v", err)
			return tNumber, err
		}
		_, err := stmt.Exec(seat_id, tNumber)
		if err != nil {
			log.Println("Error while get tikets number:  = %v", err)
			return tNumber, err
		}
		//Log Assign Tickets
		_, err = db.OrderLog("asigneTicket", "ticketAsignedForSeat", "this", "", "tNumber="+tNumber, "true", "")
		if err != nil {
			log.Println("Error while get tikets number:  = %v", err)
			return tNumber, err
		}
	}

	return tNumber, nil
}

//func (db Database) assignTicketForSeat(svgId string) (bool, error) {
//
//	stmt, err := db.Conn.Prepare(unreserveSeatQuery)
//	if err != nil {
//		return false, err
//	}
//	defer stmt.Close()
//
//	_, err = stmt.Exec(svgId)
//	if err != nil {
//		return false, err
//	}
//
//	return true, err
//}

func (db Database) GetMaxMinZoneTarifs() ([]*ent.MaxMinTafif, error) {
	stt := make([]*ent.MaxMinTafif, 0)
	stmt, err := db.Conn.Prepare(selectZoneMaxMin)
	if err != nil {
		log.Println("Error while max t values:  = %v", err)
		return stt, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Error while max t values:  = %v", err)
		return stt, err
	}
	for rows.Next() {
		var (
			zone int
			name string
			max  int
			min  int
		)

		err = rows.Scan(&zone, &name, &max, &min)
		st := ent.MaxMinTafif{zone, name, max, min}
		stt = append(stt, &st)
	}

	return stt, nil
}
