package models

import (
	"fmt"
	"signal-chat/cmd/server/storage"
	"time"
)

type Message struct {
	storage.TableItem
	SenderID   string `dynamodbav:"senderId"`
	CipherText string `dynamodbav:"cipherText"`
}

func NewMessage(recipientId, senderId, cipherText string) *Message {
	return &Message{
		TableItem: storage.TableItem{
			PartitionKey: MessagePartitionKey(recipientId),
			SortKey:      MessageSortKey(time.Now().UnixNano()),
			CreatedAt:    getTimestamp(),
		},
		SenderID:   senderId,
		CipherText: cipherText,
	}
}

func (m *Message) GetPartitionKey() string {
	return m.PartitionKey
}

func (m *Message) GetSortKey() string {
	return m.SortKey
}

func MessagePartitionKey(recipientId string) string { return AccountPartitionKey(recipientId) }

func MessageSortKey(timestamp int64) string {
	return fmt.Sprintf("msg#%d", timestamp)
}
