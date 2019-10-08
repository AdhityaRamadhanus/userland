package security

import "time"

var (
	UserAccessTokenExpiration   = time.Second * 60 * 10 // 10 minutes
	TFATokenExpiration          = time.Second * 60 * 2  // 2 minutes
	ForgotPassExpiration        = time.Second * 60 * 5  // 5 minutes
	EmailVerificationExpiration = time.Second * 60 * 2  // 2 minutes
)
