package mailing

import (
	"os"
	"time"

	"github.com/sarulabs/di"
)

var (
	ClientBuilder = di.Def{
		Name:  "mailing-client",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			mailingClient := NewMailingClient(
				os.Getenv("USERLAND_MAIL_HOST"),
				WithClientTimeout(time.Second*5),
				WithBasicAuth(os.Getenv("MAIL_SERVICE_BASIC_USER"), os.Getenv("MAIL_SERVICE_BASIC_PASS")),
			)

			return mailingClient, nil
		},
	}
)
