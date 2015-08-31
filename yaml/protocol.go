// Package yaml implements the registry.Protocol interface,
// and registers it with the default registry.
package yaml

import (
	"io"
	"io/ioutil"

	"github.com/teepark/represent"

	"gopkg.in/yaml.v2"
)

func init() {
	represent.Register(&Protocol{})
}

// Protocol is the YAML implementation of the registry.Protocol interface
type Protocol struct{}

// ContentType is needed for the registry.Protocol interface
func (p *Protocol) ContentType() string {
	return "application/yaml"
}

// Decode is needed for the registry.Protocol interface
func (p *Protocol) Decode(container interface{}, r io.Reader) error {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, container)
}

// Encode is needed for the registry.Protocol interface
func (p *Protocol) Encode(data interface{}, w io.Writer) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	return err
}
