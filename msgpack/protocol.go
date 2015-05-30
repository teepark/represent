// Package msgpack implements the registry.Protocol interface,
// and registers it with the default registry.
package msgpack

import (
	"github.com/ugorji/go/codec"
	"io"

	"github.com/teepark/represent"
)

var mpHandle = new(codec.MsgpackHandle)

func init() {
	represent.Register(&Protocol{})
}

// Protocol is the MsgPack implementation of the registry.Protocol interface
type Protocol struct{}

// ContentType is needed for the registry.Protocol interface
func (p *Protocol) ContentType() string {
	return "application/msgpack"
}

// Decode is needed for the registry.Protocol interface
func (p *Protocol) Decode(container interface{}, r io.Reader) error {
	return codec.NewDecoder(r, mpHandle).Decode(container)
}

// Encode is needed for the registry.Protocol interface
func (p *Protocol) Encode(data interface{}, w io.Writer) error {
	return codec.NewEncoder(w, mpHandle).Encode(data)
}
