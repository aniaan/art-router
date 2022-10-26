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
	route_count := 1000 * 100
	match_times := 1000 * 1000
	// match_times := 1

	paths := []*router.Path{}

	for i := 1; i <= route_count; i++ {
		path := "/" + GetMD5Hash(strconv.FormatInt(int64(i), 10)) + "/*"
		paths = append(paths,
			&router.Path{
				Path:    path,
				Methods: []string{"GET"},
				Backend: path,
			})
	}

	rules := []*router.Rule{
		{
			Paths: paths,
		},
	}

	// r := router.New(rules, false)
	r := router.New(rules, false)

	// cpuProfile := "cpu.pprof"
	// 采样cpu运行状态

	// f, _ := os.Create(cpuProfile)
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	start := time.Now()

	path := "/" + GetMD5Hash(strconv.FormatInt(int64(500), 10)) + "/a"

	req, _ := http.NewRequest(http.MethodGet, path, nil)

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
