package mailing

import (
	mailjet "github.com/mailjet/mailjet-apiv3-go"
)

type From struct {
	Name    string
	Address string
}

type To struct {
	Name    string
	Address string
}

type SendEmailOption struct {
	From    From
	Tos     []To
	Subject string
	Content string
}

type Service interface {
	SendEmail(opt SendEmailOption) error
}

type service struct {
	mailjetClient *mailjet.Client
}

func NewMailingService(client *mailjet.Client) Service {
	return &service{
		mailjetClient: client,
	}
}

func (s *service) SendEmail(sendOption SendEmailOption) error {
	recipients := mailjet.RecipientsV31{}

	for _, recipient := range sendOption.Tos {
		recipients = append(recipients, mailjet.RecipientV31{
			Email: recipient.Address,
			Name:  recipient.Name,
		})
	}
	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: sendOption.From.Address,
				Name:  sendOption.From.Name,
			},
			To:       &recipients,
			Subject:  sendOption.Subject,
			HTMLPart: sendOption.Content,
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.mailjetClient.SendMailV31(&messages)
	return err
}
