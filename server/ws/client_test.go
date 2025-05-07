package ws

import (
	"encoding/json"
	"signal-chat/internal/apitypes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SendMessage(t *testing.T) {
	t.Run("should send message and receive ACK", func(t *testing.T) {
		// Arrange
		fakeConn := NewFakeWebSocketConn()
		fakeStore := NewFakeMessageStore()
		client := NewClient("test-client", fakeConn, fakeStore)

		// Act
		message := &apitypes.WSMessage{
			ID:   "msg-123",
			Type: apitypes.MessageTypeNewMessage,
			Data: json.RawMessage(`{"content": "Hello"}`),
		}
		err := client.SendMessage(message)
		require.NoError(t, err)

		// Wait for message to be sent
		time.Sleep(100 * time.Millisecond)

		// get the sent message
		var sentMsg apitypes.WSMessage
		select {
		case msgBytes := <-fakeConn.writeChan:
			err := json.Unmarshal(msgBytes, &sentMsg)
			require.NoError(t, err)
		default:
			t.Fatal("No message was sent")
		}

		// Send ACK back
		ackMsg := apitypes.WSMessage{
			ID:   sentMsg.ID,
			Type: apitypes.MessageTypeAck,
		}
		ackBytes, err := json.Marshal(ackMsg)
		require.NoError(t, err)
		fakeConn.readChan <- ackBytes

		// Wait for ACK to be processed
		time.Sleep(100 * time.Millisecond)

		// Assert
		// Check that message is not in storage (since it was ACKed)
		messages, err := fakeStore.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, messages, "Message should not be in storage after ACK")
	})

	t.Run("should store message when ACK is not received", func(t *testing.T) {
		// Arrange
		fakeConn := NewFakeWebSocketConn()
		fakeStore := NewFakeMessageStore()
		client := NewClient("test-client", fakeConn, fakeStore)
		client.readWait = 100 * time.Millisecond

		// Act
		message := &apitypes.WSMessage{
			ID:   "msg-456",
			Type: apitypes.MessageTypeNewMessage,
			Data: json.RawMessage(`{"content": "Hello"}`),
		}
		err := client.SendMessage(message)
		require.NoError(t, err)

		// Wait for message to be sent
		time.Sleep(100 * time.Millisecond)

		// get the sent message
		var sentMsg apitypes.WSMessage
		select {
		case msgBytes := <-fakeConn.writeChan:
			err := json.Unmarshal(msgBytes, &sentMsg)
			require.NoError(t, err)
		default:
			t.Fatal("No message was sent")
		}

		// Wait for ACK timeout
		time.Sleep(200 * time.Millisecond)

		// Assert
		// Check that message is in storage (since ACK was not received)
		messages, err := fakeStore.LoadAll()
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, sentMsg.ID, messages[0].ID)
		assert.Equal(t, sentMsg.Type, messages[0].Type)
		assert.JSONEq(t, string(sentMsg.Data), string(messages[0].Data))
	})

	t.Run("should store message when connection is closed", func(t *testing.T) {
		// Arrange
		fakeConn := NewFakeWebSocketConn()
		fakeStore := NewFakeMessageStore()
		client := NewClient("test-client", fakeConn, fakeStore)

		// Act
		// Close the connection
		client.Close()

		// Try to send a message
		message := &apitypes.WSMessage{
			ID:   "msg-789",
			Type: apitypes.MessageTypeNewMessage,
			Data: json.RawMessage(`{"content": "Hello"}`),
		}
		err := client.SendMessage(message)

		// Assert
		assert.NoError(t, err)

		// Check that message is in storage
		messages, err := fakeStore.LoadAll()
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, apitypes.MessageTypeNewMessage, messages[0].Type)
		assert.Equal(t, `{"content": "Hello"}`, string(messages[0].Data))
	})
}

