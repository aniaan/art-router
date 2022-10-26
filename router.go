package artrouter

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// art-router implementation below is a based on the original work by
// go-chi in https://github.com/go-chi/chi/blob/master/tree.go
// (MIT licensed). It's been heavily modified for use as a HTTP router.

type (
	methodType uint
	nodeType   uint8

	// Represents leaf node in radix tree
	route struct {
		pattern        string
		backend        string
		headers        []*Header
		queries        []*Query
		paramKeys      []string
		method         methodType
		matchAllHeader bool
	}

	// Represents node and edge in radix tree
	node struct {
		// regexp matcher for regexp nodes
		rex *regexp.Regexp

		// HTTP handler endpoints on the leaf node
		routes []*route

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

	nodes []*node

	muxRule struct {
		hostRE     *regexp.Regexp
		root       *node
		host       string
		hostRegexp string
	}

	ArtRouter struct {
		rules []*muxRule
	}

	routeParams struct {
		Keys, Values []string
	}

	// Context is the default routing context
	context struct {
		headers     http.Header
		queries     url.Values
		route       *route
		path        string
		routeParams routeParams
		method      methodType
	}
)

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

const (
	ntStatic   nodeType = iota // /home
	ntRegexp                   // /{id:[0-9]+}
	ntParam                    // /{user}
	ntCatchAll                 // /api/v1/*
)

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
				typ:    segType,
				label:  search[0],
				tail:   seg.tail,
				prefix: search,
			}
			hn = child.addChild(nn, search)
		}
	}

	// if child.typ == ntParam && len(n.children[child.typ]) >= 1 {
	// 	panic("param error")
	// }

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

func (n *node) setRoute(path *Path) {
	if n.routes == nil {
		n.routes = make([]*route, 0)
	}

	paramKeys := patParamKeys(path.Path)

	method := mALL
	if len(path.Methods) != 0 {
		method = 0
		for _, m := range path.Methods {
			method |= methodMap[m]
		}
	}

	for _, p := range path.Headers {
		p.initHeaderRoute()
	}

	for _, q := range path.Queries {
		q.initQueryRoute()
	}

	r := &route{
		pattern:        path.Path,
		backend:        path.Backend,
		headers:        path.Headers,
		queries:        path.Queries,
		matchAllHeader: path.MatchAllHeader,
		paramKeys:      paramKeys,
		method:         method,
	}

	n.routes = append(n.routes, r)
}

func (root *node) insert(path *Path) (*node, error) {
	if path == nil {
		panic("param invalid")
	}

	// search := normalizePath(path.Path)
	search := path.Path

	var parent *node
	n := root

	for {

		if len(search) == 0 {
			n.setRoute(path)
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
			hn.setRoute(path)
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
			child.setRoute(path)
			return child, nil
		}

		subChild := &node{
			typ:    ntStatic,
			label:  search[0],
			prefix: search,
		}

		hn := child.addChild(subChild, search)
		hn.setRoute(path)
		return hn, nil

	}
}

func (n *node) match(context *context) *route {
	for _, r := range n.routes {
		if r.match(context) {
			return r
		}
	}

	return nil
}

func (n *node) find(path string, context *context) *route {
	nn := n
	search := path

	for t, nds := range nn.children {

		if len(nds) == 0 {
			continue
		}

		ntype := nodeType(t)

		var xn *node
		xsearch := search

		var label byte

		if search != "" {
			label = search[0]
		}

		switch ntype {
		case ntStatic:

			if xsearch == "" {
				continue
			}

			xn = nds.findEdge(label)
			if xn == nil || !strings.HasPrefix(xsearch, xn.prefix) {
				continue
			}
			xsearch = xsearch[len(xn.prefix):]

			if len(xsearch) == 0 {
				if xn.isLeaf() {
					r := xn.match(context)
					if r != nil {
						// context.routeParams.Keys = append(context.routeParams.Keys, r.paramKeys...)
						return r
					}
				}
			}
			fin := xn.find(xsearch, context)
			if fin != nil {
				// context.routeParams.Keys = append(context.routeParams.Keys, fin.paramKeys...)
				return fin
			}

		case ntParam, ntRegexp:
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
				} else if ntype == ntRegexp && p == 0 {
					continue
				}

				if ntype == ntRegexp {
					if !xn.rex.MatchString(xsearch[:p]) {
						continue
					}
				} else if strings.IndexByte(xsearch[:p], '/') != -1 {
					continue
				}

				prevlen := len(context.routeParams.Values)
				context.routeParams.Values = append(context.routeParams.Values, xsearch[:p])
				xsearch = xsearch[p:]

				if len(xsearch) == 0 {
					if xn.isLeaf() {
						r := xn.match(context)
						if r != nil {
							return r
						}
					}
				}
				fin := xn.find(xsearch, context)
				if fin != nil {
					return fin
				}
				context.routeParams.Values = context.routeParams.Values[:prevlen]
				xsearch = search
			}

		default:
			xn = nds[0]
			r := xn.match(context)
			if r != nil {
				context.routeParams.Values = append(context.routeParams.Values, xsearch)
				return r
			}
		}

	}

	return nil
}

