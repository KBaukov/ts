package amo

import (
	"bytes"
	"encoding/json"
	"github.com/KBaukov/ts/config"
	"github.com/KBaukov/ts/ent"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	refresh_token string
	access_token  string
	expire        int
	tokenType     string
	clientID      string
	ClientSecr    string
	accessUrl     string
	redirectUrl   string
	ownUser       string
	ownPass       string
	pipeLine      string
	//amoStages     interface{}
)

func init() {
	cfg := config.LoadConfig("config.json")
	AmoCfg := cfg.AmoCRMSettings
	clientID = AmoCfg.ClientId
	ClientSecr = AmoCfg.ClientSecret
	accessUrl = AmoCfg.AccessUrl
	redirectUrl = AmoCfg.RedirectUrl
	ownUser = AmoCfg.OwnerLogin
	ownPass = AmoCfg.OwnerPass
	pipeLine = AmoCfg.PipeLineId

	go RefrashTokenTicker()
	log.Println("#### Refresh Token Ticker start #####")
}

func GetCredential() {
	log.Println("#################### start Auth #############################")
	cc, _ := getCsrf()
	resp, c := mainAuth(cc)
	log.Println("RespCode: ", c)
	log.Println("auth_code: ", resp.AUTH_CODE)
	tResp, c := getAccessToken(resp.AUTH_CODE)
	access_token = tResp.ACCESS_T
	refresh_token = tResp.REFRESH_T
	expire = tResp.EXPIRE
	tokenType = tResp.T_TYPE
	log.Println("access_token: ", access_token)
	log.Println("refresh_token: ", refresh_token)
}

func getAccessToken(authCode string) (ent.AuthResponseBody, int) {

	var response ent.AuthResponseBody

	getFormUrl := "https://aperlik.amocrm.ru/oauth2/access_token"
	requestBody := ent.AuthRequestBody{clientID, ClientSecr, "authorization_code", authCode, redirectUrl}
	rbJson, _ := json.Marshal(requestBody)
	log.Println("requestBody: " + string(rbJson))

	var ff io.Reader
	ff = bytes.NewBuffer(rbJson)
	req, _ := http.NewRequest("POST", getFormUrl, ff)
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return response, respCode
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %v", err)
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error %v", err)
	}

	return response, 0

}

func mainAuth(cookies []*http.Cookie) (ent.AuthResp, int) {
	var (
		response ent.AuthResp
		sessId   string
		csrf     string
	)

	for _, c := range cookies {
		if c.Name == "session_id" {
			sessId = c.Value
		}
		if c.Name == "csrf_token" {
			csrf = c.Value
		}
	}

	log.Println("Csrf: ", csrf)
	log.Println("sessId: ", sessId)

	getFormUrl := "https://www.amocrm.ru/oauth2/authorize"
	requestBody := ent.MainAuthBody{ownUser, ownPass, csrf}
	rbJson, _ := json.Marshal(requestBody)
	log.Println("requestBody: " + string(rbJson))

	var ff io.Reader
	ff = bytes.NewBuffer(rbJson)
	req, _ := http.NewRequest("POST", getFormUrl, ff)
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return response, respCode
	}

	getFormUrl = "https://aperlik.amocrm.ru/ajax/v3/clients/" + clientID
	req, _ = http.NewRequest("GET", getFormUrl, nil)

	for _, c := range cookies {
		req.AddCookie(c)
	}
	//req.Header.Set("Cookie", "session_id="+sessionID+";")
	client = http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode = resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return response, respCode
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %v", err)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error %v", err)
	}

	return response, 0
}

func getCsrf() ([]*http.Cookie, int) {
	cookies := make([]*http.Cookie, 0)
	getFormUrl := "https://www.amocrm.ru/oauth?client_id=70b6d029-547c-4187-a01b-06a82a6e4553&mode=post_message"
	req, _ := http.NewRequest("GET", getFormUrl, nil)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return cookies, respCode
	}
	cookies = resp.Cookies()
	return cookies, 0
}

