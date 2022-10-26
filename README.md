# art-router

## Benchmark

Machine environment: MacBook Pro (13-inch, M1, 2020).

### match-host

```shell
$ go build benchmark/matchhost/matchhost.go && ./matchhost
matched res: true
route count: 1000
match times: 100000
time used  : 0.236790   sec
QPS        : 422314.758800
```

### match-static
