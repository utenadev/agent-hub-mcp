package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

// App holds dependencies for testing
type App struct {
	Logger   *log.Logger
	ExitFunc func(int)
}

// NewApp creates a new App with defaults
func NewApp() *App {
	return &App{
		Logger:   log.New(os.Stderr, "", log.LstdFlags),
		ExitFunc: os.Exit,
	}
}

func main() {
	app := NewApp()
	if err := app.Run(os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		app.Logger.Printf("Error: %v\n", err)
		app.ExitFunc(1)
	}
}

// Run executes the application with given arguments and IO
func (a *App) Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if len(args) < 2 {
		a.runHelp(stdout)
		return nil
	}

	command := args[1]

	switch command {
	case "serve":
		return a.runServe(args[2:], stdout, stderr)
	case "orchestrator":
		return a.runOrchestrator(args[2:], stdout, stderr)
	case "doctor":
		return a.runDoctor(args[2:], stdout, stderr)
	case "setup":
		return a.runSetup(args[2:], stdout, stderr)
	case "help", "--help", "-h":
		a.runHelp(stdout)
		return nil
	default:
		return fmt.Errorf("Unknown command: %s\nRun 'agent-hub help' for usage", command)
	}
}


