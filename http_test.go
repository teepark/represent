package represent

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncodeResponse(t *testing.T) {
	for i, spec := range testTable {
		reg := NewRegistry()
		for _, ct := range spec.registeredCTs {
			reg.Register(ctProt(ct))
		}

		if spec.defaultCT != "" {
			reg.SetDefault(spec.defaultCT)
		}

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code, err := reg.EncodeResponse(struct {
				Foo string
				Bar int
				Baz bool
			}{
				"foo",
				8,
				false,
			}, r, w)
			if err != nil {
				w.WriteHeader(code)
			}
		}))

		req, err := http.NewRequest("GET", s.URL+"/", nil)
		if err != nil {
			t.Fatalf("(%d) new request: %s", i, err)
		}
		req.Header.Set("Accept", spec.header)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("(%d) client.Do: %s", i, err)
		}
		ct := resp.Header.Get("Content-Type")
		if ct != spec.resultCT {
			t.Fatalf("(%d) wrong response content-type. should be %s, got %s", i, spec.resultCT, ct)
		}
		s.Close()
	}
}

func TestDecodeRequest(t *testing.T) {
	contentTypes := []string{
		"application/msgpack",
		"application/json",
		"text/html",
		"application/xml",
	}

	reg := NewRegistry()
	for _, ct := range contentTypes {
		reg.Register(ctProt(ct))
	}
	reg.SetDefault("application/json")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 256)
		code, err := reg.DecodeRequest(&b, r)
		if err != nil {
			t.Fatal(err)
		}
		if code != 200 {
			t.Fatalf("got code %d", code)
		}

		if string(b) != fmt.Sprintf("(encoded %s)", r.Header.Get("ct")) {
			t.Fatalf(
				"mismatched content. expected '%s', got '%s'",
				fmt.Sprintf("(encoded %s)", r.Header.Get("ct")),
				string(b),
			)
		}
	})

	s := httptest.NewServer(handler)

	for _, ct := range contentTypes {
		body := &bytes.Buffer{}
		ctProt(ct).Encode(nil, body)
		req, err := http.NewRequest("POST", s.URL+"/", body)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", ct)
		req.Header.Set("ct", ct)
		http.DefaultClient.Do(req)
	}
	s.Close()
}
