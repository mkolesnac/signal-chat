package client

import (
	"errors"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/serialize"
	"github.com/crossle/libsignal-protocol-go/state/record"
	"github.com/crossle/libsignal-protocol-go/util/keyhelper"
	"regexp"
	"signal-chat/api"
	"signal-chat/client/database"
	"strconv"
)

var ErrAuthInvalidEmail = errors.New("email is not a valid email address")
var ErrAuthPwdTooShort = errors.New("password too short")

type DBConnector interface {
	Open(userID string) error
	Close() error
	WriteValue(pk database.PrimaryKey, value []byte) error
}

type Auth struct {
	db        DBConnector
	apiClient *APIClient
	signedIn  bool
}

func NewAuth(db DBConnector, apiClient *APIClient) *Auth {
	return &Auth{db: db, apiClient: apiClient}
}

func (a *Auth) SignUp(email, pwd string) error {
	if !isValidEmail(email) {
		return ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return ErrAuthPwdTooShort
	}

	err := a.db.Open(email)
	if err != nil {
		return fmt.Errorf("failed to open user database: %w", err)
	}

	ik, err := a.generateIdentityKey()
	if err != nil {
		return err
	}

	serializer := serialize.NewJSONSerializer()
	spk, err := a.generateSignedPreKey(ik, serializer)
	if err != nil {
		return err
	}

	preKeys, err := a.generatePreKeys(serializer)
	if err != nil {
		return err
	}

	spkSignature := spk.Signature()
	payload := api.SignUpRequest{
		Email:             email,
		Password:          pwd,
		IdentityPublicKey: ik.PublicKey().Serialize(),
		SignedPreKey: api.SignedPreKey{
			ID:        spk.ID(),
			PublicKey: spk.KeyPair().PublicKey().Serialize(),
			Signature: spkSignature[:],
		},
		PreKeys: nil,
	}
	for _, preKey := range preKeys {
		payload.PreKeys = append(payload.PreKeys, api.PreKey{
			ID:        preKey.ID().Value,
			PublicKey: preKey.KeyPair().PublicKey().Serialize(),
		})
	}

	a.apiClient.UseBasicAuth(email, pwd)
	_, err = a.apiClient.Post("/signup", payload)
	if err != nil {
		return fmt.Errorf("got error from server: %w", err)
	}

	a.signedIn = true
	return nil
}

func (a *Auth) SignIn(email, pwd string) error {
	if !isValidEmail(email) {
		return ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return ErrAuthPwdTooShort
	}

	err := a.db.Open(email)
	if err != nil {
		return fmt.Errorf("failed to open user database: %w", err)
	}

	payload := api.SignInRequest{
		Email:    email,
		Password: pwd,
	}

	a.apiClient.UseBasicAuth(email, pwd)
	_, err = a.apiClient.Post("/signin", payload)
	if err != nil {
		return fmt.Errorf("got error from server: %w", err)
	}

	a.signedIn = true
	return nil
}

func (a *Auth) SignOut() error {
	if !a.signedIn {
		panic("not signed in")
	}

	err := a.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	a.apiClient.Authorization = ""

	return nil
}

func (a *Auth) generateIdentityKey() (*identity.KeyPair, error) {
	identityKey, err := keyhelper.GenerateIdentityKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating identity key pair: %w", err)
	}
	err = a.db.WriteValue(database.PublicIdentityKeyPK(), identityKey.PublicKey().Serialize())
	if err != nil {
		return nil, fmt.Errorf("error writing public identity key: %w", err)
	}
	ipk := identityKey.PrivateKey().Serialize()
	err = a.db.WriteValue(database.PrivateIdentityKeyPK(), ipk[:])
	if err != nil {
		return nil, fmt.Errorf("error writing private identity key: %w", err)
	}

	return identityKey, nil
}

func (a *Auth) generateSignedPreKey(identityKey *identity.KeyPair, serializer *serialize.Serializer) (*record.SignedPreKey, error) {
	signedPreKey, err := keyhelper.GenerateSignedPreKey(identityKey, 0, serializer.SignedPreKeyRecord)
	if err != nil {
		return nil, fmt.Errorf("error generating signed pre keys: %w", err)
	}
	err = a.db.WriteValue(database.SignedPreKeyPK(strconv.Itoa(int(signedPreKey.ID()))), signedPreKey.Serialize())
	if err != nil {
		return nil, fmt.Errorf("error writing signed pre key: %w", err)
	}

	return signedPreKey, nil
}

func (a *Auth) generatePreKeys(serializer *serialize.Serializer) ([]*record.PreKey, error) {
	preKeys, err := keyhelper.GeneratePreKeys(1, 100, serializer.PreKeyRecord)
	if err != nil {
		return nil, fmt.Errorf("error generating pre keys: %w", err)
	}
	for _, preKey := range preKeys {
		err = a.db.WriteValue(database.PreKeyPK(strconv.Itoa(int(preKey.ID().Value))), preKey.Serialize())
		if err != nil {
			return nil, fmt.Errorf("error writing signed pre key: %w", err)
		}
	}

	return preKeys, nil
}

func isValidEmail(email string) bool {
	// Basic email regex
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}
