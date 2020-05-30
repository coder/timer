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
	"github.com/spf13/pflag"
	"go.coder.com/cli"
	"go.coder.com/flog"
)

type rootCmd struct {
	iterations  int
	parallelism int
	quiet       bool
}

func (r *rootCmd) RegisterFlags(fl *pflag.FlagSet) {
	fl.SetInterspersed(false)
	fl.IntVarP(&r.iterations, "iterations", "n", 0, "number of iterations")
	fl.IntVarP(&r.parallelism, "parallelism", "p", 1, "number of concurrent workers")
	fl.BoolVarP(&r.quiet, "quiet", "q", false, "don't show command output")
}

func (r *rootCmd) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:  "cmdperf",
		Usage: "[flags] [test command ...]",
		Desc:  "Measure the performance of command execution",
	}
}

func (r *rootCmd) Run(fl *pflag.FlagSet) {
	if r.iterations == 0 {
		flog.Fatal("iterations (-n) must be provided")
	}

	command := fl.Args()
	if len(command) == 0 {
		flog.Fatal("command must provided")
	}

	var (
		commands   = make(chan *exec.Cmd)
		tooksMu    sync.Mutex
		tooks      = make([]float64, 0, r.iterations)
		totalStart = time.Now()
	)
	var commandWaitGroup sync.WaitGroup
	for i := 0; i < r.parallelism; i++ {
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
					flog.Fatal("command execution failed: %+v", err)
				}
			}
		}()
	}

	go func() {
		defer close(commands)
		for i := 0; i < r.iterations; i++ {
			cmd := exec.Command(command[0], command[1:]...)
			if !r.quiet {
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
	if !r.quiet {
		fmt.Fprintf(wr, "\n")
	}
	fmt.Fprintf(wr, "--- config\n")
	fmt.Fprintf(wr, "command\t%s\n", strings.Join(command, " "))
	fmt.Fprintf(wr, "iterations\t%v\n", r.iterations)
	fmt.Fprintf(wr, "parallelism\t%v\n", r.parallelism)
	fmt.Fprintf(wr, "--- percentiles\n")
	fmt.Fprintf(wr, "0\t(fastest)\t%.3f\n", sample.Quantile(0))
	fmt.Fprintf(wr, "25\t(1st quantile)\t%.3f\n", sample.Quantile(0.25))
	fmt.Fprintf(wr, "50\t(median)\t%.3f\n", sample.Quantile(0.5))
	fmt.Fprintf(wr, "75\t(3rd quantile)\t%.3f\n", sample.Quantile(0.75))
	fmt.Fprintf(wr, "100th\t(slowest)\t%.3f\n", sample.Quantile(1))
	fmt.Fprintf(wr, "--- summary\n")
	fmt.Fprintf(wr, "total\t%.3f\n", totalTook.Seconds())
	fmt.Fprintf(wr, "mean\t%.3f\n", sample.Mean())
	fmt.Fprintf(wr, "stddev\t%.3f\n", sample.StdDev())

}

func main() {
	cli.Run(&rootCmd{}, os.Args[1:], "")
}
