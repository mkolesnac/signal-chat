package encryption

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"signal-chat/internal/api"
	"testing"
)

func TestUserService_Decrypt(t *testing.T) {
	t.Run("can decrypt encrypted message", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()

		senderDB := database.NewFake()
		err := senderDB.Open("alice")
		require.NoError(t, err)
		senderEncryptor := NewEncryptionManager(senderDB, ac)
		senderBundle, err := senderEncryptor.InitializeKeyStore()
		_ = senderBundle
		require.NoError(t, err)

		recipientDB := database.NewFake()
		err = recipientDB.Open("bob")
		require.NoError(t, err)
		recipientEncryptor := NewEncryptionManager(recipientDB, ac)
		recipientBundle, err := recipientEncryptor.InitializeKeyStore()
		require.NoError(t, err)
		ac.GetResponses[api.EndpointUserKeys("bob")] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body: mustMarshal(api.GetPrekeyBundleResponse{
				PreKeyBundle: api.PreKeyBundle{
					RegistrationID: 456,
					IdentityKey:    recipientBundle.IdentityKey,
					SignedPreKey: api.PreKey{
						ID:        recipientBundle.SignedPreKey.ID,
						PublicKey: recipientBundle.SignedPreKey.PublicKey,
					},
					SignedPreKeySignature: recipientBundle.SignedPreKey.Signature,
					PreKey: api.PreKey{
						ID:        recipientBundle.PreKeys[0].ID,
						PublicKey: recipientBundle.PreKeys[0].PublicKey,
					},
				},
			}),
		}

		wantPlaintext := []byte("Hello world!")
		encrypted, err := senderEncryptor.Encrypt(wantPlaintext, "bob")
		require.NoError(t, err)

		// Act
		decrypted, err := recipientEncryptor.Decrypt(encrypted.Serialized, "alice")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, string(wantPlaintext), string(decrypted.Plaintext))
	})
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return b
}
