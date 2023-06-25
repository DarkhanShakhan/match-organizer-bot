package main

import (
	"fmt"

	"github.com/complexusprada/kaspi_perevod/kaspi"
)

// "fmt"

// "github.com/complexusprada/kaspi_perevod/kaspi"

func main() {
	device := kaspi.NewKaspiDevice()
	ticket, err := device.SignIn("username", "password")
	if err != nil {
		fmt.Println(err)
		return
	}

  
	transaction := kaspi.NewTransaction(ticket, "7475849384", 100)
  err  = transaction.Make()
  if err != nil {
		fmt.Println(err)
    return
  }

  fmt.Println("success")

	// fmt.Println(fio)

	// payments, err := kaspi.Payments("23.06.2023", "25.06.2023", ticket)
	// if err != nil {
	// 	fmt.Println(err)
	//    return
	// }
	//
	//  matchPayments := kaspi.MatchPayments("178", payments)
	//  participant := kaspi.MatchParticipantPayment("178", "2873", payments)
	//
	//
	// fmt.Println(matchPayments)
	// fmt.Println(*participant)
}
