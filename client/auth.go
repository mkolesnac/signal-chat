package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/serialize"
	"github.com/crossle/libsignal-protocol-go/state/record"
	"github.com/crossle/libsignal-protocol-go/util/keyhelper"
	"net/http"
	"regexp"
	"signal-chat/client/database"
	"signal-chat/client/models"
	"signal-chat/internal/api"
	"strconv"
)

var ErrAuthInvalidEmail = errors.New("email is not a valid email address")
var ErrAuthPwdTooShort = errors.New("password too short")

type AuthAPI interface {
	StartSession(username, password string) error
	Close() error
	Post(route string, payload any) (int, []byte, error)
}

type AuthDatabase interface {
	Open(userID string) error
	Close() error
	Write(pk database.PrimaryKey, value []byte) error
}

type Auth struct {
	db        AuthDatabase
	apiClient AuthAPI
	signedIn  bool
}

func NewAuth(db AuthDatabase, apiClient AuthAPI) *Auth {
	return &Auth{db: db, apiClient: apiClient}
}

func (a *Auth) SignUp(email, pwd string) (models.User, error) {
	if !isValidEmail(email) {
		return models.User{}, ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return models.User{}, ErrAuthPwdTooShort
	}

	err := a.db.Open(email)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to open user database: %w", err)
	}

	ik, err := a.storeIdentityKey()
	if err != nil {
		return models.User{}, err
	}

	serializer := serialize.NewJSONSerializer()
	spk, err := a.storeSignedPreKey(ik, serializer)
	if err != nil {
		return models.User{}, err
	}

	preKeys, err := a.storePreKeys(serializer)
	if err != nil {
		return models.User{}, err
	}

	spkSignature := spk.Signature()
	payload := api.SignUpRequest{
		UserName:          email,
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

	status, body, err := a.apiClient.Post(api.EndpointSignUp, payload)
	if err != nil {
		return models.User{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return models.User{}, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.SignUpResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return models.User{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	err = a.apiClient.StartSession(email, pwd)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to start API session: %w", err)
	}

	a.signedIn = true
	user := models.User{
		ID:       resp.UserID,
		Username: email,
	}
	return user, nil
}

func (a *Auth) SignIn(email, pwd string) (models.User, error) {
	if !isValidEmail(email) {
		return models.User{}, ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return models.User{}, ErrAuthPwdTooShort
	}

	err := a.db.Open(email)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to open user database: %w", err)
	}

	payload := api.SignInRequest{
		Username: email,
		Password: pwd,
	}

	status, body, err := a.apiClient.Post(api.EndpointSignIn, payload)
	if err != nil {
		return models.User{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return models.User{}, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.SignInResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return models.User{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	err = a.apiClient.StartSession(email, pwd)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to start API session: %w", err)
	}

	a.signedIn = true
	user := models.User{
		ID:       resp.UserID,
		Username: resp.Username,
	}
	return user, nil
}

func (a *Auth) SignOut() error {
	if !a.signedIn {
		panic("not signed in")
	}

	if err := a.apiClient.Close(); err != nil {
		return fmt.Errorf("failed to close API session: %w", err)
	}

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

func (a *Auth) storeIdentityKey() (*identity.KeyPair, error) {
	identityKey, err := keyhelper.GenerateIdentityKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating identity key pair: %w", err)
	}
	err = a.db.Write(database.PublicIdentityKeyPK(), identityKey.PublicKey().Serialize())
	if err != nil {
		return nil, fmt.Errorf("error writing public identity key: %w", err)
	}
	ipk := identityKey.PrivateKey().Serialize()
	err = a.db.Write(database.PrivateIdentityKeyPK(), ipk[:])
	if err != nil {
		return nil, fmt.Errorf("error writing private identity key: %w", err)
	}

	return identityKey, nil
}

func (a *Auth) storeSignedPreKey(identityKey *identity.KeyPair, serializer *serialize.Serializer) (*record.SignedPreKey, error) {
	signedPreKey, err := keyhelper.GenerateSignedPreKey(identityKey, 0, serializer.SignedPreKeyRecord)
	if err != nil {
		return nil, fmt.Errorf("error generating signed pre keys: %w", err)
	}
	err = a.db.Write(database.SignedPreKeyPK(strconv.Itoa(int(signedPreKey.ID()))), signedPreKey.Serialize())
	if err != nil {
		return nil, fmt.Errorf("error writing signed pre key: %w", err)
	}

	return signedPreKey, nil
}

func (a *Auth) storePreKeys(serializer *serialize.Serializer) ([]*record.PreKey, error) {
	preKeys, err := keyhelper.GeneratePreKeys(1, 100, serializer.PreKeyRecord)
	if err != nil {
		return nil, fmt.Errorf("error generating pre keys: %w", err)
	}
	for _, preKey := range preKeys {
		err = a.db.Write(database.PreKeyPK(strconv.Itoa(int(preKey.ID().Value))), preKey.Serialize())
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
