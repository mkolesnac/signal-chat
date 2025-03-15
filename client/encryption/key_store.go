package encryption

import (
	"encoding/binary"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	grouprecord "github.com/crossle/libsignal-protocol-go/groups/state/record"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/protocol"
	"github.com/crossle/libsignal-protocol-go/serialize"
	"github.com/crossle/libsignal-protocol-go/state/record"
	"signal-chat/client/database"
	"strconv"
	"strings"
)

type KeyStore struct {
	db         database.DB
	serializer *serialize.Serializer
}

func NewKeyStore(db database.DB, serializer *serialize.Serializer) *KeyStore {
	return &KeyStore{
		db:         db,
		serializer: serializer,
	}
}

func (k *KeyStore) StoreIdentityKeyPair(keyPair *identity.KeyPair) error {
	pubBytes := keyPair.PublicKey().PublicKey().PublicKey()
	privBytes := keyPair.PrivateKey().Serialize()
	bytes := append(pubBytes[:], privBytes[:]...)
	err := k.db.Write("identityKeyPair", bytes)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeyStore) GetIdentityKeyPair() *identity.KeyPair {
	bytes, err := k.db.Read("identityKeyPair")
	if err != nil {
		panic(err)
	}
	if bytes == nil {
		return nil
	}

	pubKey := identity.NewKeyFromBytes([32]byte(bytes[0:32]), 0)
	privKey := ecc.NewDjbECPrivateKey([32]byte(bytes[32:]))

	return identity.NewKeyPair(&pubKey, privKey)
}

func (k *KeyStore) GetLocalRegistrationId() uint32 {
	bytes, err := k.db.Read("registrationId")
	if err != nil {
		panic(err)
	}
	if bytes == nil {
		return 0
	}

	return binary.BigEndian.Uint32(bytes)
}

func (k *KeyStore) SaveIdentity(address *protocol.SignalAddress, identityKey *identity.Key) {
	key := fmt.Sprintf("identity#%v", address.String())
	bytes := identityKey.PublicKey().PublicKey()
	err := k.db.Write(key, bytes[:])
	if err != nil {
		panic(err)
	}
}

func (k *KeyStore) IsTrustedIdentity(address *protocol.SignalAddress, identityKey *identity.Key) bool {
	key := fmt.Sprintf("identity#%v", address.String())
	bytes, err := k.db.Read(key)
	if err != nil {
		panic(err)
	}
	if bytes == nil {
		// if there is no entry for the recipient in the local store the identity key is considered trusted
		return true
	}

	// If a key is found, Identity key is considered trusted only if it matches the saved key for the recipient
	stored := identity.NewKeyFromBytes([32]byte(bytes), 0)
	return stored.Fingerprint() == identityKey.Fingerprint()
}

func (k *KeyStore) LoadPreKey(preKeyID uint32) *record.PreKey {
	key := fmt.Sprintf("preKey#%d", preKeyID)
	bytes, err := k.db.Read(key)
	if err != nil {
		panic(err)
	}
	if bytes == nil {
		return nil
	}

	preKey, err := record.NewPreKeyFromBytes(bytes, k.serializer.PreKeyRecord)
	if err != nil {
		panic(err)
	}

	return preKey
}

func (k *KeyStore) StorePreKey(preKeyID uint32, preKeyRecord *record.PreKey) {
	key := fmt.Sprintf("preKey#%d", preKeyID)
	err := k.db.Write(key, preKeyRecord.Serialize())
	if err != nil {
		panic(err)
	}
}

func (k *KeyStore) ContainsPreKey(preKeyID uint32) bool {
	key := fmt.Sprintf("preKey#%d", preKeyID)
	bytes, err := k.db.Read(key)
	if err != nil {
		panic(err)
	}
	return bytes != nil
}

func (k *KeyStore) RemovePreKey(preKeyID uint32) {
	key := fmt.Sprintf("preKey#%d", preKeyID)
	err := k.db.Delete(key)
	if err != nil {
		panic(err)
	}
}

func (k *KeyStore) LoadSession(address *protocol.SignalAddress) *record.Session {
	key := fmt.Sprintf("session#%v", address.String())
	bytes, err := k.db.Read(key)
	if err != nil {
		panic(err)
	}
	if bytes == nil {
		return record.NewSession(k.serializer.Session, k.serializer.State)
	}

	session, err := record.NewSessionFromBytes(bytes, k.serializer.Session, k.serializer.State)
	if err != nil {
		panic(err)
	}

	return session
}

