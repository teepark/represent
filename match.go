package represent

import (
	"errors"
	"mime"
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

// Match parses a mimetype header (like Accept) and selects the most suitable
// registered Protocol. It will return errors resulting from a malformed header
// or a nil Protocol if nothing matches.
func Match(header string) (Protocol, error) {
	return globalReg.Match(header)
}

func (reg *Registry) buildSpec(fullHeader string) (*acceptSpec, error) {
	// per the RFC, the lack of an accept header says they accept anything
	if fullHeader == "" {
		fullHeader = "*/*"
	}

	// check the cache by the header string
	spec := reg.checkCache(fullHeader)
	if spec != nil {
		return spec, nil
	}

	// allocate a spec
	spec = new(acceptSpec)
	spec.types = make(map[string]*acceptGroup, 0)

	// multiple mimetypes are comma-delimited
	for _, header := range strings.Split(fullHeader, ",") {
		// check validity and separate the options
		mt, params, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, err
		}

		// get the "q" float option
		q, err := qval(params)
		if err != nil {
			return nil, err
		}

		// handle the catch-all
		if mt == "*/*" {
			spec.starQ = q
			continue
		}

		// separate major/minor types (/minor optional)
		types := strings.SplitN(mt, "/", 2)
		major, minor := types[0], ""
		if len(types) == 2 {
			minor = types[1]
		}

		// traverse into the spec by major type,
		// creating the sub-struct if needed
		group, ok := spec.types[major]
		if !ok {
			group = &acceptGroup{0, make(map[string]float64)}
			spec.types[major] = group
		}
		if minor == "*" {
			// major-type-specific catchall (ie text/*)
			group.starQ = q
		} else {
			// a specific major/minor mimetype
			group.subTypes[minor] = q
		}
	}

	// populate the cache with the result
	reg.storeCache(fullHeader, spec)

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
		return 0, errors.New("qval must be <=1")
	}
	return q, nil
}

// Match on a registry performs the same operation as the Match function, just
// matches against the set of Protocols registered on the specific registry.
func (reg *Registry) Match(header string) (Protocol, error) {
	spec, err := reg.buildSpec(header)
	if err != nil {
		return nil, err
	}

	type protMatch struct {
		prot Protocol
		q    float64
	}
	var matches []protMatch

	reg.mut.RLock()
	defer reg.mut.RUnlock()

	for _, prot := range reg.protocols {
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
		if prot == reg.defaultProtocol {
			return prot, nil
		}
	}

	// just pick the first one
	return bestProts[0], nil
}
