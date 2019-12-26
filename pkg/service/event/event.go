package event

var (
	LoginEvent              = "user.authentication.login"
	ForgotPasswordEvent     = "user.authentication.forgot_password"
	ChangeEmailRequestEvent = "user.profile.change_email_request"
	ChangeEmailEvent        = "user.profile.change_email"
	ChangePasswordEvent     = "user.profile.change_password"
	EnableTFAEvent          = "user.profile.enable_tfa"
	DisableTFAEvent         = "user.profile.disable_tfa"
)
