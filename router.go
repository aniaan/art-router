package artrouter

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
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

// Sort the list of nodes by label
func (ns nodes) Sort()              { sort.Sort(ns); ns.tailSort() }
func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Less(i, j int) bool { return ns[i].label < ns[j].label }

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

func (n *node) getEdge(ntyp nodeType, label, tail byte, rexpat string) *node {
	childs := n.children[ntyp]
	for i := 0; i < len(childs); i++ {
		if childs[i].label == label && childs[i].tail == tail {
			if ntyp == ntRegexp && childs[i].prefix != rexpat {
				continue
			}
			return childs[i]
		}
	}

	return nil
}

// addChild appends the new `child` node to the tree using the `pattern` as the trie key.
// For a URL router like chi's, we split the static, param, regexp and wildcard segments
// into different nodes. In addition, addChild will recursively call itself until every
// pattern segment is added to the url pattern tree as individual nodes, depending on type.
func (n *node) addChild(child *node, prefix string) *node {
	search := prefix

	// handler leaf node added to the tree is the child.
	// this may be overridden later down the flow
	hn := child

	// Parse next segment
	seg := patNextSegment(search)

	segType := seg.nodeType

	// Add child depending on next up segment
	switch segType {

	case ntStatic:
		// Search prefix is all static (that is, has no params in path)
		// noop

	case ntCatchAll:

	default:
		// Search prefix contains a param, regexp or wildcard

		ps := seg.ps
		pe := seg.pe

		if ps == 0 {
			// Route starts with a param
			child.typ = segType

			if segType == ntRegexp {
				rex, err := regexp.Compile(seg.rexpat)
				if err != nil {
					panic(fmt.Sprintf("invalid regexp pattern '%s' in route param", seg.rexpat))
				}
				child.prefix = seg.rexpat
				child.rex = rex
			}

			child.tail = seg.tail // for params, we set the tail

			if pe != len(search) {
				// add static edge for the remaining part, split the end.
				// its not possible to have adjacent param nodes, so its certainly
				// going to be a static node next.

				// prefix require update?
				child.prefix = search[:pe]

				search = search[pe:] // advance search position

				nn := &node{
					typ:    ntStatic, // after update
					label:  search[0],
					prefix: search,
				}
				hn = child.addChild(nn, search)
			}

		} else if ps > 0 {
			// Route has some param

			// starts with a static segment
			child.typ = ntStatic
			child.prefix = search[:ps]
			child.rex = nil

			// add the param edge node
			search = search[ps:]

			nn := &node{
				typ:   segType,
				label: search[0],
				tail:  seg.tail,
			}
			hn = child.addChild(nn, search)
		}
	}

	n.children[child.typ] = append(n.children[child.typ], child)
	n.children[child.typ].Sort()
	return hn
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

		var seg segment
		if label == '{' || label == '*' {
			seg = patNextSegment(search)
		}

		parent = n
		n = n.getEdge(seg.nodeType, label, seg.tail, seg.rexpat)

		if n == nil {
			child := &node{label: label, tail: seg.tail, prefix: search}
			hn := parent.addChild(child, search)
			hn.setEndpoint(method, pattern, handler)
			return hn, nil
		}

		if n.typ > ntStatic {
			search = search[seg.pe:]
			continue
		}

		//n.prefix compare
		commonPrefix := longestPrefix(search, n.prefix)


	}

	return nil, nil
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
