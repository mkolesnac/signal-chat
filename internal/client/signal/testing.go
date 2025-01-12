package signal

import (
	grouprecord "github.com/crossle/libsignal-protocol-go/groups/state/record"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/protocol"
	"github.com/crossle/libsignal-protocol-go/state/record"
)

type StubStore struct {
	identityKey       *identity.KeyPair
	registrationId    uint32
	trustedIdentities map[*protocol.SignalAddress]*identity.Key
	preKeys           map[uint32]*record.PreKey
}

func (s *StubStore) GetIdentityKeyPair() *identity.KeyPair {
	return s.identityKey
}

func (s *StubStore) GetLocalRegistrationId() uint32 {
	return s.registrationId
}

func (s *StubStore) SaveIdentity(address *protocol.SignalAddress, identityKey *identity.Key) {
	s.trustedIdentities[address] = identityKey
}

func (s *StubStore) IsTrustedIdentity(address *protocol.SignalAddress, identityKey *identity.Key) bool {
	trusted := s.trustedIdentities[address]
	return trusted == nil || trusted.Fingerprint() == identityKey.Fingerprint()
}

func (s *StubStore) LoadPreKey(preKeyID uint32) *record.PreKey {
	return s.preKeys[preKeyID]
}

func (s *StubStore) StorePreKey(preKeyID uint32, preKeyRecord *record.PreKey) {
	s.preKeys[preKeyID] = preKeyRecord
}

func (s *StubStore) ContainsPreKey(preKeyID uint32) bool {
	return s.preKeys[preKeyID] != nil
}

func (s *StubStore) RemovePreKey(preKeyID uint32) {
	delete(s.preKeys, preKeyID)
}

func (s *StubStore) LoadSession(address *protocol.SignalAddress) *record.Session {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) GetSubDeviceSessions(name string) []uint32 {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) StoreSession(remoteAddress *protocol.SignalAddress, record *record.Session) {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) ContainsSession(remoteAddress *protocol.SignalAddress) bool {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) DeleteSession(remoteAddress *protocol.SignalAddress) {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) DeleteAllSessions() {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) LoadSignedPreKey(signedPreKeyID uint32) *record.SignedPreKey {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) LoadSignedPreKeys() []*record.SignedPreKey {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) StoreSignedPreKey(signedPreKeyID uint32, record *record.SignedPreKey) {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) ContainsSignedPreKey(signedPreKeyID uint32) bool {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) RemoveSignedPreKey(signedPreKeyID uint32) {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) StoreSenderKey(senderKeyName *protocol.SenderKeyName, keyRecord *grouprecord.SenderKey) {
	//TODO implement me
	panic("implement me")
}

func (s *StubStore) LoadSenderKey(senderKeyName *protocol.SenderKeyName) *grouprecord.SenderKey {
	//TODO implement me
	panic("implement me")
}
