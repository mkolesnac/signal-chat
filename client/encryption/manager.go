package encryption

import (
	"encoding/json"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/keys/prekey"
	"github.com/crossle/libsignal-protocol-go/protocol"
	"github.com/crossle/libsignal-protocol-go/serialize"
	"github.com/crossle/libsignal-protocol-go/session"
	"github.com/crossle/libsignal-protocol-go/util/keyhelper"
	"github.com/crossle/libsignal-protocol-go/util/optional"
	"net/http"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"signal-chat/internal/api"
)

type encryptor interface {
	InitializeKeyStore() (api.KeyBundle, error)
	Encrypt(plaintext []byte, recipientID string) ([]byte, error)
	Decrypt(ciphertext []byte, senderID string) ([]byte, error)
}

type Manager struct {
	apiClient  apiclient.Client
	store      *KeyStore
	serializer *serialize.Serializer
	ciphers    map[string]*session.Cipher
}

func NewEncryptionManager(db database.DB, apiClient apiclient.Client) *Manager {
	serializer := serialize.NewJSONSerializer()

	return &Manager{
		apiClient:  apiClient,
		store:      NewKeyStore(db, serializer),
		serializer: serializer,
		ciphers:    make(map[string]*session.Cipher),
	}
}

// InitializeKeyStore should be called when a new user signs up
func (s *Manager) InitializeKeyStore() (api.KeyBundle, error) {
	identityPair, err := keyhelper.GenerateIdentityKeyPair()
	if err != nil {
		return api.KeyBundle{}, fmt.Errorf("error generating identity key pair: %w", err)
	}
	err = s.store.StoreIdentityKeyPair(identityPair)
	if err != nil {
		return api.KeyBundle{}, fmt.Errorf("failed to keyStore identity key pair: %w", err)
	}

	// TODO: handle errors from store methods -> now they panic on error
	signedPreKeyPair, err := keyhelper.GenerateSignedPreKey(identityPair, 0, s.serializer.SignedPreKeyRecord)
	if err != nil {
		return api.KeyBundle{}, fmt.Errorf("error generating signed pre keys: %w", err)
	}
	s.store.StoreSignedPreKey(signedPreKeyPair.ID(), signedPreKeyPair)

	preKeyPairs, err := keyhelper.GeneratePreKeys(1, 100, s.serializer.PreKeyRecord)
	if err != nil {
		return api.KeyBundle{}, fmt.Errorf("error generating pre keys: %w", err)
	}
	var preKeys []api.PreKey
	for _, preKeyPair := range preKeyPairs {
		s.store.StorePreKey(preKeyPair.ID().Value, preKeyPair)
		preKeyPub := preKeyPair.KeyPair().PublicKey().PublicKey()
		preKeys = append(preKeys, api.PreKey{
			ID:        preKeyPair.ID().Value,
			PublicKey: preKeyPub[:],
		})
	}

	signature := signedPreKeyPair.Signature()

	identityPub := identityPair.PublicKey().PublicKey().PublicKey()
	signedPub := signedPreKeyPair.KeyPair().PublicKey().PublicKey()
	return api.KeyBundle{
		IdentityKey: identityPub[:],
		SignedPreKey: api.SignedPreKey{
			ID:        signedPreKeyPair.ID(),
			PublicKey: signedPub[:],
			Signature: signature[:],
		},
		PreKeys: preKeys,
	}, nil
}

func (s *Manager) GetCurrentMaterial(sessionID string) *Material {
	addr := protocol.NewSignalAddress(sessionID, 1)
	sessionRecord := s.store.LoadSession(addr)
	sessionState := sessionRecord.SessionState()
	chainKey := sessionState.SenderChainKey()
	messageKeys := chainKey.MessageKeys()

	senderRatchet := sessionState.SenderRatchetKey().PublicKey()
	return &Material{
		RootKey:                sessionState.RootKey().Bytes(),
		SenderChainKey:         sessionState.SenderChainKey().Key(),
		SenderRatchetKey:       senderRatchet[:],
		PreviousMessageCounter: sessionState.PreviousCounter(),
		SessionVersion:         sessionState.Version(),
		MessageKeys: MessageKeys{
			CipherKey: messageKeys.CipherKey(),
			MacKey:    messageKeys.MacKey(),
			IV:        messageKeys.Iv(),
			Index:     messageKeys.Index(),
		},
	}
}

func (s *Manager) Encrypt(plaintext []byte, recipientID string) ([]byte, error) {
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

	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt the message: %w", err)
	}

	return ciphertext.Serialize(), nil
}

func (s *Manager) Decrypt(ciphertext []byte, senderID string) ([]byte, error) {
	addr := protocol.NewSignalAddress(senderID, 1)
	if !s.store.ContainsSession(addr) {
		msg, err := protocol.NewPreKeySignalMessageFromBytes(ciphertext, s.serializer.PreKeySignalMessage, s.serializer.SignalMessage)
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

	msg, err := protocol.NewSignalMessageFromBytes(ciphertext, s.serializer.SignalMessage)
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
	status, body, err := s.apiClient.Get(api.EndpointUserKeys(recipientID))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.GetPrekeyBundleResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall key bundle from server: %w", err)
	}

	return prekey.NewBundle(
		resp.PreKeyBundle.RegistrationID,
		1,
		optional.NewOptionalUint32(resp.PreKeyBundle.PreKey.ID),
		resp.PreKeyBundle.SignedPreKey.ID,
		ecc.NewDjbECPublicKey([32]byte(resp.PreKeyBundle.PreKey.PublicKey)),
		ecc.NewDjbECPublicKey([32]byte(resp.PreKeyBundle.SignedPreKey.PublicKey)),
		[64]byte(resp.PreKeyBundle.SignedPreKeySignature),
		identity.NewKey(ecc.NewDjbECPublicKey([32]byte(resp.PreKeyBundle.IdentityKey)))), nil
}
