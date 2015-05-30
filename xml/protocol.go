// Package xml implements the registry.Protocol interface,
// and registers it with the default registry.
package xml

import (
	"encoding/xml"
	"io"

	"github.com/teepark/represent"
)

func init() {
	represent.Register(&Protocol{})
}

// Protocol is the XML implementation of the registry.Protocol interface
type Protocol struct{}

// ContentType is needed for the registry.Protocol interface
func (p *Protocol) ContentType() string {
	return "application/xml"
}

// Decode is needed for the registry.Protocol interface
func (p *Protocol) Decode(container interface{}, r io.Reader) error {
	return xml.NewDecoder(r).Decode(container)
}

// Encode is needed for the registry.Protocol interface
func (p *Protocol) Encode(data interface{}, w io.Writer) error {
	return xml.NewEncoder(w).Encode(data)
}
