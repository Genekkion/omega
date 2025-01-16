package runner

import (
	"context"
	"omega/internal/log"
	"omega/internal/structs"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Runner struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup *sync.WaitGroup

	runDelay   time.Duration
	silentFlag *atomic.Bool

	channel   chan struct{}
	commands  []string
	arguments [][]string
}

func NewRunner(config structs.Config, channel chan struct{}) *Runner {
	arguments := make([][]string, 0, len(config.Commands))
	for _, command := range config.Commands {
		arguments = append(arguments, strings.Split(command, " "))
	}
	silentFlag := &atomic.Bool{}
	silentFlag.Store(false)

	return &Runner{
		channel:    channel,
		commands:   config.Commands,
		arguments:  arguments,
		waitGroup:  &sync.WaitGroup{},
		silentFlag: silentFlag,
		runDelay:   time.Duration(config.Delay) * time.Millisecond,
	}
}

func (runner *Runner) runCommands() {
	runner.ctx, runner.cancel = context.WithCancel(runner.parentCtx)
	runner.waitGroup.Add(1)

	go func() {
		start := time.Now()
		for i, command := range runner.commands {
			err := runner.runCommand(runner.ctx, i)
			if err != nil {
				if !runner.silentFlag.Load() {
					log.Error("Error occurred running command",
						"command", command, "error", err)
				}
				return
			}
		}
		log.Info("End of execution", "duration", time.Now().Sub(start))

		runner.waitGroup.Done()
	}()
}

func (runner *Runner) runCommand(ctx context.Context, index int) error {
	arguments := runner.arguments[index]
	cmd := exec.CommandContext(ctx, arguments[0], arguments[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.WaitDelay = 0
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	ctxDone := false
	select {
	case <-ctx.Done():
		ctxDone = true
	default:
	}

	log.Debug("Process exited",
		"command", runner.commands[index],
		"code", cmd.ProcessState.ExitCode(),
		"contextDone", ctxDone,
	)
	return err
}

func (runner *Runner) Run(runnerCtx context.Context) error {
	log.Debug("Runner started")
	defer log.Debug("Runner shutting down")

	runner.parentCtx = runnerCtx
	runner.ctx, runner.cancel = context.WithCancel(context.Background())

	for {
		select {
		case <-runnerCtx.Done():
			runner.resetState()
			return nil

		case <-runner.channel:
			runner.resetState()

			// Small delay
			time.Sleep(runner.runDelay)

			runner.runCommands()
		}
	}
}

func (runner *Runner) resetState() {
	runner.silentFlag.Store(true)
	runner.cancel()
	runner.waitGroup.Wait()
	runner.silentFlag.Store(false)
}
