package artrouter

import (
	"fmt"
	"strings"
)

type segment struct {
	key      string
	rexpat   string
	ps       int
	pe       int
	nodeType nodeType
	tail     byte
}

func patNextSegment(pattern string) segment {
	ps := strings.Index(pattern, "{")
	ws := strings.Index(pattern, "*")

	if ps < 0 && ws < 0 {
		// we return the entire thing
		return segment{
			nodeType: ntStatic,
			pe:       len(pattern),
		}
	}

	// Sanity check
	if ps >= 0 && ws >= 0 && ws < ps {
		panic("wildcard '*' must be the last pattern in a route, otherwise use a '{param}'")
	}

	// Wildcard pattern as finale

	if ps >= 0 {

		var tail byte = '/' // Default endpoint tail to / byte
		// Param/Regexp pattern is next
		nt := ntParam

		// Read to closing } taking into account opens and closes in curl count (cc)
		cc := 0
		pe := ps
		for i, c := range pattern[ps:] {
			if c == '{' {
				cc++
			} else if c == '}' {
				cc--
				if cc == 0 {
					pe = ps + i
					break
				}
			}
		}
		if pe == ps {
			panic("route param closing delimiter '}' is missing")
		}

		key := pattern[ps+1 : pe]
		pe++ // set end to next position

		if pe < len(pattern) {
			tail = pattern[pe]
		}

		var rexpat string
		if idx := strings.Index(key, ":"); idx >= 0 {
			nt = ntRegexp
			rexpat = key[idx+1:]
			key = key[:idx]
		}

		if len(rexpat) > 0 {
			if rexpat[0] != '^' {
				rexpat = "^" + rexpat
			}
			if rexpat[len(rexpat)-1] != '$' {
				rexpat += "$"
			}
		}

		return segment{
			nodeType: nt,
			key:      key,
			rexpat:   rexpat,
			tail:     tail,
			ps:       ps,
			pe:       pe,
		}
	}

	if ws < len(pattern)-1 {
		panic("wildcard '*' must be the last value in a route. trim trailing text or use a '{param}' instead")
	}

	return segment{
		nodeType: ntCatchAll,
		key:      "*",
		ps:       ws,
		pe:       len(pattern),
	}
}

func patParamKeys(pattern string) []string {
	pat := pattern
	paramKeys := []string{}
	for {
		seg := patNextSegment(pat)
		if seg.nodeType == ntStatic {
			return paramKeys
		}
		for i := 0; i < len(paramKeys); i++ {
			if paramKeys[i] == seg.key {
				panic(fmt.Sprintf("routing pattern '%s' contains duplicate param key, '%s'", pattern, seg.key))
			}
		}
		paramKeys = append(paramKeys, seg.key)
		pat = pat[seg.pe:]
	}
}

// longestPrefix finds the length of the shared prefix
// of two strings
func longestPrefix(k1, k2 string) int {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}
	var i int
	for i = 0; i < max; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	return i
}

func normalizePath(path string) string {
	if path[len(path)-1] == '/' {
		return path[:len(path)-1]
	}
	return path
}

// StrInSlice returns whether the string is in the slice.
func StrInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}
