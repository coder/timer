# timer

`timer` is like `time` but repeats your command and provides basic statistics on execution time.

It's inspired by `perf stat`, but works on MacOS.

## Install

```go
go get cdr.dev/timer
```

## Basic Usage

```shell script
$ timer -n curl google.com
--- timer config
command        curl google.com
iterations     10
parallelism    1
--- percentiles
0        (fastest)         0.037s
25       (1st quantile)    0.041s
50       (median)          0.044s
75       (3rd quantile)    0.049s
100th    (slowest)         0.059s
--- summary
total     0.455s
mean      0.046s
stddev    0.006s
```
_[Apache Bench](https://httpd.apache.org/docs/2.4/programs/ab.html) is typically better for website ;)_

## Parallelism

You can use the `-p` flag to configure the number of parallel threads.

```shell script
$ timer -n 4 -p 2 sleep 1s
--- timer config
command        sleep 1s
iterations     4
parallelism    2
--- percentiles
0        (fastest)         1.005s
25       (1st quantile)    1.005s
50       (median)          1.006s
75       (3rd quantile)    1.006s
100th    (slowest)         1.006s
--- summary
total     2.013s
mean      1.006s
stddev    0.001s
```

## Example: Benchmark google.com