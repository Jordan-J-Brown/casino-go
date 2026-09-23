package natsx

import (
	"encoding/json"
	"fmt"

	"github.com/Jordan-J-Brown/casino-go/pkg/mimex"
	"google.golang.org/protobuf/proto"
)

// Decodes the raw message data into [Msg.Data].
//
// See [AutoDecoder].
type Decoder[T any] interface {
	Decode(msg *Msg[T]) error
}

type DecodeFunc[T any] func(msg *Msg[T]) error

func (fn DecodeFunc[T]) Decode(msg *Msg[T]) error {
	return fn(msg)
}

// Returns a decoder that will automatically decode to JSON if it detects the content type, otherwise Protobuf will be.
func AutoDecoder[T any]() DecodeFunc[T] {
	return func(msg *Msg[T]) (err error) {
		if isJson(msg.Data) {
			msg.ContentType = mimex.ContentTypeJSON
			return json.Unmarshal(msg.Data, &msg.Body)
		}

		protoMsg, ok := any(&msg.Body).(proto.Message)
		if !ok {
			err = fmt.Errorf("unexpected type for proto decode: %T", msg.Body)
		}

		msg.ContentType = mimex.ContentTypeProtobuf

		return proto.Unmarshal(msg.Data, protoMsg)
	}
}

// TODO: implement properly
func isJson(data []byte) bool {
	if size := len(data); size >= 2 && data[0] == '{' && data[size-1] == '}' {
		return true
	}
	return false
}
