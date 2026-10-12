package worker

type OutboundMsgType int8

const (
	MessageTypeFinished OutboundMsgType = iota
	MessageTypeError
)

type OutboundMsg struct {
	Type  OutboundMsgType
	Value any
}
