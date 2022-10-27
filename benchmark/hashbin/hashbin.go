package main

import (
	"fmt"
	"time"
)

func main() {
	count := 1000 * 1000 * 100

	// for i := 1; i <= 256; i++ {
	// 	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	// 	var c byte = byte(i)
	// 	h(count, c)
	// 	b(count, c)
	// 	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	// }

	st1("/tianji/exchange/popup/task/abc", "/tianji/exchange/popup/task/abc", count)
	st2("/tianji/exchange/popup/task/abc", "/tianji/exchange/popup/task/abc", count)
}

func h(count int, c byte) {
	data := make(map[byte]byte)

	for i := 1; i <= 256; i++ {
		data[byte(i)] = byte(i)
	}

	// var c byte = byte(127)
	var r byte

	start := time.Now()

	for a := 0; a < count; a++ {
		if item, ok := data[c]; ok {
			r = item
		}
	}

	duration := time.Since(start).Seconds()
	fmt.Printf("hash search matched res: %d time used %f sec\n", r, duration)
}

func b(count int, c byte) {
	data := []byte{}

	for i := 1; i <= 256; i++ {
		data = append(data, byte(i))
	}

	// var c byte = byte(127)
	var r byte

	start := time.Now()

	for a := 0; a < count; a++ {

		num := 256
		idx := 0
		i, j := 0, num-1
		for i <= j {
			idx = i + (j-i)/2
			if c > data[idx] {
				i = idx + 1
			} else if c < data[idx] {
				j = idx - 1
			} else {
				i = num // breaks cond
			}
		}
		if data[idx] == c {
			r = data[idx]
		}

	}

	duration := time.Since(start).Seconds()
	fmt.Printf("bin search matched res: %d time used %f sec\n", r, duration)
}

func st1(s1, s2 string, count int) {
	start := time.Now()
	var res bool
	for a := 0; a < count; a++ {
		res = s1 == s2
	}

	duration := time.Since(start).Seconds()
	fmt.Printf("hash search matched res: %t time used %f sec\n", res, duration)
}

func st2(s1, s2 string, count int) {
	start := time.Now()
	var res bool
	for a := 0; a < count; a++ {
		res = se(s1, s2)
	}

	duration := time.Since(start).Seconds()
	fmt.Printf("hash search matched res: %t time used %f sec\n", res, duration)
}

func se(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}
