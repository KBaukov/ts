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
	//gob.Register(gocloak.JWT{})
	//gob.Register( "" )
	//gob.Register(ent.User{})

	//sessStore.Options = &sessions.Options{
	//	Domain:   "*",
	//	Path:     "/",
	//	MaxAge:   3600 * 8, // 8 hours
	//	HttpOnly: false,
	//}
	//
	//client = gocloak.NewClient(configURL)
	//ctx = context.Background()
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

		if r.URL.Path == "/api/seatstates" && r.Method == "GET" {

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
				hub.sendDataToWeb(msg, "TS_system")
			}
			apiDataResponse(w, []int{}, err)
			return
		}

		if r.URL.Path == "/api/messagesend" {

			message := r.FormValue("msg")

			hub.sendDataToWeb(message, "TS_system")

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
			pData := ent.PayData{"test_api_00000000000000000000002", "Fortune 2050: Оплата билетов", amount,
				"RUB", email, orderNumber, email, "mini", 3, ext}

			apiDataResponse(w, pData, nil)
			return
		}

		//if r.URL.Path == "/api/registry/add" {
		//	var (
		//		err    error
		//		dept   int
		//		selfId int
		//	)
		//
		//	pKey := r.PostFormValue("parent_key");
		//	env := r.FormValue("env");
		//
		//	lev := r.PostFormValue("dept");
		//	if lev != "" && lev != "nan" {
		//		dept, err = strconv.Atoi(lev)
		//		if err != nil {
		//			log.Println("err:", err.Error())
		//		}
		//	}
		//
		//	value := r.PostFormValue("value")
		//	key := r.PostFormValue("key")
		//	description := r.PostFormValue("description")
		//	ext := r.PostFormValue("ext")
		//	owner := r.PostFormValue("owner")
		//	sId := r.PostFormValue("self_id")
		//	if sId != "" && sId != "nan" {
		//		selfId, err = strconv.Atoi(sId)
		//		if err != nil {
		//			log.Println("err:", err.Error())
		//		}
		//	}
		//
		//	id, err := db.RegistryNodeCreate(pKey, selfId, key, value, dept, description, ext, env, owner, user);
		//	if err != nil {
		//		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		//		log.Printf("Ошибка создания ноды реестра (parent_key: %v, key: %v,value: %v, dept: %i, description: %v, ext: %v): %v", pKey, key, value, dept, description, ext, err)
		//	}
		//
		//	apiDataResponse(w, id, err)
		//	return;
		//}
		//
		//if r.URL.Path == "/api/registry/update" {
		//	var (
		//		selId int
		//		err   error
		//		dept  int
		//	)
		//
		//	pKey := r.PostFormValue("parent_key");
		//	env := r.FormValue("env");
		//	sId := r.PostFormValue("self_id");
		//	if sId != "" && sId != "nan" {
		//		selId, err = strconv.Atoi(sId)
		//		if err != nil {
		//			log.Println("err:", err.Error())
		//		}
		//	}
		//	lev := r.PostFormValue("dept");
		//	if lev != "" && lev != "nan" {
		//		dept, err = strconv.Atoi(lev)
		//		if err != nil {
		//			log.Println("err:", err.Error())
		//		}
		//	}
		//	value := r.PostFormValue("value")
		//	key := r.PostFormValue("key")
		//	oldKey := r.PostFormValue("old_key")
		//	description := r.PostFormValue("description")
		//	ext := r.PostFormValue("ext")
		//	owner := r.PostFormValue("owner")
		//
		//	err = db.RegistryNodeUpdate(pKey, oldKey, selId, key, value, dept, description, ext, env, owner, user);
		//	if err != nil {
		//		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		//		log.Printf("Ошибка изменения ноды реестра (old_key: %v,key: %v,value: %v, dept: %i, description: %v, ext: %v): %v", oldKey, key, value, dept, description, ext, err)
		//	}
		//
		//	apiDataResponse(w, []string{}, err)
		//	return;
		//}
		//
		//if r.URL.Path == "/api/registry/delete" {
		//	var id int;
		//	var err error;
		//	key := strings.ToLower(r.FormValue("key"));
		//	env := strings.ToLower(r.FormValue("env"));
		//
		//	err = db.RegistryNodeDelete(key, env, user);
		//	if err != nil {
		//		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		//		log.Printf("Ошибка elfktybz ноды реестра (id: %i): %v", id, err)
		//	}
		//
		//	apiDataResponse(w, []string{}, err)
		//	return;
		//}
		//
		//if r.URL.Path == "/api/dict" {
		//
		//	key := strings.ToLower(r.FormValue("key"))
		//	env := strings.ToLower(r.FormValue("env"));
		//
		//	dict, err := db.GetDictionary(key, env)
		//	if err != nil {
		//		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		//		log.Printf("Ошибка доступа у справочнику (key: %v): %v", key, err)
		//	}
		//
		//	apiDataResponse(w, dict, err)
		//	return;
		//}
		//if r.URL.Path == "/api/node/diff" {
		//
		//	pKey := r.FormValue("parent_key");
		//	env := r.FormValue("env");
		//	registry, err := db.GetTreeNodeDiff(pKey, env)
		//	if err != nil {
		//		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		//		log.Printf("Ошибка получения реестра (parKey: %v): %v", pKey, err)
		//	}
		//	apiDataResponse(w, registry, err)
		//	return;
		//}
		//if r.URL.Path == "/api/node/history" {
		//
		//	key := r.FormValue("key");
		//	env := r.FormValue("env");
		//	history, err := db.GetChangeHistory(key, env);
		//	if err != nil {
		//		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		//		log.Printf("Ошибка получения реестра (parKey: %v): %v", key, err)
		//	}
		//	apiDataResponse(w, history, err)
		//	return;
		//}
		////###################
		//if r.URL.Path == "/api/users" {
		//	users, err := db.GetUsers()
		//	apiDataResponse(w, users, err)
		//}
		//if r.URL.Path == "/api/user/edit" {
		//	id := r.PostFormValue("id")
		//	intId, err := strconv.Atoi(id)
		//	if err != nil {
		//		log.Println("err:", err.Error())
		//	}
		//
		//	login := r.PostFormValue("login")
		//	pass := r.PostFormValue("pass")
		//	pass, err = HashPass(pass)
		//	if err != nil {
		//		log.Println("Ошибка хеширования", err)
		//	}
		//	usrType := r.PostFormValue("user_type")
		//	actFlag := r.PostFormValue("active_flag")
		//	lastVs := r.PostFormValue("last_visit")
		//	lastV, err := time.Parse("2006-01-02T00:00:00Z", lastVs)
		//	if err != nil {
		//		log.Println("date forma validation error:", err.Error())
		//	}
		//
		//	_, err = db.UpdUser(intId, login, pass, usrType, actFlag, lastV)
		//	apiDataResponse(w, []int{}, err)
		//}
		//if r.URL.Path == "/api/user/delete" {
		//	id := r.PostFormValue("id")
		//	intId, err := strconv.Atoi(id)
		//	if err != nil {
		//		log.Println("err:", err.Error())
		//	}
		//
		//	_, err = db.DelUser(intId)
		//	apiDataResponse(w, []int{}, err)
		//}

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

//########################## helpers ############################

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
