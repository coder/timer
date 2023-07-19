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
	rootCmd.PersistentFlags().IntVarP(&iterations, "iterations", "n", 0, "number of iterations")
	rootCmd.PersistentFlags().IntVarP(&parallelism, "parallelism", "p", 1, "number of concurrent workers")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "don't show command output")
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
	// This newline prefix helps when the output is mangled.
	if !quiet {
		fmt.Fprintf(wr, "\n")
	}
	fmt.Fprintf(wr, "--- timer config\n")
	fmt.Fprintf(wr, "command\t%s\n", strings.Join(command, " "))
	fmt.Fprintf(wr, "iterations\t%v\n", iterations)
	fmt.Fprintf(wr, "parallelism\t%v\n", parallelism)
	fmt.Fprintf(wr, "--- percentiles\n")
	fmt.Fprintf(wr, "0\t(fastest)\t%.3fs\n", sample.Quantile(0))
	fmt.Fprintf(wr, "25\t(1st quantile)\t%.3fs\n", sample.Quantile(0.25))
	fmt.Fprintf(wr, "50\t(median)\t%.3fs\n", sample.Quantile(0.5))
	fmt.Fprintf(wr, "75\t(3rd quantile)\t%.3fs\n", sample.Quantile(0.75))
	fmt.Fprintf(wr, "100th\t(slowest)\t%.3fs\n", sample.Quantile(1))
	fmt.Fprintf(wr, "--- summary\n")
	fmt.Fprintf(wr, "total\t%.3fs\n", totalTook.Seconds())
	fmt.Fprintf(wr, "mean\t%.3fs\n", sample.Mean())
	fmt.Fprintf(wr, "stddev\t%.3fs\n", sample.StdDev())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
