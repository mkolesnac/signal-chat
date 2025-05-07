package encryption

import (
	"errors"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/crossle/libsignal-protocol-go/groups"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/keys/prekey"
	"github.com/crossle/libsignal-protocol-go/protocol"
	"github.com/crossle/libsignal-protocol-go/serialize"
	"github.com/crossle/libsignal-protocol-go/session"
	"github.com/crossle/libsignal-protocol-go/util/keyhelper"
	"github.com/crossle/libsignal-protocol-go/util/optional"
	"signal-chat/client/database"
	"signal-chat/internal/apitypes"
)

type encryptor interface {
	InitializeKeyStore() (apitypes.KeyBundle, error)
	CreateEncryptionGroup(groupID string, recipientIDs []string) (map[string][]byte, error)
	ProcessSenderKeyDistributionMessage(groupID, senderID string, encryptedMsg []byte) error
	GroupEncrypt(groupID string, plaintext []byte) (*EncryptedMessage, error)
	GroupDecrypt(groupID, senderID string, ciphertext []byte) (*DecryptedMessage, error)
}

type PreKeyAPI interface {
	GetPreKeyBundle(id string) (apitypes.GetPreKeyBundleResponse, error)
}

type Manager struct {
	apiClient  PreKeyAPI
	store      *KeyStore
	serializer *serialize.Serializer
	ciphers    map[string]*session.Cipher
}

func NewEncryptionManager(db database.DB, apiClient PreKeyAPI) *Manager {
	serializer := serialize.NewJSONSerializer()

	return &Manager{
		apiClient:  apiClient,
		store:      NewKeyStore(db, serializer),
		serializer: serializer,
		ciphers:    make(map[string]*session.Cipher),
	}
}

// InitializeKeyStore should be called when a new user signs up
func (s *Manager) InitializeKeyStore() (apitypes.KeyBundle, error) {
	identityPair, err := keyhelper.GenerateIdentityKeyPair()
	if err != nil {
		return apitypes.KeyBundle{}, fmt.Errorf("error generating identity key pair: %w", err)
	}
	err = s.store.StoreIdentityKeyPair(identityPair)
	if err != nil {
		return apitypes.KeyBundle{}, fmt.Errorf("failed to keyStore identity key pair: %w", err)
	}

	// TODO: handle errors from store methods -> now they panic on error
	signedPreKeyPair, err := keyhelper.GenerateSignedPreKey(identityPair, 0, s.serializer.SignedPreKeyRecord)
	if err != nil {
		return apitypes.KeyBundle{}, fmt.Errorf("error generating signed pre keys: %w", err)
	}
	s.store.StoreSignedPreKey(signedPreKeyPair.ID(), signedPreKeyPair)

	preKeyPairs, err := keyhelper.GeneratePreKeys(1, 10, s.serializer.PreKeyRecord)
	if err != nil {
		return apitypes.KeyBundle{}, fmt.Errorf("error generating pre keys: %w", err)
	}
	var preKeys []apitypes.PreKey
	for _, preKeyPair := range preKeyPairs {
		s.store.StorePreKey(preKeyPair.ID().Value, preKeyPair)
		preKeyPub := preKeyPair.KeyPair().PublicKey().PublicKey()
		preKeys = append(preKeys, apitypes.PreKey{
			ID:        preKeyPair.ID().Value,
			PublicKey: preKeyPub[:],
		})
	}

	signature := signedPreKeyPair.Signature()

	identityPub := identityPair.PublicKey().PublicKey().PublicKey()
	signedPub := signedPreKeyPair.KeyPair().PublicKey().PublicKey()
	return apitypes.KeyBundle{
		IdentityKey: identityPub[:],
		SignedPreKey: apitypes.SignedPreKey{
			ID:        signedPreKeyPair.ID(),
			PublicKey: signedPub[:],
			Signature: signature[:],
		},
		PreKeys: preKeys,
	}, nil
}

func (s *Manager) CreateEncryptionGroup(groupID string, recipientIDs []string) (map[string][]byte, error) {
	keyName := protocol.NewSenderKeyName(groupID, protocol.NewSignalAddress("-", 1)) // - is name for my key
	builder := groups.NewGroupSessionBuilder(s.store, s.serializer)
	keyMsg, err := builder.Create(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to create group session: %w", err)
	}

	keyMsgBytes := keyMsg.Serialize()
	messages := make(map[string][]byte)
	for _, id := range recipientIDs {
		ciphertext, err := s.pairwiseEncrypt(keyMsgBytes, id)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt key distribution message for user %s: %w", id, err)
		}
		messages[id] = ciphertext
	}

	return messages, nil
}

