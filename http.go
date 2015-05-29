package represent

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

// ErrNoMatch is for when no suitable protocol is found
var ErrNoMatch = errors.New("no protocol matched the request")

// Decode uses an appropriate protocol to decode an http request body into a
// container. It returns a suggested HTTP response code and an error (if the
// error is nil the response code will be StatusOK, 200).
func Decode(container interface{}, r *http.Request) (int, error) {
	return globalReg.Decode(container, r)
}

// Encode will serialize data with a format acceptable by the HTTP client
// (using the Accept header of the request). If there is an error it will NOT
// send the response but will instead return a suggested HTTP response code and
// the error. If it returns (200, nil) then the data has already been written
// to the response.
func Encode(data interface{}, r *http.Request, w http.ResponseWriter) (int, error) {
	return globalReg.Encode(data, r, w)
}

// Decode as a Registry method performs the same job as the global function,
// but using the protocols registered with a specific registry.
func (reg *Registry) Decode(container interface{}, r *http.Request) (int, error) {
	protocol, err := reg.Match(r.Header.Get("Content-Type"))
	if err != nil {
		return http.StatusBadRequest, err
	}

	if protocol == nil {
		return http.StatusUnsupportedMediaType, ErrNoMatch
	}

	err = protocol.Decode(container, r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

// Encode performs the same job as the global function, but matches against the
// set of Protocols registered with the specific registry.
func (reg *Registry) Encode(data interface{}, r *http.Request, w http.ResponseWriter) (int, error) {
	protocol, err := reg.Match(r.Header.Get("Accept"))
	if err != nil {
		return http.StatusBadRequest, err
	}

	if protocol == nil {
		return http.StatusNotAcceptable, ErrNoMatch
	}

	b := new(bytes.Buffer)
	err = protocol.Encode(data, b)
	if err != nil {
		return http.StatusBadRequest, err
	}

	w.Header().Set("Content-Type", protocol.ContentType())
	w.WriteHeader(http.StatusOK)
	io.Copy(w, b)
	return http.StatusOK, nil
}
