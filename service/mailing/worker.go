package mailing

import (
	"encoding/json"
	"errors"

	"github.com/gocraft/work"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
)

type Worker struct {
	mailjetClient *mailjet.Client
}

var (
	templateMap = map[string]struct {
		TemplateGenerator TemplateGenerator
	}{
		"otp": {
			TemplateGenerator: OTPTemplate,
		},
	}
)

func NewWorker(mailjetClient *mailjet.Client) *Worker {
	return &Worker{
		mailjetClient: mailjetClient,
	}
}

func (w *Worker) EnquiryJob(job *work.Job) error {
	// build recipients
	sendOption := SendEmailOption{}
	jsonBytes, _ := json.Marshal(job.Args)
	json.Unmarshal(jsonBytes, &sendOption)

	recipients := mailjet.RecipientsV31{}

	for _, recipient := range sendOption.To {
		recipients = append(recipients, mailjet.RecipientV31{
			Email: recipient.Address,
			Name:  recipient.Name,
		})
	}

	// generate content
	template, templatePresent := templateMap[sendOption.Template]
	if !templatePresent {
		return errors.New("Tempalte not found")
	}

	content, _ := template.TemplateGenerator(sendOption.TemplateArgs)
	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: sendOption.From.Address,
				Name:  sendOption.From.Name,
			},
			To:       &recipients,
			Subject:  sendOption.Subject,
			HTMLPart: content,
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := w.mailjetClient.SendMailV31(&messages)
	return err
}
