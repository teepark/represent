package represent

import (
	"fmt"
	"io"
	"mime"
	"sync"
	"sync/atomic"
)

var globalReg Registry

type currentRegistry struct {
	protocols   []Protocol
	defaultProt Protocol
}

// Registry is a container for Protocols
type Registry struct {
	lock      sync.RWMutex
	container atomic.Value
}

// Protocol is the plug-in interface for a specific serialization format
// implementation. These can be created and registered with the global registry
// to make them eligible for selection based on request Accept headers.
type Protocol interface {
	// ContentType returns the content-type this Protocol handles
	ContentType() string

	// Decode reads content from a Reader and deserializes it into a container
	Decode(interface{}, io.Reader) error

	// Encode serializes an object to a Writer
	Encode(interface{}, io.Writer) error
}

// Register stores a Protocol as a content-type handler on a registry instance
func (reg *Registry) Register(p Protocol) {
	if _, _, err := mime.ParseMediaType(p.ContentType()); err != nil {
		panic(err)
	}

	reg.lock.Lock()

	current := reg.container.Load()
	var next currentRegistry
	if current == nil {
		next = currentRegistry{[]Protocol{p}, p}
	} else {
		next = currentRegistry{
			append(current.(currentRegistry).protocols, p),
			current.(currentRegistry).defaultProt,
		}
	}

	reg.container.Store(next)

	reg.lock.Unlock()
}

// SetDefault sets the default content type for a specific registry.
func (reg *Registry) SetDefault(contentType string) {
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		panic(err)
	}

	reg.lock.Lock()
	defer reg.lock.Unlock()

	current := reg.container.Load()
	if current == nil {
		panic(fmt.Sprintf("no protocol registered for '%s'", contentType))
	}
	cur := current.(currentRegistry)

	var d Protocol
	for _, prot := range cur.protocols {
		if prot.ContentType() == contentType {
			d = prot
		}
	}
	if d == nil {
		panic(fmt.Sprintf("no protocol registered for '%s'", contentType))
	}

	reg.container.Store(currentRegistry{cur.protocols, d})
}

// Register sets a Protocol implementation as the handler for its content type.
func Register(p Protocol) {
	globalReg.Register(p)
}

// SetDefault sets the content type to prefer in the event of match ties
// (especially because the Accept header contained */*).
func SetDefault(contentType string) {
	globalReg.SetDefault(contentType)
}