func CreateContact(name string, email string, phone string) (string, int) {
	sn := strings.Split(name, " ")
	fields := make([]ent.CustomFieldValue, 0)

	valPhone := ent.StrValue{phone, 4049, "WORK"}
	valuesPhone := make([]*ent.StrValue, 0)
	valuesPhone = append(valuesPhone, &valPhone)
	phoneField := ent.CustomFieldValue{8333, "Телефон", "PHONE", "multitext", valuesPhone}
	fields = append(fields, phoneField)

	valEmail := ent.StrValue{email, 4061, "WORK"}
	valuesEmail := make([]*ent.StrValue, 0)
	valuesEmail = append(valuesEmail, &valEmail)
	emailFiels := ent.CustomFieldValue{8335, "Email", "EMAIL", "multitext", valuesEmail}
	fields = append(fields, emailFiels)

	cont := ent.Contact{name, sn[0], sn[1], 0, fields}
	bb := make([]*ent.Contact, 0)
	bb = append(bb, &cont)

	rbJson, _ := json.Marshal(bb)
	log.Println("contactJson: " + string(rbJson))

	//contactCreateUrl := "https://aperlik.amocrm.ru/api/v4/contacts"
	//var ff io.Reader
	//ff = bytes.NewBuffer(rbJson)
	//req, _ := http.NewRequest("POST", contactCreateUrl, ff)
	//req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Authorization", "Bearer "+access_token)
	//client := http.Client{}
	//resp, err := client.Do(req)
	//if err != nil {
	//	log.Printf("Error %v", err)
	//}
	//defer resp.Body.Close()
	//respCode := resp.StatusCode
	//if respCode < 200 || respCode > 204 {
	//	eee, _ := ioutil.ReadAll(resp.Body)
	//	log.Printf("Error %v", string(eee))
	//	return "", respCode
	//}
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Printf("Error %v", err)
	//}

	//tt := string(body)
	//bg := strings.Index(tt, "{\"contacts\":[{\"id\":")
	//tt = tt[bg+19:]
	//end := strings.Index(tt, ",")
	//contId := tt[:end]
	//log.Printf("#### contactId: %v", tt)

	return string(rbJson), 0
}

func UpdateContact(name string, email string, phone string, contactId string) (string, int) {
	sn := strings.Split(name, " ")
	fields := make([]ent.CustomFieldValue, 0)

	valPhone := ent.StrValue{phone, 4049, "WORK"}
	valuesPhone := make([]*ent.StrValue, 0)
	valuesPhone = append(valuesPhone, &valPhone)
	phoneField := ent.CustomFieldValue{8333, "Телефон", "PHONE", "multitext", valuesPhone}
	fields = append(fields, phoneField)

	valEmail := ent.StrValue{email, 4061, "WORK"}
	valuesEmail := make([]*ent.StrValue, 0)
	valuesEmail = append(valuesEmail, &valEmail)
	emailFiels := ent.CustomFieldValue{8335, "Email", "EMAIL", "multitext", valuesEmail}
	fields = append(fields, emailFiels)

	cId, err := strconv.Atoi(contactId)
	if err != nil {
		log.Println("err:", err.Error())
	}

	cont := ent.UpdContact{cId, name, sn[0], sn[1], 0, fields}
	bb := make([]*ent.UpdContact, 0)
	bb = append(bb, &cont)

	rbJson, _ := json.Marshal(bb)
	log.Println("requestBody: " + string(rbJson))

	contactCreateUrl := "https://aperlik.amocrm.ru/api/v4/contacts"
	var ff io.Reader
	ff = bytes.NewBuffer(rbJson)
	req, _ := http.NewRequest("PATCH", contactCreateUrl, ff)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "", respCode
	}
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Printf("Error %v", err)
	//}
	//
	//tt := string(body)
	//bg := strings.Index(tt, "{\"contacts\":[{\"id\":")
	//tt = tt[bg+19:]
	//end := strings.Index(tt, ",")
	//contId := tt[:end]
	//log.Printf("#### contactId: %v", tt)

	return "OK", 0
}

func GetContacts(contId string) (string, int) {

	url := "https://aperlik.amocrm.ru/api/v4/contacts/" + contId
	//url := "https://aperlik.amocrm.ru/api/v4/leads"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "", respCode
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %v", err)
	}

	return string(body), 0
}

