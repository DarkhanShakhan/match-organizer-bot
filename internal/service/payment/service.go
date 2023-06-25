package payment

import "github.com/DarkhanShakhan/telegram-bot-template/internal/service/payment/kaspi"

const (
	username = "usernmae"
	password = "password"
)

type Service interface {
	GetPayments(beginDate, endDate string) ([]kaspi.PaymentDetails, error)
	GetPaymentsByMatchID(beginDate, endDate string, matchId string) ([]kaspi.PaymentDetails, error)
	GetUserPayment(beginDate, endDate string, userID, matchID string) (*kaspi.PaymentDetails, error)
	MakePayment(phoneNumber string, amount int) error
}

type service struct {
}

func New() Service {
	return &service{}
}

func (s *service) GetPayments(beginDate, endDate string) ([]kaspi.PaymentDetails, error) {
	bank := kaspi.NewKaspiDevice()
	ticket, err := bank.SignIn(username, password)
	if err != nil {
		return nil, err
	}

	payments, err := bank.Payments(beginDate, endDate, ticket)
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (s *service) GetPaymentsByMatchID(beginDate, endDate string, matchId string) ([]kaspi.PaymentDetails, error) {
	bank := kaspi.NewKaspiDevice()
	ticket, err := bank.SignIn(username, password)
	if err != nil {
		return nil, err
	}

	payments, err := bank.Payments(beginDate, endDate, ticket)
	if err != nil {
		return nil, err
	}

	matchPayments := bank.MatchPayments(matchId, payments)

	return matchPayments, nil
}

func (s *service) GetUserPayment(beginDate, endDate string, userID, matchID string) (*kaspi.PaymentDetails, error) {
	bank := kaspi.NewKaspiDevice()
	ticket, err := bank.SignIn(username, password)
	if err != nil {
		return nil, err
	}

	payments, err := bank.Payments(beginDate, endDate, ticket)
	if err != nil {
		return nil, err
	}

	matchPayments := bank.MatchParticipantPayment(matchID, userID, payments)

	return matchPayments, nil

}

func (s *service) MakePayment(phoneNumber string, amount int) error {
	bank := kaspi.NewKaspiDevice()
	ticket, err := bank.SignIn(username, password)
	if err != nil {
		return err
	}

	transaction := kaspi.NewTransaction(ticket, phoneNumber, amount)
	err = transaction.Make()
	if err != nil {
		return err
	}

  return nil
}
