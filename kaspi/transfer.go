package kaspi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
)

// requisite-input-methods: history-requisite, pick-contact, manual-phone
const (
	sourceAccountType    = "own-kaspi-gold"
	targetAccountType    = "ext-kaspi-gold"
	requisiteInputMethod = "manual-phone"
	currency             = "KZT"
	sourceAccount        = "A_KZ65722C000026118468"
	feeAmount            = 0
)

type Transaction struct {
	Currency             string
	SourceAccountType    string
	TargetAccountType    string
	SourceAccount        string
	Amount               int
	PhoneNumber          string
	Ticket               Ticket
	FeeAmount            int
	RequisiteInputMethod string
}

func NewTransaction(ticket Ticket, phoneNumber string, amount int) Transaction {
	return Transaction{
		Currency:             currency,
		SourceAccountType:    sourceAccountType,
		TargetAccountType:    targetAccountType,
		SourceAccount:        sourceAccount,
		PhoneNumber:          phoneNumber,
		Amount:               amount,
		Ticket:               ticket,
		FeeAmount:            feeAmount,
		RequisiteInputMethod: requisiteInputMethod,
	}
}

type transferId *string

func (t Transaction) Make() error {
	fio, err := t.getTargetFio()
	if err != nil {
		return err
	}

	id := t.register(fio)
	if id == nil {
		return errors.New("invalid transferId")
	}

	_, err = t.process(id)
  if err != nil {
    return  err
  }

  return nil
}

func (t Transaction) process(id transferId) (bool, error) {
	body := []byte(fmt.Sprintf(`{
  "transferId": "%v",
  "requestParams": {}
} `, *id))

	url := fmt.Sprintf("https://transfers.kaspi.kz/api/kaspi-client/process")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", *t.Ticket)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	var object map[string]any
	if err := json.NewDecoder(res.Body).Decode(&object); err != nil {
		return false, err
	}

	data, ok := object["data"].(map[string]any)
	if !ok {
		return false, fmt.Errorf("No 'data' in json response")
	}

	status, ok := data["status"].(string)
	if !ok {
		return false, fmt.Errorf("No 'status' in json response")
	}

	if status != "accepted" {
		return false, errors.New("transaction failed")
	}

	return true, nil

}

func (t Transaction) register(targetFio string) transferId {
	body := []byte(fmt.Sprintf(`{
  {
  "sourceAccount": {
    "productId": "%v",
    "type": "own-kaspi-gold",
    "currency": "KZT"
  },
  "targetAccount": {
    "type": "ext-kaspi-gold",
    "currency": "KZT",
    "phoneNumber": "%v",
    "cardHolderName": "%v"
  },
  "transferAmount": %v,
  "feeAmount": 0,
  "requisiteInputMethod": "manual-phone",
  "requestParams": {}
} }`, sourceAccount, t.PhoneNumber, targetFio, t.Amount))

	url := fmt.Sprintf("https://transfers.kaspi.kz/api/kaspi-client/register")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", *t.Ticket)


	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}

  dump, _ := httputil.DumpResponse(res, true)
  fmt.Println(string(dump))

	var object map[string]any
	if err := json.NewDecoder(res.Body).Decode(&object); err != nil {
		return nil
	}

	data, ok := object["data"].(map[string]any)
	if !ok {
		return nil
	}

	fio, ok := data["transferId"].(string)
	if !ok {
		return nil
	}

	return &fio
}

// {"error":{"type":"VALIDATION","title":"По номеру телефона не найден клиент. Укажите номер карты"}}
func (t Transaction) getTargetFio() (string, error) {
	body := []byte(fmt.Sprintf(`{
  "phoneNumber": "%v",
  "requisiteInputMethod": "%v" }`, t.PhoneNumber, requisiteInputMethod))

	url := fmt.Sprintf("https://transfers.kaspi.kz/api/kaspi-client/ext-kaspi-gold/validate")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", *t.Ticket)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var object map[string]any
	if err := json.NewDecoder(res.Body).Decode(&object); err != nil {
		return "", err
	}

	data, ok := object["data"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("No 'data' in json response")
	}

	fio, ok := data["fio"].(string)
	if !ok {
		return "", fmt.Errorf("No 'fio' in json response")
	}

	return fio, nil
}
