package kaspi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// senisitive should be in env
const (
	retrieverId          = "CvKgcikChiJ" //accepts any value, used for sms
	installId            = "770c19b2-07e0-48f3-8509-ceda41600139"
	platformVersion      = "11"
	appVersion           = "5.23"
	appBuild             = "502"
	deviceId             = "58e3ce032b573027"
	deviceBrand          = "Redmi"
	deviceModel          = "220333QAG"
	frontCameraAvailable = true
	remoteAddress        = "192.168.1.112" //accepts any local ip address
)

type Kaspi struct {
	installId            string
	retrieverId          string
	platformVersion      string
	appVersion           string
	appBuild             string
	deviceId             string
	deviceBrand          string
	deviceModel          string
	frontCameraAvailable bool
	remoteAddress        string
}

type Ticket *string

func NewKaspiDevice() Kaspi {
	return Kaspi{
		installId:            installId,
		retrieverId:          retrieverId,
		platformVersion:      platformVersion,
		appBuild:             appBuild,
		appVersion:           appVersion,
		deviceId:             deviceId,
		deviceBrand:          deviceBrand,
		deviceModel:          deviceModel,
		frontCameraAvailable: true,
		remoteAddress:        "192.168.1.111",
	}
}

func (k Kaspi) SignIn(phoneNumber string, password string) (Ticket, error) {

	body := []byte(fmt.Sprintf(`{
  "data": {
    "login": "%v",
    "password": "%v",
    "retrieverId": "%v" 
  },
  "deviceInfo": {
    "installId": "%v",
    "deviceId": "%v",
    "platformVersion": "%v",
    "appVersion": "%v",
    "appBuild": "%v",
    "deviceBrand": "%v",
    "deviceModel": "%v",
    "frontCameraAvailable": %v
  },
  "remoteAddress": "%v"
}`, phoneNumber, password, k.retrieverId, k.installId, k.deviceId, k.platformVersion, k.appVersion, k.appBuild, k.deviceBrand, k.deviceModel, k.frontCameraAvailable, k.remoteAddress))

	url := fmt.Sprintf("https://signin.kaspi.kz/sessions/api/v1/ExtSession/SignIn")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var object map[string]any
	if err := json.NewDecoder(res.Body).Decode(&object); err != nil {
		return nil, err
	}

	data, ok := object["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("No 'data' in json response")
	}

	ticket, ok := data["ssoTicket"].(string)
	if !ok {
		return nil, fmt.Errorf("No 'ssoTicket' in json response")
	}

	return &ticket, nil
}

type PaymentDetails struct {
	ID     string
	Name   string
	Amount int
	Date   string
}

// date format example: 24.06.2023
func (k Kaspi) Payments(beginDate string, endDate string, ticket Ticket) ([]PaymentDetails, error) {
	var paymentDetails []PaymentDetails
	url := fmt.Sprintf("https://mybank.kaspi.kz/bank/goldapi1/api/v1/Gold/GetGoldStatement/1?beginDate=%v&endDate=%v", beginDate, endDate)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ticket", *ticket)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	var object map[string]any
	if err := json.NewDecoder(res.Body).Decode(&object); err != nil {
		return nil, err
	}

	data, ok := object["ops"].([]any)
	if !ok {
		return nil, fmt.Errorf("No 'ops' in json response")
	}

	for _, v := range data {
		ops, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot get elements of 'ops' in json response")

		}

		id, ok := ops["t_dtls"].(string)
		if !ok {
			continue
		}

		name, ok := ops["dtls"].(string)
		if !ok {
			return nil, fmt.Errorf("No 'dtls' in json response")
		}

		amount, ok := ops["o_a"].(float64)
		if !ok {
			return nil, fmt.Errorf("No 'o_a' in json response")
		}

		date, ok := ops["op_d"].(string)
		if !ok {
			return nil, fmt.Errorf("No 'op_d' in json response")
		}

		paymentDetails = append(paymentDetails, PaymentDetails{Name: name, Amount: int(amount), Date: date, ID: id})

	}

	return paymentDetails, nil

}

func (k Kaspi) MatchPayments(matchId string, payments []PaymentDetails) []PaymentDetails {
	var matchPayments []PaymentDetails

	for _, v := range payments {
		id := strings.Split(v.ID, ":")[0]
		if id == matchId {
			matchPayments = append(matchPayments, v)

		}
	}

	return matchPayments
}


func (k Kaspi) MatchRevenue(matchId string, payments []PaymentDetails) int {
	var totalSum int

	for _, v := range payments {
		id := strings.Split(v.ID, ":")[0]
		if id == matchId {
      totalSum += v.Amount
		}
	}

	return totalSum
}

func (k Kaspi) MatchParticipantPayment(matchId string, participantId string, payments []PaymentDetails) *PaymentDetails {
	for _, v := range payments {
		id := fmt.Sprintf("%v:%v", matchId, participantId)
		if v.ID == id {
			return &v
		}
	}

	return nil
}

func (k Kaspi) SendCapital() error {
  return nil
}
