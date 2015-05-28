package represent

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func TestRegistryValue(t *testing.T) {
	clearRegistry()

	pi := &protImpl{}
	Register(pi)

	current := registry.Load()
	if current == nil {
		t.Fatal("not registered")
	}
	reg, ok := current.(currentRegistry)
	if !ok {
		t.Fatal("wrong type in registry")
	}

	if len(reg.protocols) != 1 {
		t.Fatal("wrong # of protocols registered:", len(reg.protocols))
	}

	if reg.protocols[0] != pi {
		t.Fatal("wrong protocol")
	}
}

func TestSetDefault(t *testing.T) {
	clearRegistry()

	pi := &protImpl{}
	Register(pi)
	SetDefault(pi.ContentType())

	current := registry.Load()
	if current == nil {
		t.Fatal("not registered")
	}
	reg, ok := current.(currentRegistry)
	if !ok {
		t.Fatal("wrong type in registry")
	}

	if reg.defaultProt != pi {
		t.Fatal("wrong default protocol")
	}
}

func TestRegisterBadCTPanics(t *testing.T) {
	if !panics(func() {
		Register(ctProt("application/"))
	}) {
		t.Fatal("registering bad content-type didn't panic")
	}
}

func panics(f func()) (p bool) {
	defer func() {
		r := recover()
		p = r != nil
	}()
	f()
	return
}

type protImpl struct{}

func (pi *protImpl) ContentType() string {
	return "text/plain"
}
func (pi *protImpl) Decode(data interface{}, r io.Reader) error {
	ioutil.ReadAll(r)
	return nil
}
func (pi *protImpl) Encode(data interface{}, w io.Writer) error {
	fmt.Fprint(w, "testing")
	return nil
}
