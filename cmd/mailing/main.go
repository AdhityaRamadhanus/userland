package main

import (
	"fmt"

	"github.com/joho/godotenv"

	"github.com/AdhityaRamadhanus/userland/service/mailing"
)

func main() {
	godotenv.Load()

	str, err := mailing.OTPTemplate(mailing.OTPTemplateArgs{
		Recipient: "test aja",
		OTP:       "454545",
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	// mailjetClient := mailjet.NewMailjetClient(os.Getenv("MJ_APIKEY_PUBLIC"), os.Getenv("MJ_APIKEY_PRIVATE"))
	// template, _ := ioutil.ReadFile("mailjet/templates/otp.html")
	// messagesInfo := []mailjet.InfoMessagesV31{
	// 	mailjet.InfoMessagesV31{
	// 		From: &mailjet.RecipientV31{
	// 			Email: "adhitya.ramadhanus@gmail.com",
	// 			Name:  "Mailjet Pilot",
	// 		},
	// 		To: &mailjet.RecipientsV31{
	// 			mailjet.RecipientV31{
	// 				Email: "adhitya.ramadhanus@icehousecorp.com",
	// 				Name:  "passenger 1",
	// 			},
	// 		},
	// 		Subject:  "Your email flight plan!",
	// 		HTMLPart: string(template),
	// 	},
	// }
	// messages := mailjet.MessagesV31{Info: messagesInfo}
	// res, err := mailjetClient.SendMailV31(&messages)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Data: %+v\n", res)
}
