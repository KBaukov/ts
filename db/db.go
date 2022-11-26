package db

import (
	_ "encoding/json"
	"github.com/KBaukov/ts/ent"
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

//var (
//	RoomData    =	make(map[string]ent.SensorsData)
//	FloorData    =	make(map[string]ent.FloorData)
//	sensorMx	=	sync.RWMutex{}
//)

const (
	getSeatsStatusesQuery = `select * from seats where tariff_id i= (select tariff_id from tariff where event_id= $1 and tariff_name=$2) order by seat_id;`

	getSeatStatesQuery = `select seat_id, svg_id, state from seats where tariff_id in (select tariff_id from tariff where event_id= $1 ) order by seat_id;`
	setSeatStateQuery  = `update seats set state = $1 where svg_id = $2;`

	getSeatsInfoQuery = `select s.svg_id, t.tariff_name, s.seat->>'zone' as zone, s.seat->>'row' as row_number, s.seat->>'place' as seat_number,  
t.price::numeric::int as price from seats s, tariff t  
where s.svg_id  = $1
and s.tariff_id =t.tariff_id;`

	getEventTarifsQuery = `select tariff_id, event_id, price::numeric::int as price, price_currency, tariff_name, tariff_zone_id, count 
							from tariff where event_id = $1 order by tariff_id;`

	reserveSeatQuery    = `insert into reserv (seat_id) values ($1);`
	unreserveSeatQuery  = `delete from reserv where seat_id = (select seat_id from seats where svg_id = $1);`
	checkSeatForReserve = `select seat_id from seats where svg_id = $1 and state = 0;`

	getLastOrderByEvent = `select order_number from orders where order_id = (select max(order_id) as order_id from orders where event_id = $1 );`
	createNewOrder      = `insert into orders (order_number, order_staus, order_log, customer_data, event_id, amount, amount_currency) 
values ($1, $2, ( $3 )::jsonb, ( $4 )::jsonb, $5, $6, $7);`

	getAmountBySeats = `select sum(t.price)::numeric::int as amount, t.price_currency  from seats s, tariff t  
where s.svg_id in (select unnest(string_to_array($1, ' ')) as dd)
and s.tariff_id = t.tariff_id
group by price_currency;`
)

// database структура подключения к базе данных
type Database struct {
	Conn *sqlx.DB
}

// dbService представляет интерфейс взаимодействия с базой данных
type DbService interface {
	GetLastId(table string) (int, error)

	GetSeats() ([]ent.Seat, error)

	GetSeatStates() ([]ent.SeatState, error)

	GetSeatsInfo() ([]ent.SeatInfo, error)

	SetSeatStates() error

	GetEventTarifs() ([]ent.Tafif, error)

	ReserveSeat() (bool, error)
	UnReserveSeat() (bool, error)

	CreateOrder() (string, int, error)

	GetNewOrderNumber() (string, error)

	CalculateOrderAmount() (int, error)

	//GetTreeNodeDiff(key string, env string,nextEnv string) ([]ent.RegistryItem, error)
	//GetNodeChilds(key string, env string) ([]ent.RegistryItemDiff, error)
	//GetRegistry(int, string) ([]*ent.RegistryItem, error)
	//RegistryNodeCreate(int, string, string, int, string, string, string, ent.User) (int, error)
	//RegistryNodeUpdate(int, int, int, string, string, int, string, string, string, ent.User) error
	//RegistryNodeDel(int, string, ent.User) error
	//GetRegistryNode(string, string) (ent.RegistryItem, error)
	//GetDictionary(string, string) ([]*ent.DictionaryItem, error)
	//GetAccessRules(string) ([]*ent.AuthorizationExpression, error)

}

// newDB открывает соединение с базой данных
func NewDB(connectionString string) (Database, error) {
	dbConn, err := sqlx.Open("postgres", connectionString)
	log.Println(connectionString)
	return Database{Conn: dbConn}, err
}

// #################################################################
func (db Database) GetSeats(event_id int, tarif string) ([]*ent.Seat, error) {

	//db.Conn.Conn(nil);

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

func (db Database) CreateOrder(name string, email string, phone string, svgIds string, eventId int) (string, int, error) {
	//определение следующего номера заказа
	orderNumber, err := db.GetNewOrderNumber(eventId)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, err
	}

	//определение суммы заказа
	amount, err := db.CalculateOrderAmount(svgIds)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, err
	}

	// insert into orders (order_number, order_staus, order_log, customer_data, event_id) values ($1, $2. $3, $4, $5);
	stmt, err := db.Conn.Prepare(createNewOrder)
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, err
	}
	defer stmt.Close()

	pp := name + "|" + email + "|" + name + "|" + phone + "|" + svgIds
	orderLog := "{\"log_time\":\"" + time.Now().Format("2006-01-02T15:04:05") + "\", \"action\":\"orderCreat\",\"actor\":\"system\",\"params\":\"" + pp + "\"}"

	cData := "{\"name\":\"" + name + "\", \"email\":\"" + email + "\", \"phone\":\"" + phone + "\",\"seats\":\"" + svgIds + "\" }"

	_, err = stmt.Exec(orderNumber, 0, orderLog, cData, eventId, amount, "RUB")
	if err != nil {
		log.Println("Error while create new order:  = %v", err)
		return "", 0, err
	}
	return orderNumber, amount, err
}

