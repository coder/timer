# timer

`timer` is like `time` but repeats your command and provides basic statistics on execution time.

It's inspired by `perf stat`, but works on macOS.

## Install

```go
go get github.com/coder/timer@master
```

## Basic Usage

```shell script
$ timer -n 10 curl google.com
--- config
command        curl google.com
iterations     10
parallelism    1
unit           1ms
--- histogram
208.449-228.108  70%  ████████████████▏  7
228.108-247.766  10%  ██▎                1
247.766-267.425  0%   ▏                  
267.425-287.084  0%   ▏                  
287.084-306.742  10%  ██▎                1
306.742-326.401  0%   ▏                  
326.401-346.060  0%   ▏                  
346.060-365.719  10%  ██▎                1
--- summary
total     2.463s
mean      242.043
median    222.534
stddev    50.767
```

## Parallelism

You can use the `-p` flag to configure the number of parallel threads.

```shell script
$ timer -n 4 -p 2 sleep 1
--- config
command        sleep 1
iterations     4
parallelism    2
unit           1s
--- histogram
1.012-1.014  50%  ████████████████▏  2
1.014-1.016  25%  ████████▏          1
1.016-1.018  0%   ▏                  
1.018-1.020  25%  ████████▏          1
--- summary
total     2s
mean      1.015
median    1.014
stddev    0.004
```

## Similar Projects

- [bench (Haskell)](https://hackage.haskell.org/package/bench)
- [Apache Bench](https://httpd.apache.org/docs/2.4/programs/ab.html)
