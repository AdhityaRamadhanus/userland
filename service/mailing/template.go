package mailing

import (
	"bytes"
	"html/template"
)

type TemplateGenerator func(args map[string]interface{}) (string, error)

func OTPTemplate(args map[string]interface{}) (string, error) {
	var tpl bytes.Buffer
	tmpl, err := template.ParseFiles("templates/mailing/otp.html")
	if err != nil {
		return "", err
	}

	otpTemplateArgs := struct {
		Recipient string `json:"recipient"`
		OTP       string `json:"otp"`
		OTPType   string `json:"otp_type"`
	}{
		Recipient: args["recipient"].(string),
		OTP:       args["otp"].(string),
		OTPType:   args["otp_type"].(string),
	}
	if err = tmpl.Execute(&tpl, otpTemplateArgs); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
