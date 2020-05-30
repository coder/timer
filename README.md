# cmdperf

`cmdperf` measures the performance of a command by running it many times and producing runtime statistics.

It is inspired by `perf stat`, but works on MacOS.

## Install

```go
go get cdr.dev/cmdperf
```

## Basic Usage

```shell script
$ cmdperf -n 10 echo hello
hello
hello
hello
hello
hello
hello
hello
hello
hello
hello
hello
-- cmdperf
iterations  10
fastest     5ms
median      7ms
slowest     10ms
stddev      2
```