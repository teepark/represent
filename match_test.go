package represent

import (
	"io"
	"sync/atomic"
	"testing"
)

type testSpec struct {
	header        string
	registeredCTs []string
	defaultCT     string
	resultCT      string
}

var testTable = []testSpec{
	{ // trigger the default with no header
		header: "",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "application/json",
		resultCT:  "application/json",
	},
	{ // trigger the first one registered with no header
		header: "",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "",
		resultCT:  "application/msgpack",
	},
	{ // trigger the one in the accept header
		header: "application/yaml",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "application/json",
		resultCT:  "application/yaml",
	},
	{ // trigger the default with * / *
		header: "*/*",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "application/json",
		resultCT:  "application/json",
	},
	{ // trigger the highest q value (1 default)
		header: "application/msgpack;q=0.7, application/yaml",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "application/json",
		resultCT:  "application/yaml",
	},
	{ // trigger the highest q value
		header: "application/msgpack;q=0.7, application/yaml;q=0.4",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "application/json",
		resultCT:  "application/msgpack",
	},
	{ // closest match takes precedence
		header: "text/*;q=0.9, application/yaml;q=0.6, text/html;q=0.5",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
			"text/html",
		},
		defaultCT: "application/json",
		resultCT:  "application/yaml",
	},
	{ // default breaks a tie
		header: "*/*, application/json;q=0.8",
		registeredCTs: []string{
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "application/yaml",
		resultCT:  "application/yaml",
	},
	{ // matches without a minor type
		header: "foo",
		registeredCTs: []string{
			"foo",
			"bar/baz",
		},
		defaultCT: "bar/baz",
		resultCT:  "foo",
	},
	{ // earliest match wins when default ruled out
		header: "application/*,text/html;q=0.5",
		registeredCTs: []string{
			"text/html",
			"application/msgpack",
			"application/json",
			"application/yaml",
		},
		defaultCT: "text/html",
		resultCT:  "application/msgpack",
	},
}

func TestMatching(t *testing.T) {
	for _, test := range testTable {
		clearRegistry()

		for _, ct := range test.registeredCTs {
			Register(ctProt(ct))
		}
		if test.defaultCT != "" {
			SetDefault(test.defaultCT)
		}

		prot, err := Match(test.header)
		if err != nil {
			t.Fatal("Match failure:", err)
		}

		if prot.ContentType() != test.resultCT {
			t.Fatalf(
				"mismatched contentType, expected '%s', got '%s'",
				test.resultCT,
				prot.ContentType(),
			)
		}
	}
}

func clearRegistry() {
	globalReg.lock.Lock()
	defer globalReg.lock.Unlock()

	globalReg.container = atomic.Value{}
}

type ctProt string

func (c ctProt) ContentType() string {
	return string(c)
}
func (c ctProt) Decode(data interface{}, r io.Reader) error {
	return nil
}
func (c ctProt) Encode(data interface{}, w io.Writer) error {
	return nil
}
