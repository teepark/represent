package represent

import (
	"io"
	"mime"
	"sync"

	"github.com/golang/groupcache/lru"
)

const cacheSize = 256

var globalReg = NewRegistry()

// Registry is a container for Protocols
type Registry struct {
	mut             sync.RWMutex
	protocols       []Protocol
	defaultProtocol Protocol

	cacheLock sync.Mutex
	specCache *lru.Cache
}

// NewRegistry creates a new, empty Registry
func NewRegistry() *Registry {
	return &Registry{
		specCache: lru.New(cacheSize),
	}
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

// Register stores a Protocol as a content-type handler on a registry instance.
func (reg *Registry) Register(p Protocol) {
	if _, _, err := mime.ParseMediaType(p.ContentType()); err != nil {
		panic(err)
	}

	reg.mut.Lock()
	reg.protocols = append(reg.protocols, p)
	reg.mut.Unlock()
}

// SetDefault sets the default content type for a specific registry.
func (reg *Registry) SetDefault(contentType string) {
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		panic(err)
	}

	reg.mut.Lock()
	for _, p := range reg.protocols {
		if p.ContentType() == contentType {
			reg.defaultProtocol = p
			break
		}
	}
	reg.mut.Unlock()
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

func (reg *Registry) checkCache(header string) *acceptSpec {
	reg.cacheLock.Lock()

	item, ok := reg.specCache.Get(header)

	reg.cacheLock.Unlock()

	if ok {
		return item.(*acceptSpec)
	}
	return nil
}

func (reg *Registry) storeCache(header string, spec *acceptSpec) {
	reg.cacheLock.Lock()

	reg.specCache.Add(header, spec)

	reg.cacheLock.Unlock()
}
