package handle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/KBaukov/ts/config"
	"github.com/KBaukov/ts/db"
	"github.com/KBaukov/ts/ent"
	"github.com/Nerzal/gocloak"
	"github.com/gorilla/sessions"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	sessStore     = sessions.NewCookieStore([]byte("33446a9dcf9ea060a0a6532b166da32f304af0de"))
	cfg           = config.LoadConfig("config.json")
	mainTemplate  = cfg.FrontRoute.MainTemplate
	loginTemplate = cfg.FrontRoute.LoginTemplate
	webres        = cfg.FrontRoute.WebResFolder
	clientID      = "ts_app"
	clientSecret  = "7c093669-2e99-41dd-b25f-6cbb7a24dd8d"
	redirectURL   = "http://localhost:8081/login"
	configURL     = "http://10.200.42.66:8080"
	state         = "somestate"
	realm         = "TicketSystem"
	client        gocloak.GoCloak
	ctx           context.Context
)

type sessData struct {
	token    gocloak.JWT
	userInfo map[string]interface{}
	user     ent.User
}

func init() {

}

type errData struct {
	Error_Code    int
	Error_Message string
}

func ServeHome(w http.ResponseWriter, r *http.Request) {
	log.Println("###: ", r.URL.Path)
	http.ServeFile(w, r, "./webres/index.html")
}

func ServeWebRes(w http.ResponseWriter, r *http.Request) {
	log.Println("###: ", r.URL.Path)
	//if strings.Contains(r.URL.Path, webres) {
	//	http.ServeFile(w, r, "."+r.URL.Path)
	//} else {
	http.ServeFile(w, r, "./"+webres+r.URL.Path)
	//}

}

func ServePagesRes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var ff string
	log.Println("###: ", path)
	if path == "/paysuccess" {
		ff = "/pages/paysuccess.html"
	}
	if path == "/crocusrules" {
		ff = "/pages/crocus_rules.html"
	}
	if path == "/confidential" {
		ff = "/pages/confidential.html"
	}
	if path == "/vozvrat" {
		ff = "/pages/vozvrat.html"
	}
	if path == "/oferta" {
		ff = "/pages/oferta.html"
	}

	http.ServeFile(w, r, "./"+webres+ff)
	//}

}

