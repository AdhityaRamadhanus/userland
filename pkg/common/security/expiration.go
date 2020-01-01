package security

import "time"

var (
	RefreshAccessTokenExpiration = time.Second * 60 * 60 * 24 // one day
	UserAccessTokenExpiration    = time.Second * 60 * 10      // 10 minutes
	TFATokenExpiration           = time.Second * 60 * 2       // 2 minutes
	ForgotPassExpiration         = time.Second * 60 * 5       // 5 minutes
	EmailVerificationExpiration  = time.Second * 60 * 2       // 2 minutes
)
