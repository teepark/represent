package represent

import (
	"fmt"
	"io"
	"mime"
	"sync"
	"sync/atomic"
)

var (
	writeMux sync.Mutex
	registry atomic.Value
)

type currentRegistry struct {
	protocols   []Protocol
	defaultProt Protocol
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

// Set a Protocol implementation as the handler for its content type.
func Register(p Protocol) {
	if _, _, err := mime.ParseMediaType(p.ContentType()); err != nil {
		panic(err)
	}

	writeMux.Lock()

	current := registry.Load()
	var next currentRegistry
	if current == nil {
		next = currentRegistry{[]Protocol{p}, p}
	} else {
		next = currentRegistry{
			append(current.(currentRegistry).protocols, p),
			current.(currentRegistry).defaultProt,
		}
	}

	registry.Store(next)

	writeMux.Unlock()
}

// TODO: document
func SetDefault(contentType string) {
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		panic(err)
	}

	writeMux.Lock()
	defer writeMux.Unlock()

	current := registry.Load()
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

	registry.Store(currentRegistry{cur.protocols, d})
}
