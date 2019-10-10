package mailing

import (
	"bytes"
	"html/template"
)

type OTPTemplateArgs struct {
	Recipient string `json:"recipient"`
	OTP       string `json:"otp"`
	OTPType   string `json:"otp_type"`
}

func OTPTemplate(args OTPTemplateArgs) (string, error) {
	var tpl bytes.Buffer
	tmpl, err := template.ParseFiles("templates/mailing/otp.html")
	if err != nil {
		return "", err
	}
	if err = tmpl.Execute(&tpl, args); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
