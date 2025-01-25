package apiclient

import "encoding/base64"

func basicAuthorization(username, password string) string {
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return "Basic " + credentials
}
