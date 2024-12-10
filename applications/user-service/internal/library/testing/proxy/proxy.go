package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Settings struct {
	Port      int      // kubectl proxy --port [Port] specification - defaults to 8080.
	Arguments []string // additional arguments to pass into the kubectl proxy command
}

type Variadic func(options *Settings)

// options represents a default constructor.
func options() *Settings {
	return &Settings{
		Port:      8080,
		Arguments: []string{},
	}
}

type Proxy struct {
	cmd     *exec.Cmd
	options *Settings
}

func (p *Proxy) Process() *os.Process {
	if p.cmd != nil {
		return p.cmd.Process
	}

	return nil
}

func (p *Proxy) Start(ctx context.Context) {
	defer time.Sleep(time.Millisecond * 250)

	arguments := []string{
		"proxy",
		"--port",
		strconv.Itoa(p.options.Port),
	}

	arguments = append(arguments, p.options.Arguments...)

	p.cmd = exec.CommandContext(ctx, "kubectl", arguments...)
	if e := p.cmd.Start(); e != nil {
		slog.ErrorContext(ctx, "Failed to Start Kubectl Proxy", slog.String("error", e.Error()))

		e = fmt.Errorf("failed to start kubectl proxy: %w", e)
		panic(e)
	}

	slog.InfoContext(ctx, "Started Kubectl Proxy", slog.Int("port", p.options.Port))

	return
}

func (p *Proxy) Stop(ctx context.Context) {
	if p.cmd != nil && p.cmd.Process != nil {
		if e := p.cmd.Process.Kill(); e != nil {
			slog.ErrorContext(ctx, "Failed to Stop Kubectl Proxy", slog.String("error", e.Error()))

			e = fmt.Errorf("failed to stop kubectl proxy: %w", e)
			panic(e)
		}
	}

	slog.DebugContext(ctx, "Successfully Stopped Kubectl Proxy")
}

func New(settings ...Variadic) *Proxy {
	o := options()
	for _, configuration := range settings {
		configuration(o)
	}

	return &Proxy{
		options: o,
	}
}