func (k *KeyStore) GetSubDeviceSessions(name string) []uint32 {
	prefix := fmt.Sprintf("session#%v", name)
	items, err := k.db.Query(prefix)
	if err != nil {
		panic(err)
	}

	var ids []uint32
	for key := range items {
		str := strings.TrimPrefix(key, prefix)
		deviceID, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			panic(err)
		}
		ids = append(ids, uint32(deviceID))
	}

	return ids
}

func (k *KeyStore) StoreSession(remoteAddress *protocol.SignalAddress, record *record.Session) {
	key := fmt.Sprintf("session#%v", remoteAddress.String())
	err := k.db.Write(key, record.Serialize())
	if err != nil {
		panic(err)
	}
}

func (k *KeyStore) ContainsSession(remoteAddress *protocol.SignalAddress) bool {
	key := fmt.Sprintf("session#%v", remoteAddress.String())
	bytes, err := k.db.Read(key)
	if err != nil {
		panic(err)
	}
	return bytes != nil
}

func (k *KeyStore) DeleteSession(remoteAddress *protocol.SignalAddress) {
	key := fmt.Sprintf("session#%v", remoteAddress.String())
	err := k.db.Delete(key)
	if err != nil {
		panic(err)
	}
}

func (k *KeyStore) DeleteAllSessions() {
	prefix := fmt.Sprintf("session#")
	items, err := k.db.Query(prefix)
	if err != nil {
		panic(err)
	}

	for key := range items {
		err := k.db.Delete(key)
		if err != nil {
			panic(err)
		}
	}
}

func (k *KeyStore) LoadSignedPreKey(signedPreKeyID uint32) *record.SignedPreKey {
	key := fmt.Sprintf("signedPreKey#%d", signedPreKeyID)
	bytes, err := k.db.Read(key)
	if err != nil {
		panic(err)
	}
	if bytes == nil {
		return nil
	}

	signedPreKey, err := record.NewSignedPreKeyFromBytes(bytes, k.serializer.SignedPreKeyRecord)
	if err != nil {
		panic(err)
	}

	return signedPreKey
}

func (k *KeyStore) LoadSignedPreKeys() []*record.SignedPreKey {
	prefix := fmt.Sprintf("signedPreKey#")
	items, err := k.db.Query(prefix)
	if err != nil {
		panic(err)
	}

	var signedPreKeys []*record.SignedPreKey
	for key := range items {
		signedPreKey, err := record.NewSignedPreKeyFromBytes(items[key], k.serializer.SignedPreKeyRecord)
		if err != nil {
			panic(err)
		}

		signedPreKeys = append(signedPreKeys, signedPreKey)
	}

	return signedPreKeys
}

func (k *KeyStore) StoreSignedPreKey(signedPreKeyID uint32, record *record.SignedPreKey) {
	key := fmt.Sprintf("signedPreKey#%d", signedPreKeyID)
	bytes := record.Serialize()
	err := k.db.Write(key, bytes)
	if err != nil {
		panic(err)
	}
}

func (k *KeyStore) ContainsSignedPreKey(signedPreKeyID uint32) bool {
	key := fmt.Sprintf("signedPreKey#%d", signedPreKeyID)
	bytes, err := k.db.Read(key)
	if err != nil {
		panic(err)
	}

	return bytes != nil
}

func (k *KeyStore) RemoveSignedPreKey(signedPreKeyID uint32) {
	key := fmt.Sprintf("signedPreKey#%d", signedPreKeyID)
	err := k.db.Delete(key)
	if err != nil {
		panic(err)
	}
}

func (k *KeyStore) StoreSenderKey(senderKeyName *protocol.SenderKeyName, keyRecord *grouprecord.SenderKey) {
	panic("invalid key")
	//key := fmt.Sprintf("senderKey#%v", *senderKeyName)
	//err := k.db.Write(key, keyRecord.Serialize())
	//if err != nil {
	//	panic(err)
	//}
}

func (k *KeyStore) LoadSenderKey(senderKeyName *protocol.SenderKeyName) *grouprecord.SenderKey {
	panic("invalid key")

	//key := fmt.Sprintf("senderKey#%v", *senderKeyName)
	//bytes, err := k.db.Read(key)
	//if err != nil {
	//	panic(err)
	//}
	//if bytes == nil {
	//	return nil
	//}
	//
	//senderKey, err := grouprecord.NewSenderKeyFromBytes(bytes, k.serializer.SenderKeyRecord, k.serializer.SenderKeyState)
	//if err != nil {
	//	panic(err)
	//}
	//
	//return senderKey
}
