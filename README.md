# art-router

## Benchmark

Machine environment: MacBook Pro (15-inch, 2019). 2.6 GHz Intel Core i7

### match-host

```shell
$ go build benchmark/matchhost/matchhost.go && ./matchhost
matched res: true
route count: 1000
match times: 100000
time used  : 0.259201   sec
QPS        : 385801.564169
```

### match-static

>> 做了懒加载, benchmark已经上去了，稍后补充
>> 相比lua-resty-radixtree的性能，主要耗时在net.splitHost和url.Query()上，lua-resty-radixtree的benchmark参数host和query是提前预置好的，所以qps会高,如果把net.splitHost和url.Query()也做成固定值，qps可以做到1300w

```shell
$ go build benchmark/matchstatic/matchstatic.go && ./matchstatic

matched res: true
route count: 100000
match times: 10000000
time used  : 1.612419   sec
QPS        : 6201861.876831
```

### match-prefix

```shell
$ go build benchmark/matchprefix/matchprefix.go && ./matchprefix
matched res: true
route count: 100000
match times: 1000000
time used  : 0.381809   sec
QPS        : 2619109.792973
```

### match-param

```shell
$ go build benchmark/matchparam/matchparam.go && ./matchparam
matched res: true
route count: 100000
match times: 10000000
time used  : 2.879797   sec
QPS        : 3472466.506624
```

### match-regexp

```shell
$ go build benchmark/matchreg/matchreg.go && ./matchreg
matched res: true
route count: 1000
match times: 1000000
time used  : 0.427930   sec
QPS        : 2336832.646749
```
