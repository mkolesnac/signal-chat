package api

const prefix = "/v1"

const (
	EndpointSignUp        = prefix + "/signup"
	EndpointSignIn        = prefix + "/signin"
	EndpointSignOut       = prefix + "/signout"
	EndpointConversations = prefix + "/conversations"
	EndpointMessages      = prefix + "/messages"
	EndpointParticipants  = prefix + "/participants"
)

func EndpointUser(userId string) string {
	return prefix + "/user/" + userId
}

func EndpointUserKeys(userId string) string {
	return prefix + "/user/" + userId + "/keys"
}
