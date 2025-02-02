package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/keys/identity"
	"github.com/crossle/libsignal-protocol-go/serialize"
	"github.com/crossle/libsignal-protocol-go/state/record"
	"github.com/crossle/libsignal-protocol-go/util/keyhelper"
	"net/http"
	"regexp"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"signal-chat/internal/api"
	"strconv"
)

var ErrAuthInvalidEmail = errors.New("email is not a valid email address")
var ErrAuthPwdTooShort = errors.New("password too short")

type AuthAPIClient interface {
	apiclient.Client
	SetAuthorization(username, password string)
	ClearAuthorization()
}

type AuthDatabase interface {
	Open(userID string) error
	Close() error
	Write(pk database.PrimaryKey, value []byte) error
}

type Auth struct {
	db        AuthDatabase
	apiClient AuthAPIClient
	signedIn  bool
}

func NewAuth(db AuthDatabase, apiClient AuthAPIClient) *Auth {
	return &Auth{db: db, apiClient: apiClient}
}

func (a *Auth) SignUp(email, pwd string) (User, error) {
	if !isValidEmail(email) {
		return User{}, ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return User{}, ErrAuthPwdTooShort
	}

	err := a.db.Open(email)
	if err != nil {
		return User{}, fmt.Errorf("failed to open user database: %w", err)
	}

	ik, err := a.storeIdentityKey()
	if err != nil {
		return User{}, err
	}

	serializer := serialize.NewJSONSerializer()
	spk, err := a.storeSignedPreKey(ik, serializer)
	if err != nil {
		return User{}, err
	}

	preKeys, err := a.storePreKeys(serializer)
	if err != nil {
		return User{}, err
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

	a.apiClient.SetAuthorization(email, pwd)
	var resp api.SignUpResponse
	status, err := a.apiClient.Post(api.EndpointSignUp, payload, &resp)
	if err != nil {
		return User{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return User{}, fmt.Errorf("server returned error: %s", resp.Error)
	}

	a.signedIn = true
	user := User{
		ID:       resp.UserID,
		Username: email,
	}
	return user, nil
}

func (a *Auth) SignIn(email, pwd string) (User, error) {
	if !isValidEmail(email) {
		return User{}, ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return User{}, ErrAuthPwdTooShort
	}

	err := a.db.Open(email)
	if err != nil {
		return User{}, fmt.Errorf("failed to open user database: %w", err)
	}

	payload := api.SignInRequest{
		Username: email,
		Password: pwd,
	}

	a.apiClient.SetAuthorization(email, pwd)
	var resp api.SignInResponse
	status, err := a.apiClient.Post(api.EndpointSignIn, payload, &resp)
	if err != nil {

		return User{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return User{}, fmt.Errorf("server returned error: %s", resp.Error)
	}

	a.signedIn = true
	user := User{
		ID:       resp.UserID,
		Username: resp.Username,
	}
	return user, nil
}

func (a *Auth) SignOut() error {
	if !a.signedIn {
		panic("not signed in")
	}

	err := a.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	a.apiClient.ClearAuthorization()

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

func basicAuthorization(username, password string) string {
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return "Basic " + credentials
}
