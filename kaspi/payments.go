package kaspi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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
