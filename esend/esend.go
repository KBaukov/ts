package esend

import (
	"fmt"
	"github.com/KBaukov/ts/ent"
	"log"
	"net/http"
	"net/url"
)

var (
	apiUrl   = "http://localhost:9010"
	resource = "/"
)

func init() {

}

type errData struct {
	Error_Code    int
	Error_Message string
}

func SendTickets(data []*ent.TicketForSend) ([]*ent.SendTicketStatus, error) {

	sts := make([]*ent.SendTicketStatus, 0)
	for i, tData := range data {

		log.Printf("%v Pepare data for Ticket Send %v: ticketNumber: ", i, tData.ORDER_NUMBER)
		//req, err := http.NewRequest("GET", "http://localhost:9010/", nil)

		params := url.Values{}
		params.Add("Email", tData.EMAIL)
		params.Add("Customer", tData.NAME)
		params.Add("TicketNumber", tData.TICKET_NUMBER)
		params.Add("OrderDetail", tData.ORDER_NUMBER+" от "+tData.ORDER_DATE)
		params.Add("Section", tData.SECTOR)
		params.Add("Row", tData.ROW)
		params.Add("Seat", tData.SEAT)

		fmt.Println("URL:>", apiUrl+resource)

		u, _ := url.ParseRequestURI(apiUrl)
		u.Path = resource
		u.RawQuery = params.Encode()
		urlStr := fmt.Sprintf("%v", u)

		client := &http.Client{}
		r, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			log.Printf("Http request error: %v", r)
			return nil, err
		}
		r.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(r)
		//fmt.Println(resp.Status)
		if err != nil {
			log.Printf("Http request error: %v", resp)
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			log.Printf("%v Ticket success Sended %v: ticketNumber: ", i, tData.TICKET_NUMBER)
			ts := ent.SendTicketStatus{tData.ORDER_NUMBER, tData.TICKET_NUMBER, "OK"}
			sts = append(sts, &ts)
			//db.OrderLog("ticketSend", "ticketSendSucces", tData.ORDER_NUMBER, "0", "Билет успешно отправлен: "+ticketNumber, "true", "")
		} else {
			log.Printf("%v Ticket Not Sended %v: ticketNumber: ", i, tData.TICKET_NUMBER)
			ts := ent.SendTicketStatus{tData.ORDER_NUMBER, tData.TICKET_NUMBER, "OK"}
			sts = append(sts, &ts)
			//db.OrderLog("ticketSend", "ticketSendNoSucces", tData.ORDER_NUMBER, "-1", "Билет не отправлен: "+ticketNumber, "false", resp.Status)
		}
	}

	return sts, nil

}