func LeadCreat(leadData ent.LeadData, utmData ent.UtmData, contactJson string) (string, string, int) {

	leadJson := "[ { \"name\": \"" + leadData.NAME + "\", \"price\": " + strconv.Itoa(leadData.PRICE) +
		", \"status_id\": " + leadData.STATUS + ", \"created_by\": 0, \"pipeline_id\": " + pipeLine + ", " +
		" \"custom_fields_values\":[" +
		"{ \"field_id\": 1109847, \"values\": [ {\"value\": \"" + leadData.COUNT + "\"} ] }," +
		"{ \"field_id\": 1109829, \"values\": [ {\"value\": \"" + leadData.ZONE + "\"} ] }," +
		"{ \"field_id\": 1109837, \"values\": [ {\"value\": \"" + leadData.ROW + "\"} ] }," +
		"{ \"field_id\": 1109835, \"values\": [ {\"value\": \"" + leadData.SEAT + "\"} ] }," +
		"{ \"field_id\": 1109845, \"values\": [ {\"value\": \"" + leadData.ORDER + "\"} ] }," +
		"{ \"field_id\": 1109849, \"values\": [ {\"value\": \"" + leadData.DESCRIPTION + "\"} ] }," +
		"{ \"field_id\": 1099035, \"values\": [ {\"value\": \"" + leadData.REFLINK + "\"} ] }," +
		"{ \"field_id\": 1109043, \"values\": [ {\"value\": \"" + strconv.Itoa(leadData.DAMOUNT) + "\"} ] }," +
		"{ \"field_id\": 8341, \"values\": [ {\"value\": \"" + utmData.UTM_CONTENT + "\"} ] }," +
		"{ \"field_id\": 8343, \"values\": [ {\"value\": \"" + utmData.UTM_MEDIUM + "\"} ] }," +
		"{ \"field_id\": 8345, \"values\": [ {\"value\": \"" + utmData.UTM_COMPAIGN + "\"} ] }," +
		"{ \"field_id\": 8347, \"values\": [ {\"value\": \"" + utmData.UTM_SOURCE + "\"} ] }," +
		"{ \"field_id\": 8349, \"values\": [ {\"value\": \"" + utmData.UTM_TERM + "\"} ] }," +
		"{ \"field_id\": 8351, \"values\": [ {\"value\": \"" + utmData.UTM_REFERRER + "\"} ] }" +
		"]," +
		"\"_embedded\":{ \"contacts\":" +
		contactJson +
		"}" +
		" } ]"

	log.Printf("leadJson %v", leadJson)

	var ff io.Reader
	ff = bytes.NewBuffer([]byte(leadJson))

	url := "https://aperlik.amocrm.ru/api/v4/leads/complex"
	req, _ := http.NewRequest("POST", url, ff)
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "", "", respCode
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %v", err)
	}
	tt := string(body)
	bg := strings.Index(tt, "\"id\":")
	ttl := tt[bg+5:]
	end := strings.Index(ttl, ",")
	leadId := ttl[:end]
	log.Printf("#### leadId: %v", leadId)

	bg = strings.Index(tt, "\"contact_id\":")
	ttc := tt[bg+13:]
	end = strings.Index(ttc, ",")
	contId := ttc[:end]
	log.Printf("#### contactId: %v", contId)

	return leadId, contId, respCode
}

func LeadUpdate(leadId string, leadData ent.LeadData) (string, int) {

	leadJson := "[ { \"id\":" + leadId + ", \"price\": " + strconv.Itoa(leadData.PRICE) + "," +
		" \"custom_fields_values\":[" +
		"{ \"field_id\": 1109847, \"values\": [ {\"value\": \"" + leadData.COUNT + "\"} ] }," +
		"{ \"field_id\": 1109829, \"values\": [ {\"value\": \"" + leadData.ZONE + "\"} ] }," +
		"{ \"field_id\": 1109837, \"values\": [ {\"value\": \"" + leadData.ROW + "\"} ] }," +
		"{ \"field_id\": 1109835, \"values\": [ {\"value\": \"" + leadData.SEAT + "\"} ] }," +
		"{ \"field_id\": 1109845, \"values\": [ {\"value\": \"" + leadData.ORDER + "\"} ] }," +
		"{ \"field_id\": 1109849, \"values\": [ {\"value\": \"" + leadData.DESCRIPTION + "\"} ] }," +
		"{ \"field_id\": 1109043, \"values\": [ {\"value\": \"" + strconv.Itoa(leadData.DAMOUNT) + "\"} ] }," +
		"{ \"field_id\": 1099035, \"values\": [ {\"value\": \"" + leadData.REFLINK + "\"} ] }" +
		"]" +
		" } ]"

	log.Printf("leadJson %v", leadJson)

	var ff io.Reader
	ff = bytes.NewBuffer([]byte(leadJson))

	url := "https://aperlik.amocrm.ru/api/v4/leads"
	req, _ := http.NewRequest("PATCH", url, ff)
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "string(eee)", respCode
	}

	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Printf("Error %v", err)
	//}

	return "OK", respCode
}

