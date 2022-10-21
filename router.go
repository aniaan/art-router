package artrouter

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/cases"
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

func (s endpoints) Value(method methodType) *endpoint {
	mh, ok := s[method]
	if !ok {
		mh = &endpoint{}
		s[method] = mh
	}
	return mh
}

type nodes []*node

// Sort the list of nodes by label
func (ns nodes) Sort()              { sort.Sort(ns); ns.tailSort() }
func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Less(i, j int) bool { return ns[i].label < ns[j].label }

// tailSort pushes nodes with '/' as the tail to the end of the list for param nodes.
// The list order determines the traversal order.
func (ns nodes) tailSort() {
	for i := len(ns) - 1; i >= 0; i-- {
		// param node label is {,
		if ns[i].typ > ntStatic && ns[i].tail == '/' {
			ns.Swap(i, len(ns)-1)
			return
		}
	}
}

func (ns nodes) findEdge(label byte) *node {
	// static nodes find
	num := len(ns)
	idx := 0
	i, j := 0, num-1
	for i <= j {
		idx = i + (j-i)/2
		if label > ns[idx].label {
			i = idx + 1
		} else if label < ns[idx].label {
			j = idx - 1
		} else {
			i = num // breaks cond
		}
	}
	if ns[idx].label != label {
		return nil
	}
	return ns[idx]
}

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

	// case ntCatchAll:

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

	if child.typ == ntParam && len(n.children[child.typ]) >= 1 {
		panic("param error")
	}

	n.children[child.typ] = append(n.children[child.typ], child)
	n.children[child.typ].Sort()
	return hn
}

func (n *node) replaceChild(label, tail byte, child *node) {
	for i := 0; i < len(n.children[child.typ]); i++ {
		if n.children[child.typ][i].label == label && n.children[child.typ][i].tail == tail {
			n.children[child.typ][i] = child
			n.children[child.typ][i].label = label
			n.children[child.typ][i].tail = tail
			return
		}
	}
	panic("replacing missing child")
}

func (n *node) setEndpoint(method methodType, pattern string, handler http.Handler) {
	// Set the handler for the method type on the node
	if n.endpoints == nil {
		n.endpoints = make(endpoints)
	}

	paramKeys := patParamKeys(pattern)
	if method&mSTUB == mSTUB {
		n.endpoints.Value(mSTUB).handler = handler
	}
	if method&mALL == mALL {
		h := n.endpoints.Value(mALL)
		h.handler = handler
		h.pattern = pattern
		h.paramKeys = paramKeys
		for _, m := range methodMap {
			h := n.endpoints.Value(m)
			h.handler = handler
			h.pattern = pattern
			h.paramKeys = paramKeys
		}
	} else {
		h := n.endpoints.Value(method)
		h.handler = handler
		h.pattern = pattern
		h.paramKeys = paramKeys
	}
}

func (n *node) find(method methodType, path string) *node {
	nn := n
	search := path

	for t, nds := range nn.children {

		if len(nds) == 0 {
			continue
		}

		ntype := nodeType(t)

		var xn *node
		xsearch := search

		label := search[0]

		switch ntype {
		case ntStatic:
			xn = nds.findEdge(label)
			if xn == nil || !strings.HasPrefix(xsearch, xn.prefix) {
				continue
			}
			xsearch = xsearch[len(xn.prefix):]

			if len(xsearch) == 0 {
				if !xn.isLeaf() {
					continue
				}

				h := xn.endpoints[method]
				if h == nil {
					continue
				}
				// rctx.routeParams.Keys = append(rctx.routeParams.Keys, h.paramKeys...)
				return xn
			}

			fin := xn.find(method, xsearch)
			if fin != nil {
				return fin
			}

			continue

		case ntParam:
			// short-circuit and return no matching route for empty param values
			if xsearch == "" {
				continue
			}

			xn = nds[0]
			p := strings.IndexByte(xsearch, xn.tail)

			if p < 0 {
				if xn.tail == '/' {
					// xsearch is param value
					p = len(search)
				} else {
					continue
				}
			} else if strings.IndexByte(xsearch[:p], '/') != -1 {
				// avoid a match across path segments
				continue
			}

			// TODO param value record
			xsearch = xsearch[p:]

			if len(xsearch) == 0 {
				// xsearch is end and match the param check
				if !xn.isLeaf() {
					continue
				}
				h := xn.endpoints[method]
				if h != nil && h.handler != nil {
					// rctx.routeParams.Keys = append(rctx.routeParams.Keys, h.paramKeys...)
					return xn
				}

				// flag that the routing context found a route, but not a corresponding
				// supported method
				// rctx.methodNotAllowed = true

				// method not allow, continue check
				continue
			}

			fin := xn.find(method, xsearch)

			if fin != nil {
				return fin
			}

			continue

		case ntRegexp:
			if xsearch == "" {
				continue
			}

			for idx := 0; idx < len(nds); idx++ {
				xn = nds[idx]

				p := strings.IndexByte(xsearch, xn.tail)

				if p < 0 {
					if xn.tail == '/' {
						// xsearch is param value
						p = len(search)
					} else {
						continue
					}
				} else if p == 0 {
					continue
				}

				if !xn.rex.MatchString(xsearch[:p]) {
					continue
				}

				xsearch = xsearch[p:]

				if len(xsearch) == 0 {
					if !xn.isLeaf() {
						continue
					}

					h := xn.endpoints[method]
					if h != nil && h.handler != nil {
						// rctx.routeParams.Keys = append(rctx.routeParams.Keys, h.paramKeys...)
						return xn
					}
					continue
				}

				fin := xn.find(method, xsearch)
				if fin != nil {
					return fin
				}

				xsearch = search

			}

		default:
			xn = nn

			if !xn.isLeaf() {
				continue
			}

			h := xn.endpoints[method]
			if h == nil {
				continue
			}
			// rctx.routeParams.Keys = append(rctx.routeParams.Keys, h.paramKeys...)
			return xn

		}

	}

	return nil
}

func (n *node) isLeaf() bool {
	return n.endpoints != nil
}

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

	search := normalizePath(pattern)

	var parent *node
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

		// n.prefix compare
		commonPrefix := longestPrefix(search, n.prefix)

		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:]
			continue
		}

		// split the node
		child := &node{
			typ:    ntStatic,
			prefix: search[:commonPrefix],
		}

		parent.replaceChild(label, seg.tail, child)

		n.label = n.prefix[commonPrefix]
		n.prefix = n.prefix[commonPrefix:]
		child.addChild(n, n.prefix)

		search = search[commonPrefix:]
		if len(search) == 0 {
			child.setEndpoint(method, pattern, handler)
			return child, nil
		}

		subChild := &node{
			typ:    ntStatic,
			label:  search[0],
			prefix: search,
		}

		hn := child.addChild(subChild, search)
		hn.setEndpoint(method, pattern, handler)
		return hn, nil

	}
}

func (ar *ArtRouter) Find() {
}