func (n *node) isLeaf() bool {
	return n.routes != nil
}

func (r *route) match(context *context) bool {
	// method match
	if context.method&r.method == 0 {
		return false
	}

	if len(r.headers) > 0 && !r.matchHeaders(context.headers) {
		return false
	}

	if len(r.queries) > 0 && !r.matchQueries(context.queries) {
		return false
	}

	return true
}

func (r *route) matchHeaders(headers http.Header) bool {
	if len(r.headers) == 0 {
		return true
	}

	if r.matchAllHeader {
		for _, h := range r.headers {
			v := headers.Get(h.Key)
			if len(h.Values) > 0 && !StrInSlice(v, h.Values) {
				return false
			}

			if h.Regexp != "" && !h.headerRE.MatchString(v) {
				return false
			}
		}
	} else {
		for _, h := range r.headers {
			v := headers.Get(h.Key)
			if StrInSlice(v, h.Values) {
				return true
			}

			if h.Regexp != "" && h.headerRE.MatchString(v) {
				return true
			}
		}
	}

	return r.matchAllHeader
}

func (r *route) matchQueries(query url.Values) bool {
	if len(r.queries) == 0 {
		return true
	}

	for _, q := range r.queries {
		v := query.Get(q.Key)
		if len(q.Values) > 0 && !StrInSlice(v, q.Values) {
			return false
		}

		if q.Regexp != "" && !q.re.MatchString(v) {
			return false
		}
	}

	return true
}

func newMuxRule(rule *Rule) *muxRule {
	var hostRE *regexp.Regexp

	if rule.HostRegexp != "" {
		var err error
		hostRE, err = regexp.Compile(rule.HostRegexp)
		// defensive programming
		if err != nil {
			panic(fmt.Sprintf("BUG: compile %s failed: %v", rule.HostRegexp, err))
		}
	}

	root := &node{}
	for _, path := range rule.Paths {
		_, err := root.insert(path)
		if err != nil {
			panic(err)
		}
	}

	return &muxRule{
		host:       rule.Host,
		hostRegexp: rule.HostRegexp,
		hostRE:     hostRE,
		root:       root,
	}
}

func (mr *muxRule) match(host string) bool {
	if mr.host == "" && mr.hostRE == nil {
		return true
	}

	if mr.host != "" && mr.host == host {
		return true
	}
	if mr.hostRE != nil && mr.hostRE.MatchString(host) {
		return true
	}

	return false
}

func New(rules []*Rule) ArtRouter {
	router := ArtRouter{
		rules: make([]*muxRule, 0),
	}

	for _, rule := range rules {
		mr := newMuxRule(rule)
		router.rules = append(router.rules, mr)

	}

	return router
}

func (ar *ArtRouter) Search(req *http.Request) *context {
	host := req.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	method := methodMap[req.Method]
	// path := normalizePath(req.URL.Path)
	path := req.URL.Path

	context := &context{
		method:  method,
		path:    path,
		headers: req.Header,
		queries: req.URL.Query(),
	}

	for _, rule := range ar.rules {
		if !rule.match(host) {
			continue
		}
		route := rule.root.find(path, context)

		if route != nil {
			context.route = route
			context.routeParams.Keys = append(context.routeParams.Keys, route.paramKeys...)
			return context
		}
	}

	return context
}
