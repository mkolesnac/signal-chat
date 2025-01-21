package database

import (
	"fmt"
)

type PrimaryKey string

func PublicIdentityKeyPK() PrimaryKey {
	return PrimaryKey("identityKey#public")
}

func PrivateIdentityKeyPK() PrimaryKey {
	return PrimaryKey("identityKey#private")
}

func SignedPreKeyPK(id string) PrimaryKey {
	return PrimaryKey(fmt.Sprintf("signedPreKey#%s", id))
}

func PreKeyPK(id string) PrimaryKey {
	return PrimaryKey(fmt.Sprintf("preKey#%s", id))
}

func ConversationPK(id string) PrimaryKey { return PrimaryKey(fmt.Sprintf("conversation#%s", id)) }

func MessagePK(conversationID, messageID string) PrimaryKey {
	return PrimaryKey(fmt.Sprintf("message#%s:%s", conversationID, messageID))
}
