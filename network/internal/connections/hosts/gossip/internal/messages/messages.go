package messages

const (
	SealTypeControl = "control"
	SealTypeData    = "data"
)

type MessageSeal struct {
	Type      string
	SessionId string
	Mac       []byte // HMAC of the data together with the sequence number
	Sequence  int
	Encrypted []byte
}

type ControlMessage struct {
	Succeed    bool
	ErrorType  string
	MessageSeq int
}

const (
	ErrorTypeUnreadableSeal = "unreadable_seal"
	ErrorTypeWrongSession   = "wrong_session"
	ErrorTypeWrongSeq       = "wrong_sequence"
	ErrorTypeInvalidSeal    = "invalid_seal"
	ErrorTypeInternalError  = "internal_error"
)
