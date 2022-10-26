package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aniaan/art-router/router"
)

func main() {
	route_count := 1000 * 1
	match_times := 1000 * 100
	// match_times := 1

	path := "/12345"

	rules := []*router.Rule{}

	for i := 1; i <= route_count; i++ {
		host := GetMD5Hash(strconv.FormatInt(int64(i), 10)) + ".abc.com"
		rules = append(rules, &router.Rule{
			Host: host,
			Paths: []*router.Path{
				{
					Path:    path,
					Methods: []string{"GET"},
					Backend: host,
				},
			},
		})
	}

	r := router.New(rules, false)

	start := time.Now()
	host := GetMD5Hash(strconv.FormatInt(int64(500), 10)) + ".abc.com"

	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Host = host

	var context *router.Context
	for i := 1; i <= match_times; i++ {
		context = r.Search(req)
	}

	duration := time.Since(start).Seconds()
	fmt.Printf("matched res: %t\n", context.Route != nil)
	fmt.Printf("route count: %d\n", route_count)
	fmt.Printf("match times: %d\n", match_times)
	fmt.Printf("time used  : %f  %s\n", duration, " sec")
	fmt.Printf("QPS        : %f", float64(match_times)/duration)
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// local routes = {}
// for i = 1, route_count do
//     routes[i] = {paths = {path}, priority = i, hosts = {ngx.md5(i)}, metadata = i}
// end

// local rx = radix.new(routes)

// ngx.update_time()
// local start_time = ngx.now()

// local res
// local opts = {
//     host = ngx.md5(500),
// }
// for _ = 1, match_times do
//     res = rx:match(path, opts)
// end

// ngx.update_time()
// local used_time = ngx.now() - start_time
// ngx.say("matched res: ", res)
// ngx.say("route count: ", route_count)
// ngx.say("match times: ", match_times)
// ngx.say("time used  : ", used_time, " sec")
// ngx.say("QPS        : ", math.floor(match_times / used_time))
