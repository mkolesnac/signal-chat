package apitypes

const prefix = "/v1"

const (
	EndpointSignUp        = prefix + "/signup"
	EndpointSignIn        = prefix + "/signin"
	EndpointSignOut       = prefix + "/signout"
	EndpointConversations = prefix + "/conversations"
	EndpointMessages      = prefix + "/messages"
	EndpointUsers         = prefix + "/users"
	EndpointUser          = prefix + "/users/:id"
	EndpointPreKeyBundle  = prefix + "/prekeys/:id"
)
