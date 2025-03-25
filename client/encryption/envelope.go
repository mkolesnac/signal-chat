package encryption

import "github.com/crossle/libsignal-protocol-go/protocol"

type Envelope struct {
	KeyID     uint32
	Version   int
	Iteration uint32
	Signature []byte
}

func newEnvelope(ciphertextMsg protocol.CiphertextMessage) *Envelope {
	if keyMsg, ok := ciphertextMsg.(*protocol.SenderKeyMessage); ok {
		signature := keyMsg.Signature()
		return &Envelope{
			KeyID:     keyMsg.KeyID(),
			Iteration: keyMsg.Iteration(),
			Version:   int(keyMsg.Version()),
			Signature: signature[:],
		}
	}

	panic("Received not supported message type")
}
