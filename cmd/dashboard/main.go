package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yklcs/agent-hub-mcp/internal/db"
	"github.com/yklcs/agent-hub-mcp/internal/ui"
)

// DashboardApp holds dependencies for testing
type DashboardApp struct {
	Logger   *log.Logger
	ExitFunc func(int)
	DBOpener func(string) (*db.DB, error)
}

// NewDashboardApp creates a new DashboardApp with defaults
func NewDashboardApp() *DashboardApp {
	return &DashboardApp{
		Logger:   log.New(os.Stderr, "", log.LstdFlags),
		ExitFunc: os.Exit,
		DBOpener: db.Open,
	}
}

func main() {
	app := NewDashboardApp()
	if err := app.Run(os.Args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		app.Logger.Printf("Error: %v\n", err)
		app.ExitFunc(1)
	}
}

// Run executes the application with given arguments and IO
func (a *DashboardApp) Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	dbPath := "agent-hub.db"
	if len(args) > 1 {
		dbPath = args[1]
	}

	database, err := a.DBOpener(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	model := ui.NewModel(database)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