func (s *Manager) ProcessSenderKeyDistributionMessage(groupID, senderID string, encryptedMsg []byte) error {
	plaintext, err := s.pairwiseDecrypt(encryptedMsg, senderID)
	if err != nil {
		return fmt.Errorf("failed to decrypt key distribution message from user %s: %w", senderID, err)
	}

	keyName := protocol.NewSenderKeyName(groupID, protocol.NewSignalAddress(senderID, 1))
	builder := groups.NewGroupSessionBuilder(s.store, s.serializer)
	keyMessage, err := protocol.NewSenderKeyDistributionMessageFromBytes(plaintext, s.store.serializer.SenderKeyDistributionMessage)
	if err != nil {
		return fmt.Errorf("failed to deserialize sender key distribution message: %w", err)
	}

	builder.Process(keyName, keyMessage)
	return nil
}

func (s *Manager) GroupEncrypt(groupID string, plaintext []byte) (*EncryptedMessage, error) {
	keyName := protocol.NewSenderKeyName(groupID, protocol.NewSignalAddress("-", 1))

	senderKey := s.store.LoadSenderKey(keyName)
	if senderKey == nil {
		return nil, errors.New("sender key for group not found")
	}

	builder := groups.NewGroupSessionBuilder(s.store, s.serializer)
	cipher := groups.NewGroupCipher(builder, keyName, s.store)
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt group message: %w", err)
	}

	return newEncryptedMessage(ciphertext), nil
}

func (s *Manager) GroupDecrypt(groupID, senderID string, ciphertext []byte) (*DecryptedMessage, error) {
	keyName := protocol.NewSenderKeyName(groupID, protocol.NewSignalAddress(senderID, 1))
	senderKey := s.store.LoadSenderKey(keyName)
	if senderKey == nil {
		return nil, errors.New("sender key for group not found")
	}

	msg, err := protocol.NewSenderKeyMessageFromBytes(ciphertext, s.serializer.SenderKeyMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize cipthertext: %w", err)
	}

	builder := groups.NewGroupSessionBuilder(s.store, s.serializer)
	cipher := groups.NewGroupCipher(builder, keyName, s.store)
	plaintext, err := cipher.Decrypt(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt group message: %w", err)
	}

	return newDecryptedMessage(plaintext, msg), nil
}

func (s *Manager) pairwiseEncrypt(plaintext []byte, recipientID string) ([]byte, error) {
	addr := protocol.NewSignalAddress(recipientID, 1)
	var cipher *session.Cipher
	if !s.store.ContainsSession(addr) {
		bundle, err := s.getPreKeyBundle(recipientID)
		if err != nil {
			return nil, err
		}

		builder := session.NewBuilderFromSignal(s.store, addr, s.serializer)
		err = builder.ProcessBundle(bundle)
		if err != nil {
			return nil, fmt.Errorf("failed to create session with user '%s' due to error: %w", addr.Name(), err)
		}

		cipher = session.NewCipher(builder, addr)
	} else {
		cipher = session.NewCipherFromSession(addr, s.store, s.store, s.store, s.serializer.PreKeySignalMessage, s.serializer.SignalMessage)
	}

	encrypted, err := cipher.Encrypt(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt the message: %w", err)
	}

	return encrypted.Serialize(), nil
}

func (s *Manager) pairwiseDecrypt(encryptedMsg []byte, senderID string) ([]byte, error) {
	addr := protocol.NewSignalAddress(senderID, 1)
	if !s.store.ContainsSession(addr) {
		msg, err := protocol.NewPreKeySignalMessageFromBytes(encryptedMsg, s.serializer.PreKeySignalMessage, s.serializer.SignalMessage)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall pre key signal message: %w", err)
		}

		builder := session.NewBuilderFromSignal(s.store, addr, s.serializer)
		cipher := session.NewCipher(builder, addr)
		plaintext, err := cipher.DecryptMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt pre key signal message: %w", err)
		}
		return plaintext, nil
	}

	msg, err := protocol.NewSignalMessageFromBytes(encryptedMsg, s.serializer.SignalMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall signal message: %w", err)
	}

	cipher := session.NewCipherFromSession(addr, s.store, s.store, s.store, s.serializer.PreKeySignalMessage, s.serializer.SignalMessage)
	plaintext, err := cipher.Decrypt(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt signal message: %w", err)
	}
	return plaintext, nil
}

func (s *Manager) getPreKeyBundle(recipientID string) (*prekey.Bundle, error) {
	resp, err := s.apiClient.GetPreKeyBundle(recipientID)
	if err != nil {
		return nil, err
	}

	return prekey.NewBundle(
		resp.PreKeyBundle.RegistrationID,
		1,
		optional.NewOptionalUint32(resp.PreKeyBundle.PreKey.ID),
		resp.PreKeyBundle.SignedPreKey.ID,
		ecc.NewDjbECPublicKey([32]byte(resp.PreKeyBundle.PreKey.PublicKey)),
		ecc.NewDjbECPublicKey([32]byte(resp.PreKeyBundle.SignedPreKey.PublicKey)),
		[64]byte(resp.PreKeyBundle.SignedPreKey.Signature),
		identity.NewKey(ecc.NewDjbECPublicKey([32]byte(resp.PreKeyBundle.IdentityKey)))), nil
}
