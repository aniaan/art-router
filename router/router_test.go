package router

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	hStub := "hStub"
	hIndex := "hIndex"
	hFavicon := "hFavicon"
	hArticleList := "hArticleList"
	hArticleNear := "hArticleNear"
	hArticleShow := "hArticleShow"
	hArticleShowRelated := "hArticleShowRelated"
	hArticleShowOpts := "hArticleShowOpts"
	hArticleSlug := "hArticleSlug"
	hArticleByUser := "hArticleByUser"
	hUserList := "hUserList"
	hUserShow := "hUserShow"
	hAdminCatchall := "hAdminCatchall"
	hAdminAppShow := "hAdminAppShow"
	hAdminAppShowCatchall := "hAdminAppShowCatchall"
	hUserProfile := "hUserProfile"
	hUserSuper := "hUserSuper"
	hUserAll := "hUserAll"
	hHubView1 := "hHubView1"
	hHubView2 := "hHubView2"
	hHubView3 := "hHubView3"

	rules := []*Rule{
		{
			Paths: []*Path{
				{
					Path:    "/",
					Methods: []string{"GET"},
					Backend: hIndex,
				},

				{
					Path:    "/favicon.ico",
					Methods: []string{"GET"},
					Backend: hFavicon,
				},

				{
					Path:    "/pages/*",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/article",
					Methods: []string{"GET"},
					Backend: hArticleList,
				},

				{
					Path:    "/article/",
					Methods: []string{"GET"},
					Backend: hArticleList,
				},

				{
					Path:    "/article/near",
					Methods: []string{"GET"},
					Backend: hArticleNear,
				},

				{
					Path:    "/article/{id}",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/article/{id}",
					Methods: []string{"GET"},
					Backend: hArticleShow,
				},

				{
					Path:    "/article/{id}",
					Methods: []string{"GET"},
					Backend: hArticleShow,
				},

				{
					Path:    "/article/@{user}",
					Methods: []string{"GET"},
					Backend: hArticleByUser,
				},

				{
					Path:    "/article/{sup}/{opts}",
					Methods: []string{"GET"},
					Backend: hArticleShowOpts,
				},

				{
					Path:    "/article/{id}/{opts}",
					Methods: []string{"GET"},
					Backend: hArticleShowOpts,
				},

				{
					Path:    "/article/{iffd}/edit",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/article/{id}//related",
					Methods: []string{"GET"},
					Backend: hArticleShowRelated,
				},

				{
					Path:    "/article/slug/{month}/-/{day}/{year}",
					Methods: []string{"GET"},
					Backend: hArticleSlug,
				},

				{
					Path:    "/admin/user",
					Methods: []string{"GET"},
					Backend: hUserList,
				},

				{
					Path:    "/admin/user/",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/admin/user/",
					Methods: []string{"GET"},
					Backend: hUserList,
				},

				{
					Path:    "/admin/user//{id}",
					Methods: []string{"GET"},
					Backend: hUserShow,
				},

				{
					Path:    "/admin/user/{id}",
					Methods: []string{"GET"},
					Backend: hUserShow,
				},

				{
					Path:    "/admin/apps/{id}",
					Methods: []string{"GET"},
					Backend: hAdminAppShow,
				},

				{
					Path:    "/admin/apps/{id}/*",
					Methods: []string{"GET"},
					Backend: hAdminAppShowCatchall,
				},

				{
					Path:    "/admin/*",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/admin/*",
					Methods: []string{"GET"},
					Backend: hAdminCatchall,
				},

				{
					Path:    "/users/{userID}/profile",
					Methods: []string{"GET"},
					Backend: hUserProfile,
				},

				{
					Path:    "/users/super/*",
					Methods: []string{"GET"},
					Backend: hUserSuper,
				},

				{
					Path:    "/users/*",
					Methods: []string{"GET"},
					Backend: hUserAll,
				},

				{
					Path:    "/hubs/{hubID}/view",
					Methods: []string{"GET"},
					Backend: hHubView1,
				},

				{
					Path:    "/hubs/{hubID}/view/*",
					Methods: []string{"GET"},
					Backend: hHubView2,
				},

				{
					Path:    "/hubs/{hubID}/*",
					Methods: []string{"GET"},
					Backend: "sr",
				},
				{
					Path:    "/hubs/{hubID}/users",
					Methods: []string{"GET"},
					Backend: hHubView3,
				},
			},
		},
	}

	tests := []struct {
		r string   // input request path
		h string   // output matched handler
		k []string // output param keys
		v []string // output param values
	}{
		{r: "/", h: hIndex, k: nil, v: nil},
		{r: "/favicon.ico", h: hFavicon, k: nil, v: nil},

		{r: "/pages", h: "", k: nil, v: nil},
		{r: "/pages/", h: hStub, k: []string{"*"}, v: []string{""}},
		{r: "/pages/yes", h: hStub, k: []string{"*"}, v: []string{"yes"}},

		{r: "/article", h: hArticleList, k: nil, v: nil},
		{r: "/article/", h: hArticleList, k: nil, v: nil},
		{r: "/article/near", h: hArticleNear, k: nil, v: nil},
		{r: "/article/neard", h: hStub, k: []string{"id"}, v: []string{"neard"}},
		{r: "/article/123", h: hStub, k: []string{"id"}, v: []string{"123"}},
		{r: "/article/123/456", h: hArticleShowOpts, k: []string{"sup", "opts"}, v: []string{"123", "456"}},
		{r: "/article/@peter", h: hArticleByUser, k: []string{"user"}, v: []string{"peter"}},
		{r: "/article/22//related", h: hArticleShowRelated, k: []string{"id"}, v: []string{"22"}},
		{r: "/article/111/edit", h: hStub, k: []string{"iffd"}, v: []string{"111"}},
		{r: "/article/slug/sept/-/4/2015", h: hArticleSlug, k: []string{"month", "day", "year"}, v: []string{"sept", "4", "2015"}},
		{r: "/article/:id", h: hStub, k: []string{"id"}, v: []string{":id"}},

		{r: "/admin/user", h: hUserList, k: nil, v: nil},
		{r: "/admin/user/", h: hStub, k: nil, v: nil},
		{r: "/admin/user/1", h: hUserShow, k: []string{"id"}, v: []string{"1"}},
		{r: "/admin/user//1", h: hUserShow, k: []string{"id"}, v: []string{"1"}},
		{r: "/admin/hi", h: hStub, k: []string{"*"}, v: []string{"hi"}},
		{r: "/admin/lots/of/:fun", h: hStub, k: []string{"*"}, v: []string{"lots/of/:fun"}},
		{r: "/admin/apps/333", h: hAdminAppShow, k: []string{"id"}, v: []string{"333"}},
		{r: "/admin/apps/333/woot", h: hAdminAppShowCatchall, k: []string{"id", "*"}, v: []string{"333", "woot"}},

		{r: "/hubs/123/view", h: hHubView1, k: []string{"hubID"}, v: []string{"123"}},
		{r: "/hubs/123/view/index.html", h: hHubView2, k: []string{"hubID", "*"}, v: []string{"123", "index.html"}},
		{r: "/hubs/123/users", h: hHubView3, k: []string{"hubID"}, v: []string{"123"}},

		{r: "/users/123/profile", h: hUserProfile, k: []string{"userID"}, v: []string{"123"}},
		{r: "/users/super/123/okay/yes", h: hUserSuper, k: []string{"*"}, v: []string{"123/okay/yes"}},
		{r: "/users/123/okay/yes", h: hUserAll, k: []string{"*"}, v: []string{"123/okay/yes"}},
	}

	router := New(rules, false)
	assert := assert.New(t)

	for _, tt := range tests {
		req, _ := http.NewRequest(http.MethodGet, tt.r, nil)
		// fmt.Println(i)
		context := router.Search(req)

		var backend string

		if context.Route != nil {
			backend = context.Route.backend
		}

		assert.Equal(tt.h, backend)

		paramKeys := context.routeParams.Keys
		paramValues := context.routeParams.Values

		assert.Equal(tt.k, paramKeys)
		assert.Equal(tt.v, paramValues)
	}
}

func TestTreeMoar(t *testing.T) {
	hStub := "hStub"
	hStub1 := "hStub1"
	hStub2 := "hStub2"
	hStub3 := "hStub3"
	hStub4 := "hStub4"
	hStub5 := "hStub5"
	hStub6 := "hStub6"
	hStub7 := "hStub7"
	hStub8 := "hStub8"
	hStub9 := "hStub9"
	hStub10 := "hStub10"
	hStub11 := "hStub11"
	hStub12 := "hStub12"
	hStub13 := "hStub13"
	hStub14 := "hStub14"
	hStub15 := "hStub15"
	hStub16 := "hStub16"

	// TODO: panic if we see {id}{x} because we're missing a delimiter, its not possible.
	// also {:id}* is not possible.

	rules := []*Rule{
		{
			Paths: []*Path{
				{
					Path:    "/articlefun",
					Methods: []string{"GET"},
					Backend: hStub5,
				},

				{
					Path:    "/articles/{id}",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/articles/{slug}",
					Methods: []string{"DELETE"},
					Backend: hStub8,
				},

				{
					Path:    "/articles/search",
					Methods: []string{"GET"},
					Backend: hStub1,
				},

				{
					Path:    "/articles/{id}:delete",
					Methods: []string{"GET"},
					Backend: hStub8,
				},

				{
					Path:    "/articles/{iidd}!sup",
					Methods: []string{"GET"},
					Backend: hStub4,
				},

				{
					Path:    "/articles/{id}:{op}",
					Methods: []string{"GET"},
					Backend: hStub3,
				},

				{
					Path:    "/articles/{id}:{op}",
					Methods: []string{"GET"},
					Backend: hStub2,
				},

				{
					Path:    "/articles/{slug:^[a-z]+}/posts",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/articles/{id}/posts/{pid}",
					Methods: []string{"GET"},
					Backend: hStub6,
				},

				{
					Path:    "/articles/{id}/posts/{month}/{day}/{year}/{slug}",
					Methods: []string{"GET"},
					Backend: hStub7,
				},

				{
					Path:    "/articles/{id}.json",
					Methods: []string{"GET"},
					Backend: hStub10,
				},

				{
					Path:    "/articles/{id}/data.json",
					Methods: []string{"GET"},
					Backend: hStub11,
				},

				{
					Path:    "/articles/files/{file}.{ext}",
					Methods: []string{"GET"},
					Backend: hStub12,
				},

				{
					Path:    "/articles/me",
					Methods: []string{"PUT"},
					Backend: hStub13,
				},

				{
					Path:    "/pages/*",
					Methods: []string{"GET"},
					Backend: hStub,
				},

				{
					Path:    "/pages/*",
					Methods: []string{"GET"},
					Backend: hStub9,
				},

				{
					Path:    "/users/{id}",
					Methods: []string{"GET"},
					Backend: hStub14,
				},

				{
					Path:    "/users/{id}/settings/{key}",
					Methods: []string{"GET"},
					Backend: hStub15,
				},

				{
					Path:    "/users/{id}/settings/*",
					Methods: []string{"GET"},
					Backend: hStub16,
				},
			},
		},
	}

	tests := []struct {
		h string
		r string
		m string
		k []string
		v []string
	}{
		{m: "GET", r: "/articles/search", h: hStub1, k: nil, v: nil},
		{m: "GET", r: "/articlefun", h: hStub5, k: nil, v: nil},
		{m: "GET", r: "/articles/123", h: hStub, k: []string{"id"}, v: []string{"123"}},
		{m: "DELETE", r: "/articles/123mm", h: hStub8, k: []string{"slug"}, v: []string{"123mm"}},
		{m: "GET", r: "/articles/789:delete", h: hStub8, k: []string{"id"}, v: []string{"789"}},
		{m: "GET", r: "/articles/789!sup", h: hStub4, k: []string{"iidd"}, v: []string{"789"}},
		{m: "GET", r: "/articles/123:sync", h: hStub3, k: []string{"id", "op"}, v: []string{"123", "sync"}},
		{m: "GET", r: "/articles/456/posts/1", h: hStub6, k: []string{"id", "pid"}, v: []string{"456", "1"}},
		{m: "GET", r: "/articles/456/posts/09/04/1984/juice", h: hStub7, k: []string{"id", "month", "day", "year", "slug"}, v: []string{"456", "09", "04", "1984", "juice"}},
		{m: "GET", r: "/articles/456.json", h: hStub10, k: []string{"id"}, v: []string{"456"}},
		{m: "GET", r: "/articles/456/data.json", h: hStub11, k: []string{"id"}, v: []string{"456"}},
		{m: "GET", r: "/articles/files/file.zip", h: hStub12, k: []string{"file", "ext"}, v: []string{"file", "zip"}},
		{m: "GET", r: "/articles/files/photos.tar.gz", h: hStub12, k: []string{"file", "ext"}, v: []string{"photos", "tar.gz"}},
		{m: "GET", r: "/articles/files/photos.tar.gz", h: hStub12, k: []string{"file", "ext"}, v: []string{"photos", "tar.gz"}},
		{m: "PUT", r: "/articles/me", h: hStub13, k: nil, v: nil},
		{m: "GET", r: "/articles/me", h: hStub, k: []string{"id"}, v: []string{"me"}},
		{m: "GET", r: "/pages", h: "", k: nil, v: nil},
		{m: "GET", r: "/pages/", h: hStub, k: []string{"*"}, v: []string{""}},
		{m: "GET", r: "/pages/yes", h: hStub, k: []string{"*"}, v: []string{"yes"}},
		{m: "GET", r: "/users/1", h: hStub14, k: []string{"id"}, v: []string{"1"}},
		{m: "GET", r: "/users/", h: "", k: nil, v: nil},
		{m: "GET", r: "/users/2/settings/password", h: hStub15, k: []string{"id", "key"}, v: []string{"2", "password"}},
		{m: "GET", r: "/users/2/settings/", h: hStub16, k: []string{"id", "*"}, v: []string{"2", ""}},
	}

	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")
	// debugPrintTree(0, 0, tr, 0)
	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")

	router := New(rules, false)
	assert := assert.New(t)

	for _, tt := range tests {
		req, _ := http.NewRequest(tt.m, tt.r, nil)
		// fmt.Println(i)
		context := router.Search(req)

		var backend string

		if context.Route != nil {
			backend = context.Route.backend
		}

		assert.Equal(tt.h, backend)

		paramKeys := context.routeParams.Keys
		paramValues := context.routeParams.Values

		assert.Equal(tt.k, paramKeys)
		assert.Equal(tt.v, paramValues)
	}
}

func TestTreeRegexp(t *testing.T) {
	hStub1 := "hStub1"
	hStub2 := "hStub2"
	hStub3 := "hStub3"
	hStub4 := "hStub4"
	hStub5 := "hStub5"
	hStub6 := "hStub6"
	hStub7 := "hStub7"

	rules := []*Rule{
		{
			Paths: []*Path{
				{
					Path:    "/articles/{rid:^[0-9]{5,6}}",
					Methods: []string{"GET"},
					Backend: hStub7,
				},

				{
					Path:    "/articles/{zid:^0[0-9]+}",
					Methods: []string{"GET"},
					Backend: hStub3,
				},

				{
					Path:    "/articles/{name:^@[a-z]+}/posts",
					Methods: []string{"GET"},
					Backend: hStub4,
				},

				{
					Path:    "/articles/{op:^[0-9]+}/run",
					Methods: []string{"GET"},
					Backend: hStub5,
				},

				{
					Path:    "/articles/{id:^[0-9]+}",
					Methods: []string{"GET"},
					Backend: hStub1,
				},

				{
					Path:    "/articles/{id:^[1-9]+}-{aux}",
					Methods: []string{"GET"},
					Backend: hStub6,
				},

				{
					Path:    "/articles/{slug}",
					Methods: []string{"GET"},
					Backend: hStub2,
				},
			},
		},
	}

	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")
	// debugPrintTree(0, 0, tr, 0)
	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")

	tests := []struct {
		r string   // input request path
		h string   // output matched handler
		k []string // output param keys
		v []string // output param values
	}{
		{r: "/articles", h: "", k: nil, v: nil},
		{r: "/articles/12345", h: hStub7, k: []string{"rid"}, v: []string{"12345"}},
		{r: "/articles/123", h: hStub1, k: []string{"id"}, v: []string{"123"}},
		{r: "/articles/how-to-build-a-router", h: hStub2, k: []string{"slug"}, v: []string{"how-to-build-a-router"}},
		{r: "/articles/0456", h: hStub3, k: []string{"zid"}, v: []string{"0456"}},
		{r: "/articles/@pk/posts", h: hStub4, k: []string{"name"}, v: []string{"@pk"}},
		{r: "/articles/1/run", h: hStub5, k: []string{"op"}, v: []string{"1"}},
		{r: "/articles/1122", h: hStub1, k: []string{"id"}, v: []string{"1122"}},
		{r: "/articles/1122-yes", h: hStub6, k: []string{"id", "aux"}, v: []string{"1122", "yes"}},
	}

	router := New(rules, false)
	assert := assert.New(t)

	for _, tt := range tests {
		req, _ := http.NewRequest(http.MethodGet, tt.r, nil)
		// fmt.Println(i)
		context := router.Search(req)

		var backend string

		if context.Route != nil {
			backend = context.Route.backend
		}

		assert.Equal(tt.h, backend)

		paramKeys := context.routeParams.Keys
		paramValues := context.routeParams.Values

		assert.Equal(tt.k, paramKeys)
		assert.Equal(tt.v, paramValues)
	}
}

func TestTreeRegexpRecursive(t *testing.T) {
	hStub1 := "hStub1"
	hStub2 := "hStub2"

	rules := []*Rule{
		{
			Paths: []*Path{
				{
					Path:    "/one/{firstId:[a-z0-9-]+}/{secondId:[a-z0-9-]+}/first",
					Methods: []string{"GET"},
					Backend: hStub1,
				},

				{
					Path:    "/one/{firstId:[a-z0-9-_]+}/{secondId:[a-z0-9-_]+}/second",
					Methods: []string{"GET"},
					Backend: hStub2,
				},
			},
		},
	}

	tests := []struct {
		r string   // input request path
		h string   // output matched handler
		k []string // output param keys
		v []string // output param values
	}{
		{r: "/one/hello/world/first", h: hStub1, k: []string{"firstId", "secondId"}, v: []string{"hello", "world"}},
		{r: "/one/hi_there/ok/second", h: hStub2, k: []string{"firstId", "secondId"}, v: []string{"hi_there", "ok"}},
		{r: "/one///first", h: "", k: nil, v: nil},
		{r: "/one/hi/123/second", h: hStub2, k: []string{"firstId", "secondId"}, v: []string{"hi", "123"}},
	}

	router := New(rules, false)
	assert := assert.New(t)

	for _, tt := range tests {
		req, _ := http.NewRequest(http.MethodGet, tt.r, nil)
		// fmt.Println(i)
		context := router.Search(req)

		var backend string

		if context.Route != nil {
			backend = context.Route.backend
		}

		assert.Equal(tt.h, backend)

		paramKeys := context.routeParams.Keys
		paramValues := context.routeParams.Values

		assert.Equal(tt.k, paramKeys)
		assert.Equal(tt.v, paramValues)
	}
}

func TestTreeRegexMatchWholeParam(t *testing.T) {
	hStub1 := "hStub1"
	hStub2 := "hStub1"
	hStub3 := "hStub1"

	rules := []*Rule{
		{
			Paths: []*Path{
				{
					Path:    "/{id:[0-9]+}",
					Methods: []string{"GET"},
					Backend: hStub1,
				},

				{
					Path:    "/{x:.+}/foo",
					Methods: []string{"GET"},
					Backend: hStub2,
				},
				{
					Path:    "/{param:[0-9]*}/test",
					Methods: []string{"GET"},
					Backend: hStub3,
				},
			},
		},
	}

	tests := []struct {
		expectedHandler string
		url             string
	}{
		{url: "/13", expectedHandler: hStub1},
		{url: "/a13", expectedHandler: ""},
		{url: "/13.jpg", expectedHandler: ""},
		{url: "/a13.jpg", expectedHandler: ""},
		{url: "/a/foo", expectedHandler: hStub2},
		{url: "//foo", expectedHandler: ""},
		{url: "//test", expectedHandler: ""},
	}

	router := New(rules, false)
	assert := assert.New(t)

	for _, tt := range tests {
		req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
		// fmt.Println(i)
		context := router.Search(req)

		var backend string

		if context.Route != nil {
			backend = context.Route.backend
		}

		assert.Equal(tt.expectedHandler, backend)

	}
}

func BenchmarkTreeGet(b *testing.B) {
	h1 := "h1"
	h2 := "h2"

	rules := []*Rule{
		{
			Paths: []*Path{
				{
					Path:    "/",
					Methods: []string{"GET"},
					Backend: h1,
				},

				{
					Path:    "/ping",
					Methods: []string{"GET"},
					Backend: h2,
				},

				{
					Path:    "/pingall",
					Methods: []string{"GET"},
					Backend: h2,
				},

				{
					Path:    "/ping/{id}",
					Methods: []string{"GET"},
					Backend: h2,
				},

				{
					Path:    "/ping/{id}/woop",
					Methods: []string{"GET"},
					Backend: h2,
				},

				{
					Path:    "/ping/{id}/{opt}",
					Methods: []string{"GET"},
					Backend: h2,
				},

				{
					Path:    "/pinggggg",
					Methods: []string{"GET"},
					Backend: h2,
				},

				{
					Path:    "/hello",
					Methods: []string{"GET"},
					Backend: h1,
				},
			},
		},
	}

	router := New(rules, false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/ping/123/456", nil)
		router.Search(req)
	}
}
