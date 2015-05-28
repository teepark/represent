package represent

import (
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

type acceptGroup struct {
	starQ    float64
	subTypes map[string]float64
}

type acceptSpec struct {
	starQ float64
	types map[string]*acceptGroup
}

// Match parses the Accept header of a request and selects the most suitable
// registered Protocol. It will return errors resulting from a malformed header
// (400 would be an appropriate response), or a nil Protocol if nothing matches
// ("406 Not Acceptable").
func Match(r *http.Request) (Protocol, error) {
	return globalReg.Match(r)
}

// TODO: cache acceptSpecs by precise fullHeader strings

func buildSpec(fullHeader string) (*acceptSpec, error) {
	spec := new(acceptSpec)
	spec.types = make(map[string]*acceptGroup, 0)

	if fullHeader == "" {
		fullHeader = "*/*"
	}
	split := strings.Split(fullHeader, ",")

	for _, header := range split {
		mt, params, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, err
		}

		q, err := qval(params)
		if err != nil {
			return nil, err
		}

		if mt == "*/*" {
			spec.starQ = q
			continue
		}

		types := strings.SplitN(mt, "/", 2)
		major, minor := types[0], ""
		if len(types) == 2 {
			minor = types[1]
		}

		group, ok := spec.types[major]
		if !ok {
			group = &acceptGroup{0, make(map[string]float64)}
			spec.types[major] = group
		}
		if minor == "*" {
			group.starQ = q
		} else {
			group.subTypes[minor] = q
		}
	}

	return spec, nil
}

func qval(params map[string]string) (float64, error) {
	qstr, ok := params["q"]
	if !ok {
		return 1, nil
	}

	q, err := strconv.ParseFloat(qstr, 64)
	if err != nil {
		return 0, err
	}
	if q > 1 {
		return 0, errors.New("qval must not be >1")
	}
	return q, nil
}

// Match on a registry performs the same operation as the Match function, just
// matches against the set of Protocols registered on the specific registry.
func (reg *Registry) Match(r *http.Request) (Protocol, error) {
	spec, err := buildSpec(r.Header.Get("Accept"))
	if err != nil {
		return nil, err
	}

	registered := reg.container.Load()
	if registered == nil {
		return nil, nil
	}

	type protMatch struct {
		prot Protocol
		q    float64
	}
	matches := make([]protMatch, 0)

	for _, prot := range registered.(currentRegistry).protocols {
		split := strings.SplitN(prot.ContentType(), "/", 2)
		major, minor := split[0], ""
		if len(split) > 1 {
			minor = split[1]
		}

		group, ok := spec.types[major]
		if !ok {
			// nothing registered for the major type
			if spec.starQ > 0 {
				// but I have a */*
				matches = append(matches, protMatch{prot, spec.starQ})
			}
			continue
		}

		qval, ok := group.subTypes[minor]
		if !ok {
			// nothing registered for the minor type
			if group.starQ > 0 {
				// but I have a major/*
				matches = append(matches, protMatch{prot, group.starQ})
			} else if spec.starQ > 0 {
				// there was a */* and something else for the same major type,
				// but no precise major/minor match
				matches = append(matches, protMatch{prot, spec.starQ})
			}
			continue
		}

		// major/minor match
		matches = append(matches, protMatch{prot, qval})
	}

	// scan for the strongest match by qval
	var (
		maxQ      float64
		bestProts []Protocol
	)
	for _, match := range matches {
		if match.q > maxQ {
			maxQ = match.q
			bestProts = []Protocol{match.prot}
		} else if match.q == maxQ {
			bestProts = append(bestProts, match.prot)
		}
	}

	// nobody meets the criteria
	if bestProts == nil {
		return nil, nil
	}

	// see if the default one is among the matches
	for _, prot := range bestProts {
		if prot == registered.(currentRegistry).defaultProt {
			return prot, nil
		}
	}

	// just pick the first one
	return bestProts[0], nil
}
