package artrouter

import (
	"net/http"
	"regexp"
	"strings"
)

// art-router implementation below is a based on the original work by
// go-chi in https://github.com/go-chi/chi/blob/master/tree.go
// (MIT licensed). It's been heavily modified for use as a HTTP router.

type methodType uint

const (
	mSTUB methodType = 1 << iota
	mCONNECT
	mDELETE
	mGET
	mHEAD
	mOPTIONS
	mPATCH
	mPOST
	mPUT
	mTRACE
)

var mALL = mCONNECT | mDELETE | mGET | mHEAD |
	mOPTIONS | mPATCH | mPOST | mPUT | mTRACE

var methodMap = map[string]methodType{
	http.MethodConnect: mCONNECT,
	http.MethodDelete:  mDELETE,
	http.MethodGet:     mGET,
	http.MethodHead:    mHEAD,
	http.MethodOptions: mOPTIONS,
	http.MethodPatch:   mPATCH,
	http.MethodPost:    mPOST,
	http.MethodPut:     mPUT,
	http.MethodTrace:   mTRACE,
}

type nodeType uint8

const (
	ntStatic   nodeType = iota // /home
	ntRegexp                   // /{id:[0-9]+}
	ntParam                    // /{user}
	ntCatchAll                 // /api/v1/*
)

// // Represents leaf node in radix tree
type endpoint struct {
	// endpoint handler
	handler http.Handler

	// pattern is the routing pattern for handler nodes
	pattern string

	// parameter keys recorded on handler nodes
	paramKeys []string
}

type endpoints map[methodType]*endpoint

type nodes []*node

// Represents node and edge in radix tree
type node struct {
	// regexp matcher for regexp nodes
	rex *regexp.Regexp

	// HTTP handler endpoints on the leaf node
	endpoints endpoints

	// prefix is the common prefix we ignore
	prefix string

	// child nodes should be stored in-order for iteration,
	// in groups of the node type.
	children [ntCatchAll + 1]nodes

	// first byte of the child prefix
	tail byte

	// node type: static, regexp, param, catchAll
	typ nodeType

	// first byte of the prefix
	label byte
}

func (n *node) setEndpoint(method methodType, pattern string, handler http.Handler) {
}

type segment struct {
	nodeType nodeType
	key      string
	rexpat   string
	tail     byte
	ps       int
	pe       int
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
	if ws < len(pattern)-1 {
		panic("wildcard '*' must be the last value in a route. trim trailing text or use a '{param}' instead")
	}

	// ws >0 && ps < 0
	if ps < 0 {
		return segment{
			nodeType: ntCatchAll,
			key:      "*",
			ps:       ws,
			pe:       len(pattern),
		}
	}

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
		panic("chi: route param closing delimiter '}' is missing")
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

type ArtRouter struct {
	root *node
	size int
}

func New() ArtRouter {
	return ArtRouter{
		root: &node{},
	}
}

func (ar *ArtRouter) Insert(method methodType, pattern string, handler http.Handler) (*node, error) {
	// valid param method != 0 && pattern != "" && handler != nil
	if method == 0 || pattern == "" || handler == nil {
		panic("param invalid")
	}

	var parent *node
	search := pattern
	n := ar.root

	for {
		if len(search) == 0 {
			n.setEndpoint(method, pattern, handler)
			return n, nil
		}

		label := search[0]

	}

	return nil, nil
}
