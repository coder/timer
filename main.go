package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/aclements/go-moremath/stats"
	"github.com/aybabtme/uniplot/histogram"
	"github.com/coder/flog"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "timer [flags] [test command ...]",
	Short: "Measure the performance of command execution",
	Run:   executeCommand,
}

var (
	iterations  int
	parallelism int
	quiet       bool
)

func init() {
	rootCmd.Flags().IntVarP(&iterations, "iterations", "n", 0, "number of iterations")
	rootCmd.Flags().IntVarP(&parallelism, "parallelism", "p", 1, "number of concurrent workers")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't show command output")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func scaleOf(x float64) time.Duration {
	switch {
	case x < 1e-6:
		return time.Nanosecond
	case x < 1e-3:
		return time.Microsecond
	case x < 1:
		return time.Millisecond
	case x < 60:
		return time.Second
	case x < 60*60:
		return time.Minute
	case x < 24*60*60:
		return time.Hour
	default:
		return time.Millisecond
	}
}

func executeCommand(cmd *cobra.Command, args []string) {
	if iterations == 0 {
		flog.Fatalf("iterations (-n) must be provided")
	}

	command := args
	if len(command) == 0 {
		flog.Fatalf("command must provided")
	}

	var (
		commands   = make(chan *exec.Cmd)
		tooksMu    sync.Mutex
		tooks      = make([]float64, 0, iterations)
		totalStart = time.Now()
	)
	var commandWaitGroup sync.WaitGroup
	for i := 0; i < parallelism; i++ {
		commandWaitGroup.Add(1)
		go func() {
			defer commandWaitGroup.Done()

			for cmd := range commands {
				start := time.Now()
				err := cmd.Run()
				took := time.Since(start)
				tooksMu.Lock()
				tooks = append(tooks, took.Seconds())
				tooksMu.Unlock()

				if err != nil {
					flog.Fatalf("command execution failed: %+v", err)
				}
			}
		}()
	}

	go func() {
		defer close(commands)
		for i := 0; i < iterations; i++ {
			cmd := exec.Command(command[0], command[1:]...)
			if !quiet {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}
			commands <- cmd
		}
	}()

	commandWaitGroup.Wait()
	totalTook := time.Since(totalStart)

	sample := &stats.Sample{Xs: tooks}
	sample = sample.Sort()

	wr := tabwriter.NewWriter(os.Stdout, 6, 4, 4, ' ', 0)
	defer wr.Flush()
	// This newline prefix helps separate timer output from command output.
	if !quiet {
		fmt.Fprintf(wr, "\n")
	}

	mean := sample.Mean()
	// Find the order of magnitude difference between the mean and a second.
	scale := scaleOf(mean)

	// Then, scale the mean and the sample to that order of magnitude, truncating
	// noise.

	fmt.Fprintf(wr, "--- config\n")
	fmt.Fprintf(wr, "command\t%s\n", strings.Join(command, " "))
	fmt.Fprintf(wr, "iterations\t%v\n", iterations)
	fmt.Fprintf(wr, "parallelism\t%v\n", parallelism)
	fmt.Fprintf(wr, "unit\t%v\n", scale)

	for i := range tooks {
		tooks[i] = tooks[i] / scale.Seconds()
	}

	fmt.Fprintf(wr, "--- histogram\n")
	hist := histogram.Hist(minInt(iterations, 8), tooks)
	histogram.Fprintf(wr, hist, histogram.Linear(16), func(v float64) string {
		return fmt.Sprintf("%.3f", v)
	})
	fmt.Fprintf(wr, "--- summary\n")
	fmt.Fprintf(wr, "total\t%v\n", totalTook.Truncate(scale))
	fmt.Fprintf(wr, "mean\t%.3f\n", mean/scale.Seconds())
	fmt.Fprintf(wr, "median\t%.3f\n", sample.Quantile(0.5))
	fmt.Fprintf(wr, "stddev\t%.3f\n", sample.StdDev())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
