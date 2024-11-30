package models

import (
	"signal-chat/cmd/server/storage"
	"strings"
)

var (
	keysPkPrefix         = "keys#"
	identityKeySkPrefix  = "identityKey#"
	signedPreKeySkPrefix = "signedPreKey#"
	preKeySkPrefix       = "preKey#"
)

type SignedPreKey struct {
	KeyID     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,32bytes"`
	Signature []byte `json:"signature" validate:"required,64bytes"`
}

type PreKey struct {
	KeyID     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,32bytes"`
}

func GetIdentityKeyPrimaryKey(accKey storage.PrimaryKey) storage.PrimaryKey {
	accID := ToAccountID(accKey)
	return storage.PrimaryKey{
		PartitionKey: keysPkPrefix + accID,
		SortKey:      identityKeySkPrefix,
	}
}

func GetPreKeyPrimaryKey(accKey storage.PrimaryKey, keyID string) storage.PrimaryKey {
	accID := ToAccountID(accKey)
	return storage.PrimaryKey{
		PartitionKey: keysPkPrefix + accID,
		SortKey:      preKeySkPrefix + keyID,
	}
}

func GetSignedPreKeyPrimaryKey(accKey storage.PrimaryKey, keyID string) storage.PrimaryKey {
	accID := ToAccountID(accKey)
	return storage.PrimaryKey{
		PartitionKey: keysPkPrefix + accID,
		SortKey:      signedPreKeySkPrefix + keyID,
	}
}

func IsIdentityKey(r storage.Resource) bool {
	return strings.HasPrefix(r.SortKey, identityKeySkPrefix)
}

func IsPreKey(r storage.Resource) bool {
	return strings.HasPrefix(r.SortKey, preKeySkPrefix)
}

func IsSignedPreKey(r storage.Resource) bool {
	return strings.HasPrefix(r.SortKey, signedPreKeySkPrefix)
}

func ToPreKeyID(primKey storage.PrimaryKey) string {
	return strings.Split(primKey.SortKey, "#")[0]
}

func ToSignedPreKeyID(primKey storage.PrimaryKey) string {
	return strings.Split(primKey.SortKey, "#")[0]
}