func (db Database) CalculateOrderAmount(svgIds string) (int, error) {
	stmt, err := db.Conn.Prepare(getAmountBySeats)
	if err != nil {
		log.Println("Error while calculete amount:  = %v", err)
		return 0, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(svgIds)
	if err != nil {
		log.Println("Error while calculete amount:  = %v", err)
		return 0, err
	}
	var amount int
	for rows.Next() {
		var (
			cPice int
			cCurr string
		)
		err = rows.Scan(&cPice, &cCurr)
		if err != nil {
			log.Println("Error while calculete amount:  = %v", err)
			return 0, err
		}
		amount += cPice
	}

	return amount, nil
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
		var (
			//orderId     int
			orderNum string
			//orderStatus int
			//orderLog    ent.ActionLog
			//custData    ent.PayDataExt
			//evId        int
			//amount      int
			//curr        string
		)
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

//func (db Database) CreateSessinDB(user ent.User) (string, error) {
//	//tt := time.Now()
//	sessId, err := HashSess(user.LOGIN + user.USER_TYPE + time.Now().String())
//	if err != nil {
//		return "", err
//	}
//
//	stmt, err := db.Conn.Prepare(createSession)
//	if err != nil {
//		return "", err
//	}
//	defer stmt.Close()
//
//	//dur,_ := time.ParseDuration("30m")
//
//	_, err = stmt.Exec(sessId, user.ID, 30)
//	if err != nil {
//		return "", err
//	}
//
//	return sessId, nil
//}
//
//func (db Database) CheckSessinDB(token string) (int, error) {
//	stmt, err := db.Conn.Prepare(getSessionByID)
//	if err != nil {
//		return 0, err
//	}
//	defer stmt.Close()
//
//	rows, err := stmt.Query(token)
//	if err != nil {
//		return 0, err
//	}
//
//	for rows.Next() {
//		var sid string
//		var userId int
//		var expT time.Time
//		err = rows.Scan(&sid, &userId, &expT)
//		if err != nil {
//			return 0, err
//		}
//
//		if expT.Unix() < time.Now().Unix() {
//			DeleteSess(db, sid)
//			return 0, errors.New("Сесия протухла. ")
//		}
//
//		return userId, nil
//	}
//	return 0, errors.New("Сесия не найдена. ")
//}
//
//func DeleteSess(db Database, sessId string) error {
//	stmt, err := db.Conn.Prepare(deleteSession)
//	if err != nil {
//		return err
//	}
//	defer stmt.Close()
//
//	_, err = stmt.Exec(sessId)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (db Database) DeleteSessinDB(sessId string) error {
//	return DeleteSess(db, sessId)
//}
//
//func (db Database) UpdateSessinDB(token string) error {
//	stmt, err := db.Conn.Prepare(updateSession)
//	if err != nil {
//		return err
//	}
//	defer stmt.Close()
//
//	_, err = stmt.Exec(30, token)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//// ##################################################################
//func (db Database) GetLastId(table string) (int, error) {
//	var lastId int
//	stmt, err := db.Conn.Prepare("SELECT max(id) as id FROM public." + table)
//	if err != nil {
//		return -1, err
//	}
//	defer stmt.Close()
//
//	rows, err := stmt.Query()
//	for rows.Next() {
//		err = rows.Scan(&lastId)
//		if err != nil {
//			return -1, err
//		}
//	}
//
//	return lastId, err
//}
//
//// ############## Users ############################
//func (db Database) GetUsers() ([]ent.User, error) {
//	users := make([]ent.User, 0)
//	stmt, err := db.Conn.Prepare(getUsersQuery)
//	if err != nil {
//		return users, err
//	}
//	defer stmt.Close()
//
//	rows, err := stmt.Query()
//
//	for rows.Next() {
//		var (
//			uid      int
//			login    string
//			pass     string
//			active   string
//			userType string
//			lastV    time.Time
//		)
//		err = rows.Scan(&uid, &login, &pass, &active, &userType, &lastV)
//		if err != nil {
//			return users, err
//		}
//		u := ent.User{uid, login, pass, active, userType, lastV}
//		users = append(users, u)
//	}
//
//	return users, err
//}
//
//func (db Database) GetUser(userId int) (ent.User, error) {
//	var user ent.User
//	stmt, err := db.Conn.Prepare(getUserByIdQuery)
//	if err != nil {
//		return user, err
//	}
//	defer stmt.Close()
//
//	rows, err := stmt.Query(userId)
//
//	for rows.Next() {
//		var (
//			uid      int
//			login    string
//			pass     string
//			active   string
//			userType string
//			lastV    time.Time
//		)
//		err = rows.Scan(&uid, &login, &pass, &active, &userType, &lastV)
//		if err != nil {
//			return user, err
//		}
//		user = ent.User{uid, login, pass, active, userType, lastV}
//		break
//	}
//
//	return user, err
//}
//
//func (db Database) UpdUser(id int, login string, pass string, userType string, actFlag string, lastV time.Time) (bool, error) {
//
//	var lastId int
//
//	execQuery := updUserQuery
//
//	lastId, err := db.GetLastId("users")
//	if err != nil {
//		return false, err
//	}
//
//	if id > lastId {
//		execQuery = addUserQuery
//	}
//
//	stmt, err := db.Conn.Prepare(execQuery)
//	if err != nil {
//		return false, err
//	}
//
//	_, err = stmt.Exec(login, pass, userType, actFlag, lastV, id)
//	if err != nil {
//		return false, err
//	}
//
//	return true, err
//}
//
//func (db Database) DelUser(id int) (bool, error) {
//
//	stmt, err := db.Conn.Prepare(delUserQuery)
//	if err != nil {
//		return false, err
//	}
//	defer stmt.Close()
//
//	_, err = stmt.Exec(id)
//	if err != nil {
//		return false, err
//	}
//
//	return true, err
//}
//
//// ############## Users ############################
//
//func HashSess(p string) (string, error) {
//	h := sha256.New()
//	_, err := h.Write([]byte(p))
//	if err != nil {
//		return "", err
//	}
//
//	return hex.EncodeToString(h.Sum(nil)), nil
//}
