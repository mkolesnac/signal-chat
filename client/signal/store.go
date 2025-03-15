package signal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	grouprecord "github.com/crossle/libsignal-protocol-go/groups/state/record"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/protocol"
	"github.com/crossle/libsignal-protocol-go/serialize"
	"github.com/crossle/libsignal-protocol-go/state/record"
	"signal-chat/client/database"
)

var ErrInitialized = errors.New("store is already initialized. Create new store using NewStore function")
var ErrNotInitialized = errors.New("store not initialized. Use InitializeKeyStore or LoadForUser functions to initialize the store")

const (
	RegistrationIdDatabaseKey  string = "registrationID"
	PrivateIdentityDatabaseKey string = "identityKey#private"
	PublicIdentityDatabaseKey  string = "identityKey#public"
)

type Store struct {
	registrationID    uint32
	trustedIdentities map[string]*identity.Key
	identityKey       *identity.KeyPair
	signedPreKeys     map[uint32]*record.SignedPreKey
	preKeys           map[uint32]*record.PreKey
	sessions          map[*protocol.SignalAddress]*record.Session

	serializer *serialize.Serializer
	db         *database.Database
}

func NewStore(db *database.Database) *Store {
	return &Store{
		trustedIdentities: make(map[string]*identity.Key),
		signedPreKeys:     make(map[uint32]*record.SignedPreKey),
		preKeys:           make(map[uint32]*record.PreKey),
		sessions:          make(map[*protocol.SignalAddress]*record.Session),
		serializer:        serialize.NewJSONSerializer(),
		db:                db,
	}
}

func (s *Store) SetupNewUser(registrationID uint32, identityKey *identity.KeyPair, signedPreKey *record.SignedPreKey, preKeys []*record.PreKey) {
	s.registrationID = registrationID
	// Write registrationID to database in LittleEndian format
	idBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(idBytes, registrationID)
	err := s.db.Write(RegistrationIdDatabaseKey, idBytes)
	if err != nil {
		// Signal store interface function do not return error. Their implementations can
		//therefore only panic if there is some error. For the sake of consistency we use
		//panic also in this function.
		panic(fmt.Errorf("failed to store registrationID: %v", err))
	}

	s.identityKey = identityKey
	// Write private identity key to database
	privateBytes := identityKey.PrivateKey().Serialize()
	err = s.db.Write(PrivateIdentityDatabaseKey, privateBytes[:])
	if err != nil {
		panic(fmt.Errorf("failed to store private identity key: %v", err))
	}
	// Write private identity key to database
	err = s.db.Write(PublicIdentityDatabaseKey, identityKey.PublicKey().Serialize())
	if err != nil {
		panic(fmt.Errorf("failed to store public identity key: %v", err))
	}

	s.StoreSignedPreKey(signedPreKey.ID(), signedPreKey)
	for _, preKey := range preKeys {
		s.StorePreKey(preKey.ID().Value, preKey)
	}
}

func (s *Store) Serializer() *serialize.Serializer {
	return s.serializer
}

func (s *Store) GetIdentityKeyPair() *identity.KeyPair {
	s.panicIfNotInitialized()
	if s.identityKey != nil {
		return s.identityKey
	}

	// Load private key
	privateBytes := s.readValueFromDb(PrivateIdentityDatabaseKey)
	private := ecc.NewDjbECPrivateKey([32]byte(privateBytes))

	// Load public key
	publicBytes := s.readValueFromDb(PublicIdentityDatabaseKey)
	public := ecc.NewDjbECPublicKey([32]byte(publicBytes))

	pair := identity.NewKeyPair(identity.NewKey(public), private)
	// Add the identity key pair to cache
	s.identityKey = pair

	return pair
}

func (s *Store) GetLocalRegistrationId() uint32 {
	if s.registrationID != 0 {
		return s.registrationID
	}

	bytes, err := s.db.Read(RegistrationIdDatabaseKey)
	if err != nil {
		panic(fmt.Errorf("failed to read registrationID: %w", err))
	}
	if len(bytes) < 4 {
		panic(fmt.Errorf("failed to read registrationID: got %d bytes, needed at least 4", len(bytes)))
	}
	id := binary.LittleEndian.Uint32(bytes)
	s.registrationID = id
	return id
}

func (s *Store) SaveIdentity(address *protocol.SignalAddress, identityKey *identity.Key) {
	s.panicIfNotInitialized()

	id := address.Name()
	dbKey := fmt.Sprintf("identity#%s", id)
	s.writeValueToDb(dbKey, identityKey.Serialize())

	// Add identity to cache
	s.trustedIdentities[id] = identityKey
}

// IsTrustedIdentity determines whether a remote client's identity is trusted. Trust is based on
// 'trust on first use'. This means that an identity key is considered 'trusted'
// if there is no entry for the recipient in the local store, or if it matches the
// saved key for a recipient in the local store.
func (s *Store) IsTrustedIdentity(address *protocol.SignalAddress, identityKey *identity.Key) bool {
	s.panicIfNotInitialized()

	id := address.Name()
	if t, ok := s.trustedIdentities[id]; ok {
		return t.Fingerprint() == identityKey.Fingerprint()
	}

	dbKey := fmt.Sprintf("identity#%s", id)
	value := s.readValueFromDb(dbKey)
	if value == nil {
		// If there is no entry for the recipient in the db the identity key is considered trusted
		return true
	}

	storedIdentity := identity.NewKeyFromBytes([32]byte(value), 0)

	s.trustedIdentities[id] = identityKey // add identity to cache
	return storedIdentity.Fingerprint() == identityKey.Fingerprint()
}

