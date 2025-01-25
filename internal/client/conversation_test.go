package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConversation_Serialize(t *testing.T) {
	t.Run("roundtrip preserves conversation data", func(t *testing.T) {
		// Arrange
		original := Conversation{
			ID:                  "abc",
			LastMessagePreview:  "Hello there",
			LastMessageSenderID: "user#123",
		}

		// Act
		serialized, err := original.Serialize()

		// Assert
		assert.NoError(t, err)
		deserialized, err := DeserializeConversation(serialized)
		assert.NoError(t, err)
		assert.EqualExportedValuesf(t, original, deserialized, "serialized and deserialized conversations should match, got: %s, want: %s", deserialized, original)
	})
}

func TestConversation_Deserialize(t *testing.T) {
	t.Run("returns error for invalid data", func(t *testing.T) {
		// Arrange
		invalidData := []byte("invalid data")

		// Act
		_, err := DeserializeConversation(invalidData)

		// Assert
		assert.Error(t, err)
	})
}
