package handle

import (
	"encoding/base64"
	"fmt"
	"github.com/KBaukov/ts/config"
	"github.com/KBaukov/ts/db"
	"github.com/KBaukov/ts/ent"

	"encoding/json"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func init() {
	cfg := config.LoadConfig("config.json")
	WsConfig := cfg.WsConfig

	writeWait = time.Duration(WsConfig.WriteWait) * time.Second               // Time allowed to write a message to the peer.
	maxMessageSize = WsConfig.MaxMessageSize                                  // Maximum message size allowed from peer.
	pingWait = time.Duration(WsConfig.PingWait) * time.Second                 // Time allowed to read the next pong message from the peer.
	pongWait = time.Duration(WsConfig.PongWait) * time.Second                 // Time allowed to read the next pong message from the peer.
	pingPeriod = time.Duration(WsConfig.PingPeriod) * time.Second             // Send pings to peer with this period. Must be less than pongWait.
	closeGracePeriod = time.Duration(WsConfig.CloseGracePeriod) * time.Second // Time to wait before force close on connection.
	brPref = WsConfig.BrPref
	wsAllowedOrigin = WsConfig.WsAllowedOrigin

	//log.Println("Config values: ", cfg)

	go hub.run()
	log.Println("Config values: ", cfg)

	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbConnection.DBHost,
		cfg.DbConnection.DBPort,
		cfg.DbConnection.DBUser,
		cfg.DbConnection.DBPass,
		cfg.DbConnection.DBName,
	)

	dd, _ := db.NewDB(psqlconn)

	go inBackground(dd)
	log.Println("#### Clear expired reserved Ticker start #####")

	go CheckTicketsForSend(dd)
	log.Println("#### CheckTickets for sent Ticker start #####")
	//select {}

}

var (
	writeWait            time.Duration
	maxMessageSize       int64
	pingWait             time.Duration
	pongWait             time.Duration
	pingPeriod           time.Duration
	closeGracePeriod     time.Duration
	brPref               string
	wsAllowedOrigin      string
	isControlSessionOpen = false
	WsAsignConns         = make(map[string]string)
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		var user = ent.User{}
		//session := getSession(nil, r)
		//u := session.Values["user"]
		//if u != nil {
		//	user = u.(ent.User)
		//}
		deviceOrigin := r.Header.Get("Origin")
		token := r.Header.Get("Sec-WebSocket-Protocol")

		tt, err := f(user.PASS, 10, 16)
		if err != nil {
			log.Println("########## !!!!!!!!!!!!! ####################:")
		}

		if wsAllowedOrigin == deviceOrigin || token == tt {
			return true
		} else {
			log.Println("Origin not allowed:", deviceOrigin)
			return false
		}

	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.Println("{WS}:Error:", reason.Error())
	},
}

