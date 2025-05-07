package encryption

import "github.com/crossle/libsignal-protocol-go/protocol"

type EncryptedMessage struct {
	Serialized []byte
	Ciphertext []byte
	Envelope   *Envelope
}

type DecryptedMessage struct {
	Plaintext  []byte
	Ciphertext []byte
	Envelope   *Envelope
}

func newEncryptedMessage(ciphertextMsg protocol.CiphertextMessage) *EncryptedMessage {
	serialized := ciphertextMsg.(*protocol.SenderKeyMessage).SignedSerialize()

	return &EncryptedMessage{
		Serialized: serialized,
		Ciphertext: getCiphertext(ciphertextMsg),
		Envelope:   newEnvelope(ciphertextMsg),
	}
}

func newDecryptedMessage(plaintext []byte, ciphertextMsg protocol.CiphertextMessage) *DecryptedMessage {
	return &DecryptedMessage{
		Plaintext:  plaintext,
		Ciphertext: getCiphertext(ciphertextMsg),
		Envelope:   newEnvelope(ciphertextMsg),
	}
}

func getCiphertext(ciphertextMsg protocol.CiphertextMessage) []byte {
	if whisperMsg, ok := ciphertextMsg.(*protocol.SignalMessage); ok {
		return whisperMsg.Structure().CipherText
	} else if preKeyMsg, ok := ciphertextMsg.(*protocol.PreKeySignalMessage); ok {
		return preKeyMsg.WhisperMessage().Structure().CipherText
	} else if msg, ok := ciphertextMsg.(*protocol.SenderKeyMessage); ok {
		return msg.Ciphertext()
	}

	panic("Only preKey and Whisper messages are supported")
}
