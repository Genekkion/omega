package program

import (
	"context"
	"omega/internal/config"
	"omega/internal/log"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type Runner struct {
	// Should be the main context
	ctx context.Context

	// Delay between reruns
	delay time.Duration

	// To signal a rerun
	ch chan struct{}

	// Misc
	cmds []string
	args [][]string
}

func NewRunner(config config.Config, ch chan struct{}) *Runner {
	// Parse config commands into suitable cli arguments
	args := make([][]string, 0, len(config.Commands))
	for _, command := range config.Commands {
		args = append(args, strings.Split(command, " "))
	}

	silent := &atomic.Bool{}
	silent.Store(false)

	return &Runner{
		ch:    ch,
		cmds:  config.Commands,
		args:  args,
		delay: time.Duration(config.Delay) * time.Millisecond,
	}
}

// Runs all the commands from the config
func (r *Runner) runCmds(ctx context.Context) {
	start := time.Now()
	for i, command := range r.cmds {
		quit, err := r.runCmd(ctx, i)
		if err != nil {
			if !quit {
				// For when errors occur not due to the cmd ran but from Omega itself
				log.Error("Error occurred running command",
					"command", command, "error", err)
			}
			return
		}
	}
	log.Info("End of execution", "duration", time.Now().Sub(start))
}

// A.k.a. run a single program with its arguments
func (r *Runner) runCmd(ctx context.Context, index int) (quit bool, err error) {
	args := r.args[index]

	// Execute the program, using the specified context
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.WaitDelay = 0
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	select {
	case <-ctx.Done():
		quit = true
	default:
		quit = false
	}

	log.Debug("Process exited",
		"command", r.cmds[index],
		"code", cmd.ProcessState.ExitCode(),
	)
	return quit, err
}

// Start the runner
func (r *Runner) Start(ctx context.Context) error {
	log.Debug("Runner started")
	defer log.Debug("Runner shutting down")

	first := true

	for {
		// Additional block to exit faster
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		select {
		case <-ctx.Done():
			return nil

		case <-r.ch:
			// Small delay
			if first {
				first = false
			} else {
				time.Sleep(r.delay)
			}

			r.runCmds(ctx)
		}
	}
}