func (s *Store) LoadPreKey(preKeyID uint32) *record.PreKey {
	s.panicIfNotInitialized()

	if k, ok := s.preKeys[preKeyID]; ok {
		return k
	}

	dbKey := fmt.Sprintf("prekey#%v", preKeyID)
	value := s.readValueFromDb(dbKey)
	if value == nil {
		return nil // prekey not found
	}

	preKey, err := record.NewPreKeyFromBytes(value, s.serializer.PreKeyRecord)
	if err != nil {
		panic(fmt.Errorf("failed to parse pre key from bytes: %w", err))
	}

	s.preKeys[preKeyID] = preKey // add identity to cache
	return nil
}

func (s *Store) StorePreKey(preKeyID uint32, preKeyRecord *record.PreKey) {
	s.panicIfNotInitialized()

	dbKey := fmt.Sprintf("prekey#%v", preKeyID)
	s.writeValueToDb(dbKey, preKeyRecord.Serialize())

	// Add pre key to cache
	s.preKeys[preKeyID] = preKeyRecord
}

func (s *Store) ContainsPreKey(preKeyID uint32) bool {
	return s.LoadPreKey(preKeyID) != nil
}

func (s *Store) RemovePreKey(preKeyID uint32) {
	s.panicIfNotInitialized()

	delete(s.preKeys, preKeyID)
	dbKey := fmt.Sprintf("prekey#%v", preKeyID)
	s.deleteValueFromDb(dbKey)
}

func (s *Store) LoadSession(address *protocol.SignalAddress) *record.Session {
	s.panicIfNotInitialized()

	return s.sessions[address]
}

func (s *Store) GetSubDeviceSessions(name string) []uint32 {
	s.panicIfNotInitialized()

	var ids []uint32
	for addr := range s.sessions {
		if addr.Name() == name && addr.DeviceID() != 0 {
			ids = append(ids, addr.DeviceID())
		}
	}

	return ids
}

func (s *Store) StoreSession(remoteAddress *protocol.SignalAddress, record *record.Session) {
	s.panicIfNotInitialized()
	s.sessions[remoteAddress] = record
}

func (s *Store) ContainsSession(remoteAddress *protocol.SignalAddress) bool {
	s.panicIfNotInitialized()
	return s.sessions[remoteAddress] != nil
}

func (s *Store) DeleteSession(remoteAddress *protocol.SignalAddress) {
	s.panicIfNotInitialized()
	delete(s.sessions, remoteAddress)
}

func (s *Store) DeleteAllSessions() {
	s.sessions = make(map[*protocol.SignalAddress]*record.Session)
}

func (s *Store) LoadSignedPreKey(signedPreKeyID uint32) *record.SignedPreKey {
	return s.signedPreKeys[signedPreKeyID]
}

func (s *Store) LoadSignedPreKeys() []*record.SignedPreKey {
	keys := make([]*record.SignedPreKey, 0, len(s.signedPreKeys))
	for _, key := range s.signedPreKeys {
		keys = append(keys, key)
	}
	return keys
}

func (s *Store) StoreSignedPreKey(signedPreKeyID uint32, record *record.SignedPreKey) {
	s.signedPreKeys[signedPreKeyID] = record
}

func (s *Store) ContainsSignedPreKey(signedPreKeyID uint32) bool {
	return s.signedPreKeys[signedPreKeyID] != nil
}

func (s *Store) RemoveSignedPreKey(signedPreKeyID uint32) {
	delete(s.signedPreKeys, signedPreKeyID)
}

func (s *Store) StoreSenderKey(senderKeyName *protocol.SenderKeyName, keyRecord *grouprecord.SenderKey) {
	// TODO:
	panic("Not implemented")
}

func (s *Store) LoadSenderKey(senderKeyName *protocol.SenderKeyName) *grouprecord.SenderKey {
	// TODO:
	panic("Not implemented")
}

func (s *Store) panicIfInitialized() {
	if s.db == nil {
		panic(ErrInitialized)
	}
}

func (s *Store) panicIfNotInitialized() {
	if s.db == nil {
		panic(ErrNotInitialized)
	}
}

func (s *Store) readValueFromDb(key string) []byte {
	value, err := s.db.Read(key)
	if err != nil {
		panic(fmt.Errorf("cannot read key %s from Database: %w", key, err))
	}
	return value
}

func (s *Store) writeValueToDb(key string, value []byte) {
	err := s.db.Write(key, value)
	if err != nil {
		panic(fmt.Errorf("cannot write key %s to Database: %w", key, err))
	}
}

func (s *Store) deleteValueFromDb(key string) {
	err := s.db.Delete(key)
	if err != nil {
		panic(fmt.Errorf("cannot delete key %s from db: %w", key, err))
	}
}