func TestClient_SyncClient(t *testing.T) {
	t.Run("should sync stored messages to client", func(t *testing.T) {
		// Arrange
		fakeConn := NewFakeWebSocketConn()
		fakeStore := NewFakeMessageStore()

		// Store some messages
		storedMessages := []*apitypes.WSMessage{
			{
				ID:   "msg1",
				Type: apitypes.MessageTypeNewMessage,
				Data: json.RawMessage(`{"content": "Hello"}`),
			},
			{
				ID:   "msg2",
				Type: apitypes.MessageTypeNewConversation,
				Data: json.RawMessage(`{"conversationId": "conv1"}`),
			},
		}
		err := fakeStore.Store(storedMessages)
		require.NoError(t, err)

		// Act
		_ = NewClient("test-client", fakeConn, fakeStore)

		// Wait for sync to complete
		time.Sleep(100 * time.Millisecond)

		// get the sync message
		var syncMsg apitypes.WSMessage
		select {
		case msgBytes := <-fakeConn.writeChan:
			err := json.Unmarshal(msgBytes, &syncMsg)
			require.NoError(t, err)
		default:
			t.Fatal("No sync message was sent")
		}

		// Assert
		assert.Equal(t, apitypes.MessageTypeSync, syncMsg.Type)

		var syncPayload apitypes.WSSyncPayload
		err = json.Unmarshal(syncMsg.Data, &syncPayload)
		require.NoError(t, err)

		require.Len(t, syncPayload.Messages, 2)
		assert.Equal(t, "msg1", syncPayload.Messages[0].ID)
		assert.Equal(t, apitypes.MessageTypeNewMessage, syncPayload.Messages[0].Type)
		assert.JSONEq(t, `{"content": "Hello"}`, string(syncPayload.Messages[0].Data))

		assert.Equal(t, "msg2", syncPayload.Messages[1].ID)
		assert.Equal(t, apitypes.MessageTypeNewConversation, syncPayload.Messages[1].Type)
		assert.JSONEq(t, `{"conversationId": "conv1"}`, string(syncPayload.Messages[1].Data))
	})

	t.Run("should delete messages after successful sync ACK", func(t *testing.T) {
		// Arrange
		fakeConn := NewFakeWebSocketConn()
		fakeStore := NewFakeMessageStore()

		// Store some messages
		storedMessages := []*apitypes.WSMessage{
			{
				ID:   "msg1",
				Type: apitypes.MessageTypeNewMessage,
				Data: json.RawMessage(`{"content": "Hello"}`),
			},
		}
		err := fakeStore.Store(storedMessages)
		require.NoError(t, err)

		// Act
		_ = NewClient("test-client", fakeConn, fakeStore)

		// Wait for sync to complete
		time.Sleep(100 * time.Millisecond)

		// get the sync message
		var syncMsg apitypes.WSMessage
		select {
		case msgBytes := <-fakeConn.writeChan:
			err := json.Unmarshal(msgBytes, &syncMsg)
			require.NoError(t, err)
		default:
			t.Fatal("No sync message was sent")
		}

		// Send ACK back
		ackMsg := apitypes.WSMessage{
			ID:   syncMsg.ID,
			Type: apitypes.MessageTypeAck,
		}
		ackBytes, err := json.Marshal(ackMsg)
		require.NoError(t, err)
		fakeConn.readChan <- ackBytes

		// Wait for ACK to be processed
		time.Sleep(100 * time.Millisecond)

		// Assert
		// Check that messages are deleted from storage
		messages, err := fakeStore.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, messages, "Messages should be deleted after sync ACK")
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("should store not ACKed messages when closed", func(t *testing.T) {
		// Arrange
		fakeConn := NewFakeWebSocketConn()
		fakeStore := NewFakeMessageStore()
		client := NewClient("test-client", fakeConn, fakeStore)

		// Send a message
		message := &apitypes.WSMessage{
			ID:   "msg-close-123",
			Type: apitypes.MessageTypeNewMessage,
			Data: json.RawMessage(`{"content": "Hello"}`),
		}
		err := client.SendMessage(message)
		client.readWait = 10 * time.Millisecond
		require.NoError(t, err)

		// Wait for message to be sent
		time.Sleep(100 * time.Millisecond)

		// Act
		client.Close()

		// Assert
		// Check that message is in storage
		messages, err := fakeStore.LoadAll()
		require.NoError(t, err)
		require.Len(t, messages, 1)
		assert.Equal(t, apitypes.MessageTypeNewMessage, messages[0].Type)
		assert.Equal(t, `{"content": "Hello"}`, string(messages[0].Data))
	})
}
