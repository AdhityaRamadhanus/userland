package contextkey

const (
	//ClientID is context key to get http request client identifcation
	AccessToken    = iota
	AccessTokenKey = iota

	HandlerName         = iota
	CurrentHandlerState = iota
)