func LeadStatusUpdate(leadId string, leadStatusId string) (string, int) {

	leadJson := "[ { \"id\": " + leadId + ", \"status_id\": " + leadStatusId + ", \"updated_by\": 0 } ]"

	log.Printf("leadJson %v", leadJson)

	var ff io.Reader
	ff = bytes.NewBuffer([]byte(leadJson))

	url := "https://aperlik.amocrm.ru/api/v4/leads"
	req, _ := http.NewRequest("PATCH", url, ff)
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "string(eee)", respCode
	}

	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Printf("Error %v", err)
	//}

	return "OK", respCode
}

func assignContactUpdate(leadId string, contactId string) (string, int) {

	leadJson := "[ { \"id\": " + leadId + ", \"updated_by\": 0 }," +
		"\"_embedded\": [ { \"contacts\":[ { \"id\": " + contactId + "} ] } ]" +
		" ]"

	log.Printf("leadJson %v", leadJson)

	var ff io.Reader
	ff = bytes.NewBuffer([]byte(leadJson))

	url := "https://aperlik.amocrm.ru/api/v4/leads"
	req, _ := http.NewRequest("PATCH", url, ff)
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "string(eee)", respCode
	}

	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Printf("Error %v", err)
	//}

	return "OK", respCode
}

func LeadStatusTicketsUpdate(leadId string, leadStatusId string, ticketsNumbers string) (string, int) {

	leadJson := "[ { \"id\": " + leadId + ", \"status_id\": " + leadStatusId + ", \"updated_by\": 0," +
		" \"custom_fields_values\":[ { \"field_id\": 1109849, \"values\": [ {\"value\": \"Номера билетов: " + ticketsNumbers + "\"} ] } ]" +
		" } ]"

	log.Printf("leadJson %v", leadJson)

	var ff io.Reader
	ff = bytes.NewBuffer([]byte(leadJson))

	url := "https://aperlik.amocrm.ru/api/v4/leads"
	req, _ := http.NewRequest("PATCH", url, ff)
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "string(eee)", respCode
	}

	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Printf("Error %v", err)
	//}

	return "OK", respCode
}

func GetLead(leadId string) (string, int) {

	url := "https://aperlik.amocrm.ru/api/v4/leads/" + leadId
	//url := "https://aperlik.amocrm.ru/api/v4/leads"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "", respCode
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %v", err)
	}

	return string(body), 0
}

func RefrshToken() (string, int) {

	reqJson := "{\"client_id\": \"" + clientID + "\",\"client_secret\": \"" + ClientSecr + "\",\"grant_type\": \"refresh_token\"," +
		"\"refresh_token\": \"" + refresh_token + "\",\"redirect_uri\": \"" + redirectUrl + "\" }"

	log.Printf("refreshJson %v", reqJson)

	var ff io.Reader
	ff = bytes.NewBuffer([]byte(reqJson))

	url := "https://www.amocrm.ru/oauth2/access_token"
	req, _ := http.NewRequest("POST", url, ff)
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error %v", err)
	}
	defer resp.Body.Close()
	respCode := resp.StatusCode
	if respCode < 200 || respCode > 204 {
		eee, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Error %v", string(eee))
		return "string(eee)", respCode
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error %v", err)
	}

	var tData ent.AuthResponseBody

	err = json.Unmarshal(body, &tData)
	if err != nil {
		log.Printf("Error %v", err)
	}

	access_token = tData.ACCESS_T
	refresh_token = tData.REFRESH_T
	expire = tData.EXPIRE
	tokenType = tData.T_TYPE

	return "OK", respCode

}

func RefrashTokenTicker() {
	ticker := time.NewTicker(20 * time.Hour) //time.Minute)
	for now := range ticker.C {
		log.Println("Refrash token start", now)
		_, _ = RefrshToken()
		log.Println("Refrash token finished", now)
	}
}
