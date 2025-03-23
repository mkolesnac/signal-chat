package encryption

import "github.com/crossle/libsignal-protocol-go/protocol"

type Envelope struct {
	MessageType      uint32
	RatchetKey       []byte
	Counter          uint32
	PreviousCounter  uint32
	Version          int
	Mac              []byte
	SenderRatchetKey []byte
	RegistrationID   uint32
	SignedPreKeyID   uint32
	PreKeyID         uint32
	IdentityKey      []byte
	BaseKey          []byte
}

func newEnvelope(ciphertextMsg protocol.CiphertextMessage) *Envelope {
	if whisperMsg, ok := ciphertextMsg.(*protocol.SignalMessage); ok {
		return newEnvelopeFromSignalMessage(whisperMsg)
	}

	if preKeyMsg, ok := ciphertextMsg.(*protocol.PreKeySignalMessage); ok {
		envelope := newEnvelopeFromSignalMessage(preKeyMsg.WhisperMessage())
		envelope.MessageType = preKeyMsg.Type()
		envelope.RegistrationID = preKeyMsg.RegistrationID()
		envelope.SignedPreKeyID = preKeyMsg.SignedPreKeyID()
		envelope.PreKeyID = preKeyMsg.PreKeyID().Value
		baseKey := preKeyMsg.BaseKey().PublicKey()
		envelope.BaseKey = baseKey[:]
		identityKey := preKeyMsg.IdentityKey().PublicKey().PublicKey()
		envelope.IdentityKey = identityKey[:]
		return envelope
	}

	panic("Received not supported message type")
}

func newEnvelopeFromSignalMessage(signalMsg *protocol.SignalMessage) *Envelope {
	structure := signalMsg.Structure()
	senderRatchetKey := signalMsg.SenderRatchetKey().PublicKey()
	return &Envelope{
		MessageType:      signalMsg.Type(),
		RatchetKey:       structure.RatchetKey,
		Counter:          structure.Counter,
		PreviousCounter:  structure.PreviousCounter,
		Version:          structure.Version,
		Mac:              structure.Mac,
		SenderRatchetKey: senderRatchetKey[:],
	}
}
