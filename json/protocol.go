// Package json implements the registry.Protocol interface,
// and registers it with the default registry.
package json

import (
	"encoding/json"
	"io"

	"github.com/teepark/represent"
)

func init() {
	represent.Register(&Protocol{})
}

// Protocol is the JSON implementation of the registry.Protocol interface
type Protocol struct{}

// ContentType is needed for the registry.Protocol interface
func (p *Protocol) ContentType() string {
	return "application/json"
}

// Decode is needed for the registry.Protocol interface
func (p *Protocol) Decode(container interface{}, r io.Reader) error {
	return json.NewDecoder(r).Decode(container)
}

// Encode is needed for the registry.Protocol interface
func (p *Protocol) Encode(data interface{}, w io.Writer) error {
	return json.NewEncoder(w).Encode(data)
}