func ServeApi(db db.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("incoming request in: %v", r.URL.Path)
		//token := r.Header.Get("X-TOKEN")

		if r.URL.Path == "/api/seatmap" {

			eId := r.FormValue("event_id")
			eventId, err := strconv.Atoi(eId)
			if err != nil {
				log.Println("err:", err.Error())
			}
			tarif := r.FormValue("tarif")
			seatMap, err := db.GetSeats(eventId, tarif)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка получения карты мест (eventId: %v , tarif: %v): %v", eventId, tarif, err)
			}
			apiDataResponse(w, seatMap, err)
			return
		}

		if r.URL.Path == "/api/seatstates" && r.Method == "GET" { //

			eId := r.FormValue("event_id")
			eventId, err := strconv.Atoi(eId)
			if err != nil {
				log.Println("err:", err.Error())
			}
			//tarif := r.FormValue("tarif")
			seatMap, err := db.GetSeatStatess(eventId)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка получения статусов мест (eventId: %v ): %v", eventId, err)
			}
			apiDataResponse(w, seatMap, err)
			return
		}

		if r.URL.Path == "/api/seatstates" && r.Method == "POST" {

			seatIds := r.FormValue("seat_ids")
			s := r.FormValue("state")
			state, err := strconv.Atoi(s)
			if err != nil {
				log.Println("err:", err.Error())
			}

			_, err = db.SetSeatStatess(seatIds, state)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка установки статусов мест (seatIds: %v ): %v", seatIds, err)
			} else {
				msg := actionSeatUpdate(seatIds)
				hub.sendDataToWeb(msg, "TS_system", nil)
			}
			apiDataResponse(w, []int{}, err)
			return
		}

		if r.URL.Path == "/api/payaction" && r.Method == "POST" {

			seatIds := r.FormValue("seat_ids")
			orderNum := r.FormValue("order_number")
			code := r.FormValue("code")
			message := r.FormValue("message")
			success := r.FormValue("success")
			reson := r.FormValue("reson")
			stage := r.FormValue("stage")

			log.Printf("Incomming change {stage: %v, seatIds: %v, orderNum: %v, code: %v, message: %v, success: %v, reson: %v}",
				stage, seatIds, orderNum, code, message, success, reson)

			if stage == "payComplete" {
				_, err := db.SetSeatStatess(seatIds, 2)
				if err != nil {
					http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
					log.Printf("Ошибка смены статуса места (seatIds: %v ): %v", seatIds, err)
				} else {
					msg := actionSeatUpdate(seatIds)
					hub.sendDataToWeb(msg, "TS_system", nil)
				}
			}

			// Логирование в заказ
			_, err := db.OrderLog("pay", stage, orderNum, code, message, success, reson)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка записи в лог заказа (stage: %v, orderNum: %v, code: %v, message: %v, success: %v, reson: %v ): %v",
					stage, orderNum, code, message, success, reson, err)
			}

			apiDataResponse(w, []int{}, err)
			return
		}

		if r.URL.Path == "/api/messagesend" {

			message := r.FormValue("msg")

			hub.sendDataToWeb(message, "TS_system", nil)

			apiDataResponse(w, "", nil)
			return
		}

		if r.URL.Path == "/api/seatsinfo" && r.Method == "GET" {

			seatIds := r.FormValue("seat_ids")
			seatInfo, err := db.GetSeatsInfo(seatIds)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка запроса информации о местах (seatIds: %v ): %v", seatIds, err)
			}
			apiDataResponse(w, seatInfo, err)
			return
		}

		if r.URL.Path == "/api/eventtarif" && r.Method == "GET" {

			eId := r.FormValue("event_id")
			eventId, err := strconv.Atoi(eId)
			if err != nil {
				log.Println("err:", err.Error())
			}
			tariffs, err := db.GetEventTarifs(eventId)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка запроса информации о тарифах (eventId: %v ): %v", eventId, err)
			}
			apiDataResponse(w, tariffs, err)
			return
		}

		if r.URL.Path == "/api/order" && r.Method == "POST" {

			seatIds := r.FormValue("seats")
			name := r.FormValue("name")
			email := r.FormValue("email")
			phone := r.FormValue("phone")
			e := r.FormValue("event_id")
			eventId, err := strconv.Atoi(e)
			if err != nil {
				log.Println("err:", err.Error())
			}

			orderNumber, amount, err := db.CreateOrder(name, email, phone, seatIds, eventId)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка создания заказа (seatIds: %v ): %v", seatIds, err)
			} else {
				//msg := actionSeatUpdate(seatIds)
				//hub.sendDataToWeb(msg, "TS_system")
			}

			ext := ent.PayDataExt{name, email, phone, seatIds}
			pData := ent.PayData{cfg.PaySecrets.PKey, cfg.PaySecrets.Description, amount,
				cfg.PaySecrets.Curr, email, orderNumber, email,
				cfg.PaySecrets.Template, cfg.PaySecrets.AutoClose, ext}

			apiDataResponse(w, pData, nil)
			return
		}

		if r.URL.Path == "/api/seattarifs" && r.Method == "GET" {

			eId := r.FormValue("event_id")
			eventId, err := strconv.Atoi(eId)
			if err != nil {
				log.Println("err:", err.Error())
			}
			tariffs, err := db.GetSeatTarif(eventId)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка запроса информации о тарифах (eventId: %v ): %v", eventId, err)
			}
			apiDataResponse(w, tariffs, err)
			return
		}

		return
	}
}

func apiDataResponse(w http.ResponseWriter, data interface{}, err error) {
	errMsg := ""
	succes := true

	if err != nil {
		//http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка: %v", err)
		errMsg = err.Error()
		succes = false
	}

	dataResp := ent.ApiResp{SUCCESS: succes, DATA: data, MSG: errMsg}

	json, err := json.Marshal(dataResp)
	if err != nil {
		//http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		log.Printf("Ошибка маршалинга: %v", err)
		return
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	_, err = w.Write(json)
	if err != nil {
		log.Printf("Ошибка записи результата запроса: %v", err)
	}
}

// ########################## helpers ############################
func actionSeatUpdate(ids string) string {
	idss := strings.Split(ids, " ")
	var cmd = "{ \"action\":\"seatStateUpdate\", \"data\": ["
	for _, v := range idss {
		cmd += "\"" + v + "\" ,"
	}
	cmd += "\" \" ] }"
	return cmd
}

func HashPass(p string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(p))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func createSessionObject(w http.ResponseWriter, r *http.Request, o interface{}, key string) error {

	session, err := sessStore.Get(r, "RegistryAPP")
	if err != nil {
		log.Printf("Error of session storage: %v", err)
		return err
	}

	session.Values[key] = o
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error while save object in session: %v", err)
		return err
	}

	log.Println("Save object in session: succes", key)

	return nil
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := sessStore.Get(r, "RegistryAPP")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		session, err = sessStore.New(r, "RegistryAPP")
	}
	return session
}

func checkSession(t string) bool {
	//parts := strings.Split(t,".")
	//
	//header:=parts[0]
	//pl:=parts[1]
	//sign:=parts[2]
	log.Println("token", t)

	return true
}
