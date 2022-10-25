package artrouter

import (
	"fmt"
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
					Methods: []string{"mGET"},
					Backend: hIndex,
				},

				{
					Path:    "/favicon.ico",
					Methods: []string{"mGET"},
					Backend: hFavicon,
				},

				{
					Path:    "/pages/*",
					Methods: []string{"mGET"},
					Backend: hStub,
				},

				{
					Path:    "/article",
					Methods: []string{"mGET"},
					Backend: hArticleList,
				},

				{
					Path:    "/article/",
					Methods: []string{"mGET"},
					Backend: hArticleList,
				},

				{
					Path:    "/article/near",
					Methods: []string{"mGET"},
					Backend: hArticleNear,
				},

				{
					Path:    "/article/{id}",
					Methods: []string{"mGET"},
					Backend: hStub,
				},

				{
					Path:    "/article/{id}",
					Methods: []string{"mGET"},
					Backend: hArticleShow,
				},

				{
					Path:    "/article/{id}",
					Methods: []string{"mGET"},
					Backend: hArticleShow,
				},

				{
					Path:    "/article/@{user}",
					Methods: []string{"mGET"},
					Backend: hArticleByUser,
				},

				{
					Path:    "/article/{sup}/{opts}",
					Methods: []string{"mGET"},
					Backend: hArticleShowOpts,
				},

				{
					Path:    "/article/{id}/{opts}",
					Methods: []string{"mGET"},
					Backend: hArticleShowOpts,
				},

				{
					Path:    "/article/{iffd}/edit",
					Methods: []string{"mGET"},
					Backend: hStub,
				},

				{
					Path:    "/article/{id}//related",
					Methods: []string{"mGET"},
					Backend: hArticleShowRelated,
				},

				{
					Path:    "/article/slug/{month}/-/{day}/{year}",
					Methods: []string{"mGET"},
					Backend: hArticleSlug,
				},

				{
					Path:    "/admin/user",
					Methods: []string{"mGET"},
					Backend: hUserList,
				},

				{
					Path:    "/admin/user/",
					Methods: []string{"mGET"},
					Backend: hStub,
				},

				{
					Path:    "/admin/user/",
					Methods: []string{"mGET"},
					Backend: hUserList,
				},

				{
					Path:    "/admin/user//{id}",
					Methods: []string{"mGET"},
					Backend: hUserShow,
				},

				{
					Path:    "/admin/user/{id}",
					Methods: []string{"mGET"},
					Backend: hUserShow,
				},

				{
					Path:    "/admin/apps/{id}",
					Methods: []string{"mGET"},
					Backend: hAdminAppShow,
				},

				{
					Path:    "/admin/apps/{id}/*",
					Methods: []string{"mGET"},
					Backend: hAdminAppShowCatchall,
				},

				{
					Path:    "/admin/*",
					Methods: []string{"mGET"},
					Backend: hStub,
				},

				{
					Path:    "/admin/*",
					Methods: []string{"mGET"},
					Backend: hAdminCatchall,
				},

				{
					Path:    "/users/{userID}/profile",
					Methods: []string{"mGET"},
					Backend: hUserProfile,
				},

				{
					Path:    "/users/super/*",
					Methods: []string{"mGET"},
					Backend: hUserSuper,
				},

				{
					Path:    "/users/*",
					Methods: []string{"mGET"},
					Backend: hUserAll,
				},

				{
					Path:    "/hubs/{hubID}/view",
					Methods: []string{"mGET"},
					Backend: hHubView1,
				},

				{
					Path:    "/hubs/{hubID}/view/*",
					Methods: []string{"mGET"},
					Backend: hHubView2,
				},

				{
					Path:    "/hubs/{hubID}/*",
					Methods: []string{"mGET"},
					Backend: "sr",
				},
				{
					Path:    "/hubs/{hubID}/users",
					Methods: []string{"mGET"},
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
		{r: "/", h: hIndex, k: []string{}, v: []string{}},
		{r: "/favicon.ico", h: hFavicon, k: []string{}, v: []string{}},

		{r: "/pages", h: "", k: []string{}, v: []string{}},
		{r: "/pages/", h: hStub, k: []string{"*"}, v: []string{""}},
		{r: "/pages/yes", h: hStub, k: []string{"*"}, v: []string{"yes"}},

		{r: "/article", h: hArticleList, k: []string{}, v: []string{}},
		{r: "/article/", h: hArticleList, k: []string{}, v: []string{}},
		{r: "/article/near", h: hArticleNear, k: []string{}, v: []string{}},
		{r: "/article/neard", h: hStub, k: []string{"id"}, v: []string{"neard"}},
		{r: "/article/123", h: hStub, k: []string{"id"}, v: []string{"123"}},
		{r: "/article/123/456", h: hArticleShowOpts, k: []string{"id", "opts"}, v: []string{"123", "456"}},
		{r: "/article/@peter", h: hArticleByUser, k: []string{"user"}, v: []string{"peter"}},
		{r: "/article/22//related", h: hArticleShowRelated, k: []string{"id"}, v: []string{"22"}},
		{r: "/article/111/edit", h: hStub, k: []string{"iffd"}, v: []string{"111"}},
		{r: "/article/slug/sept/-/4/2015", h: hArticleSlug, k: []string{"month", "day", "year"}, v: []string{"sept", "4", "2015"}},
		{r: "/article/:id", h: hStub, k: []string{"id"}, v: []string{":id"}},

		{r: "/admin/user", h: hUserList, k: []string{}, v: []string{}},
		{r: "/admin/user/", h: hStub, k: []string{}, v: []string{}},
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

	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")
	// debugPrintTree(0, 0, tr, 0)
	// log.Println("~~~~~~~~~")
	// log.Println("~~~~~~~~~")

	router := New(rules)
	assert := assert.New(t)

	for i, tt := range tests {
		// rctx := NewRouteContext()
		req, _ := http.NewRequest(http.MethodGet, tt.r, nil)
		fmt.Println(i)
		context := router.Search(req)

		var backend string

		if context.route != nil {
			backend = context.route.backend
		}

		assert.Equal(tt.h, backend)

		// paramKeys := rctx.routeParams.Keys
		// paramValues := rctx.routeParams.Values

		// if fmt.Sprintf("%v", tt.h) != fmt.Sprintf("%v", handler) {
		// t.Errorf("input [%d]: find '%s' expecting handler:%v , got:%v", i, tt.r, tt.h, handler)
		// }
		// if !stringSliceEqual(tt.k, paramKeys) {
		// 	t.Errorf("input [%d]: find '%s' expecting paramKeys:(%d)%v , got:(%d)%v", i, tt.r, len(tt.k), tt.k, len(paramKeys), paramKeys)
		// }
		// if !stringSliceEqual(tt.v, paramValues) {
		// 	t.Errorf("input [%d]: find '%s' expecting paramValues:(%d)%v , got:(%d)%v", i, tt.r, len(tt.v), tt.v, len(paramValues), paramValues)
		// }
	}
}
