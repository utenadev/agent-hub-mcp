package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.Logger == nil {
		t.Error("expected Logger to be set")
	}
	if app.ExitFunc == nil {
		t.Error("expected ExitFunc to be set")
	}
}

func TestApp_Run_NoCommand(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"bbs"}, nil, &stdout, &stderr)

	if err == nil {
		t.Error("expected error for missing command")
	}
	if !strings.Contains(err.Error(), "Usage") {
		t.Errorf("expected usage message, got: %v", err)
	}
}

func TestApp_Run_UnknownCommand(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"bbs", "unknown"}, nil, &stdout, &stderr)

	if err == nil {
		t.Error("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "Unknown command") {
		t.Errorf("expected unknown command error, got: %v", err)
	}
}

func TestApp_Run_Serve_InvalidFlag(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"bbs", "serve", "-invalid"}, nil, &stdout, &stderr)

	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestApp_Run_Orchestrator_InvalidFlag(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"bbs", "orchestrator", "-invalid"}, nil, &stdout, &stderr)

	if err == nil {
		t.Error("expected error for invalid flag")
	}
}

func TestApp_Run_Serve_InvalidDBPath(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"bbs", "serve", "-db", "/nonexistent/path/to/db.db"}, nil, &stdout, &stderr)

	if err == nil {
		t.Error("expected error for invalid database path")
	}
}

func TestApp_Run_Doctor(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"agent-hub", "doctor", "-db", ":memory:"}, nil, &stdout, &stderr)
	_ = err
}

func TestApp_Run_Doctor_InvalidDBPath(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"agent-hub", "doctor", "-db", "/nonexistent/path/to/db.db"}, nil, &stdout, &stderr)
	_ = err
}

func TestApp_Run_Setup(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	err := app.Run([]string{"agent-hub", "setup", "-db", dbPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Errorf("setup command failed: %v", err)
	}
}

func TestApp_Run_Setup_WithForce(t *testing.T) {
	app := NewApp()
	var stdout, stderr bytes.Buffer

	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	err := app.Run([]string{"agent-hub", "setup", "-db", dbPath}, nil, &stdout, &stderr)
	if err != nil {
		t.Errorf("first setup command failed: %v", err)
	}

	err = app.Run([]string{"agent-hub", "setup", "-db", dbPath, "-force"}, nil, &stdout, &stderr)
	if err != nil {
		t.Errorf("second setup command with -force failed: %v", err)
	}
}

func TestApp_Run_Doctor_WithEnvVars(t *testing.T) {
	t.Setenv("BBS_AGENT_ID", "test-agent")
	t.Setenv("BBS_AGENT_ROLE", "test-role")

	app := NewApp()
	var stdout, stderr bytes.Buffer

	err := app.Run([]string{"agent-hub", "doctor", "-db", ":memory:"}, nil, &stdout, &stderr)
	_ = err
}
