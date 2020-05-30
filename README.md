# timer

`timer` is like `time` but repeats your command and provides basic statistics on execution time.

It's inspired by `perf stat`, but works on MacOS.

## Install

```go
go get cdr.dev/timer
```

## Basic Usage

```shell script
$ timer -n 4 sleep 1s
--- config
command        sleep 1s
iterations     4
parallelism    1
--- percentiles
0        (fastest)         1.004
25       (1st quantile)    1.004
50       (median)          1.006
75       (3rd quantile)    1.008
100th    (slowest)         1.008
--- summary
mean      1.006
stddev    0.002
```