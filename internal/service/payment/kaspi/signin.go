package kaspi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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


