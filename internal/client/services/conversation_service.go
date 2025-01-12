package services

//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/crossle/libsignal-protocol-go/keys/prekey"
//	"github.com/crossle/libsignal-protocol-go/protocol"
//	"github.com/crossle/libsignal-protocol-go/session"
//	"signal-chat/client/clientapi"
//	"signal-chat/client/models"
//	"signal-chat/client/signal"
//)
//
//type ConversationService interface {
//	ListPreviews() ([]models.ConversationPreview, error)
//	ListMessages(conversationID string) ([]models.Message, error)
//	StartConversation(recipientID string) error
//	SendMessage(conversationID string) error
//}
//
//type conversationService struct {
//	db     database.Database
//	signal signal.Store
//	api    clientapi.KeyAPI
//}
//
//func (c *conversationService) ListPreviews() ([]models.ConversationPreview, error) {
//	items, err := c.db.QueryValues("preview#")
//	if err != nil {
//		return nil, fmt.Errorf("failed to query conversation previews: %w", err)
//	}
//
//	var previews []models.ConversationPreview
//	for _, item := range items {
//		var preview models.ConversationPreview
//		if err := json.Unmarshal(item, &preview); err != nil {
//			return nil, fmt.Errorf("failed to unmarshal conversation preview: %w", err)
//		}
//	}
//
//	return previews, nil
//}
//
//func (c *conversationService) ListMessages(conversationID string) ([]models.Message, error) {
//	prefix := fmt.Sprintf("conversation#%s|message#", conversationID)
//	items, err := c.db.QueryValues(prefix)
//	if err != nil {
//		return nil, fmt.Errorf("failed to query conversation messages: %w", err)
//	}
//
//	var messages []models.Message
//	for _, item := range items {
//		var msg models.Message
//		if err := json.Unmarshal(item, &msg); err != nil {
//			return nil, fmt.Errorf("failed to unmarshal conversation message: %w", err)
//		}
//	}
//
//	return messages, nil
//}
//
//func (c *conversationService) StartConversation(recipientID, text string) error {
//	addr := protocol.NewSignalAddress(recipientID, 0)
//	sessionBuilder := session.NewBuilder(
//		c.signal,
//		c.signal,
//		c.signal,
//		c.signal,
//		addr,
//		c.signal.Serializer(),
//	)
//
//	resp, err := c.api.GetKeyBundle(recipientID)
//	if err != nil {
//		return fmt.Errorf("failed to get key bundle: %w", err)
//	}
//
//	// Build a session with a PreKey retrieved from the server.
//	bundle := prekey.NewBundle()
//	err = sessionBuilder.ProcessBundle(bundle)
//	if err != nil {
//		return fmt.Errorf("failed to process bundle: %w", err)
//	}
//
//	cipher := session.NewCipher(sessionBuilder, addr)
//	msg, err := cipher.Encrypt([]byte(text))
//	if err != nil {
//		return fmt.Errorf("failed to encrypt message: %w", err)
//	}
//	bytes := msg.Serialize()
//
//	m := protocol.NewSignalMessageFromBytes()
//}
//
//func (c *conversationService) SendMessage(conversationID, recipientID, text string) error {
//	//TODO implement me
//	panic("implement me")
//}