func ServeWs(db db.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var ws *websocket.Conn
		var err error

		devId := r.Header.Get("DeviceId")
		if devId == "" {
			suf, _ := HashPass(r.Header.Get("Sec-WebSocket-Key"))
			devId = brPref + strings.ToUpper(suf[12:18])
		}

		log.Println("incoming WS request from: ", devId)

		swp := r.Header.Get("Sec-WebSocket-Protocol")
		if swp != "" {
			headers := http.Header{"Sec-Websocket-Protocol": {swp}}
			ws, err = upgrader.Upgrade(w, r, headers)
		} else {
			ws, err = upgrader.Upgrade(w, r, nil)
		}

		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		//defer ws.Close() !!!! Important

		conn := &Conn{send: make(chan []byte, 256), ws: ws, deviceId: devId}
		hub.register <- conn
		go conn.writePump(db)
		conn.readPump(db)

		//WsConnections[deviceId] = ws
		//if(ws != nil) {
		//	//ws.SetPingHandler(ping)
		//	//ws.SetPongHandler(pong)
		//	log.Println("Create new Ws Connection: succes, device: ", deviceId)
		//	go wsProcessor(ws, db, deviceId)
		//} else {
		//	log.Println("Ws Connectionfor device: ", deviceId, " not created.")
		//}

	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Conn) readPump(db db.Database) {
	defer func() {
		hub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	devId := c.deviceId

	for {
		if c != nil {
			mt, message, err := c.ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Printf("error: %v", err)
				}
				break
			}

			log.Printf("[WS]:recive: %s, type: %d", message, mt)
			msg := string(message)

			if strings.Contains(msg, "\"action\":\"connect\"") {
				if !sendMsg(c, "{\"action\":\"connect\",\"success\":true, \"device\":\""+devId+"\" }") {
					break
				} else {
					log.Println("{action:connect,success:true}", c.deviceId)
				}
			}

			if strings.Contains(msg, "\"action\":\"assign\"") {
				assign := msg[29 : len(msg)-2]
				cc := hub.getConnByDevId(assign)
				if cc != nil {
					WsAsignConns[devId] = assign
					log.Println("#### Assign", assign, " to ", devId)
					if !sendMsg(c, "{\"action\":\"assign\",\"success\":true}") {
						break
					}
				} else {
					if !sendMsg(c, "{\"action\":\"assign\",\"success\":false}") {
						break
					}
				}

			}

			if strings.Contains(msg, "\"action\":\"resend\"") {
				var rMsg ent.ResendMessage

				err = json.Unmarshal([]byte(msg), &rMsg)
				if err != nil {
					log.Println("Error data unmarshaling: ", err)
				}

				rConn := hub.getConnByDevId(rMsg.RECIPIENT)
				sender := devId

				message, _ := base64.URLEncoding.DecodeString(rMsg.MESSAGE)
				mm := strings.Replace(string(message), "\"sender\":\"\"", "\"sender\":\""+sender+"\"", 1)
				log.Println("[WS]:resend message: ", mm, " to: ", rMsg.RECIPIENT)
				sendMsg(rConn, mm)
			}

			//###############################################################

			if strings.Contains(msg, "\"action\":\"reserve\"") {
				var rMsg ent.ReservSeatMsg
				var ansver string
				var st string

				err = json.Unmarshal([]byte(msg), &rMsg)
				if err != nil {
					log.Println("Error data unmarshaling: ", err)
				}

				isVacant, err := db.CheckSeatStatess(rMsg.SEAT_ID, 0)
				if err != nil {
					log.Println("Error while reserved seat: ", err)
				}
				if isVacant {

					isSuccess, err := db.SetSeatStatess(rMsg.SEAT_ID, 1)
					if err != nil {
						log.Println("Error while reserved seat: ", err)
					}
					if isSuccess {
						isSuccess, err = db.ReserveSeat(rMsg.SEAT_ID)
						if err != nil {
							log.Println("Error while reserved seat: ", err)
						}
						if isSuccess {
							ansver = "true"
							st = "1"
						} else {
							ansver = "false"
							st = "0"
						}

					} else {
						ansver = "false"
						st = "0"
					}

				} else {
					ansver = "false"
				}
				if !sendMsg(c, "{\"action\":\"reserve\",\"success\":"+ansver+", \"seat\":\""+rMsg.SEAT_ID+"\"}") {
					break
				} else {
					if st == "1" {
						hub.sendDataToWeb("{\"action\":\"sysreserve\",\"seat\":\""+rMsg.SEAT_ID+"\", \"state\": "+st+"}", "TS_system", c)
					}

				}

			}

			if strings.Contains(msg, "\"action\":\"unreserve\"") {
				var rMsg ent.ReservSeatMsg
				var ansver string

				err = json.Unmarshal([]byte(msg), &rMsg)
				if err != nil {
					log.Println("Error data unmarshaling: ", err)
				}

				isSuccess, err := db.SetSeatStatess(rMsg.SEAT_ID, 0)
				if err != nil {
					log.Println("Error while unreserved seat: ", err)
				}
				if isSuccess {
					isSuccess, err = db.UnReserveSeat(rMsg.SEAT_ID)
					if err != nil {
						log.Println("Error while unreserved seat: ", err)
					}
					if isSuccess {
						ansver = "true"
					} else {
						ansver = "false"
					}

				} else {
					ansver = "false"
				}
				if !sendMsg(c, "{\"action\":\"unreserve\",\"success\":"+ansver+"}") {
					break
				}

				hub.sendDataToWeb("{\"action\":\"sysreserve\",\"seat\":\""+rMsg.SEAT_ID+"\", \"state\": 0}", "TS_system", c)
			}

		} else {
			log.Println("# Connection lost    #")
			hub.unregister <- c
			break
		}

	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Conn) writePump(db db.Database) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		if c != nil {
			select {
			case message, ok := <-c.send:
				if !ok {
					// The hub closed the channel.
					c.write(websocket.CloseMessage, []byte{})
					return
				}

				c.ws.SetWriteDeadline(time.Now().Add(writeWait))
				w, err := c.ws.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)
				// Add queued chat messages to the current websocket message.
				n := len(c.send)
				for i := 0; i < n; i++ {
					//w.Write(newline)
					w.Write(<-c.send)
				}

				if err := w.Close(); err != nil {
					return
				}
			case <-ticker.C:
				//log.Println("############################ Ping ##############################")

				if err := c.write(websocket.PingMessage, []byte{}); err != nil {
					log.Println("# Send ping to " + c.GetDeviceId() + " failed.  #")
					return
				} else {
					log.Println("# Send ping to " + c.GetDeviceId() + " success.  #")
				}

			}
		} else {
			hub.unregister <- c
			log.Println("# Connection lost    #")
		}
	}
}

func sendMsg(c *Conn, m string) bool {
	if c != nil {
		c.send <- []byte(m)
		log.Println("[WS]:send:", m, " succes")
		return true
	} else {
		log.Println("[WS]:send:", m, " failed")
		return false
	}

}

//func broadCastSend(msg string) bool {
//	hub.
//}

func unAssign(devId string) {
	for key, val := range WsAsignConns {
		if val == devId {
			conn := hub.getConnByDevId(key)
			if conn != nil {
				hub.getConnByDevId(key).send <- []byte("{\"action\":\"unassign\",\"device\":\"" + val + "\"}")
			}
			delete(WsAsignConns, key)
		}
	}
}

func inBackground(db db.Database) {
	ticker := time.NewTicker(30 * time.Second) //time.Minute)
	for now := range ticker.C {
		stt, err := db.ClearExpiredReserves()
		if err != nil {
			log.Println("Error while clear status for unresed seats:  = %v : time: %v", err, now)
		}
		var n = len(stt)
		if n > 0 {
			var tt string
			for _, ss := range stt {
				tt += ",{\"seat\":\"" + ss + "\", \"state\": 0 }"
			}
			tt, err = f(tt, 1, len(tt)) //[1:len(tt)]
			if err != nil {
				log.Println("########## !!!!!!!!!!!!! ####################:")
			}
			hub.sendDataToWeb("{\"action\":\"sysunreserves\",\"seats\":[ "+tt+" ]}", "TS_system", nil)
		}
		log.Printf("#### %v Clear expired reserved success #####", n)
	}
}

func CheckTicketsForSend(db db.Database) {
	ticker := time.NewTicker(time.Minute) //time.Minute)
	for now := range ticker.C {
		_, err := db.SendTickets()
		if err != nil {
			log.Println("Error while send tickets:  = %v : time: %v", err, now)
		}
	}
}

func f(tt string, begin int, end int) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Catch Panic in ", r)
		}
	}()
	ss := tt[begin:end]

	return ss, nil
}
