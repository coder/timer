package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"go.coder.com/cli"
	"go.coder.com/flog"
)

type rootCmd struct {
	iterations  int
	parallelism int
}

func (r *rootCmd) RegisterFlags(fl *pflag.FlagSet) {
	fl.SetInterspersed(false)
	fl.IntVarP(&r.iterations, "iterations", "n", 0, "number of iterations")
	fl.IntVarP(&r.parallelism, "parallelism", "p", 1, "number of concurrent workers")
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
		commands = make(chan *exec.Cmd)
		tooksMu  sync.Mutex
		tooks    = make([]time.Duration, 0, r.iterations)
	)
	var commandWaitGroup sync.WaitGroup
	for i := 0; i < r.parallelism; i++ {
		commandWaitGroup.Add(1)
		go func() {
			defer commandWaitGroup.Done()

			for cmd := range commands {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				start := time.Now()
				err := cmd.Run()
				took := time.Since(start)
				tooksMu.Lock()
				tooks = append(tooks, took)
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
			commands <- cmd
		}
	}()

	commandWaitGroup.Wait()
	fmt.Printf("tooks: %+v\n", tooks)

}

func main() {
	//var flagArgs []string
	//for i, arg := range os.Args {
	//	// The first non-flag arguments marks the beginning of the user-provided command.
	//	if _, err := strconv.Atoi(arg); err != nil && !strings.HasPrefix(arg, "-")  {
	//		continue
	//	}
	//	flagArgs = os.Args[1:i+1]
	//}
	cli.Run(&rootCmd{}, os.Args[1:], "")
}
