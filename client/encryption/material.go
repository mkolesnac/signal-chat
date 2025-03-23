package encryption

type Material struct {
	RootKey                []byte
	SenderChainKey         []byte
	ReceiverChainKey       []byte
	SenderRatchetKey       []byte
	PreviousMessageCounter uint32
	SessionVersion         int
	MessageKeys            MessageKeys
}

type MessageKeys struct {
	CipherKey []byte
	MacKey    []byte
	IV        []byte
	Index     uint32
}
