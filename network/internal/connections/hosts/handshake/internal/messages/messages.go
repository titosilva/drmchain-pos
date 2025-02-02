package messages

type MessageShell struct {
	Cmd       string
	Data      []byte
	Signature []byte
}

type HelloMessage struct {
	SrcTag  string
	DstTag  string
	SrcAddr string
	Nonce   []byte
}

type ChallengeMessage struct {
	ChallengeNonce []byte
	Nonce          []byte
}

type AnswerMessage struct {
	EphKey         []byte
	ChallengeNonce []byte
	AcceptNonce    []byte
}

type AcceptedMessage struct {
	TcpAddr         string
	SecretSignature []byte
	SessionId       string
	AcceptNonce     []byte
}
