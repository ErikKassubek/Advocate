// Copyright (c) 2025 Erik Kassubek
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package command

import (
	"advocate/utils/flags"
	"bytes"
	"os"
	"testing"
	"time"
)

func TestRunCommand(t *testing.T) {
	// Enable output capture.
	flags.Output = true

	tests := []struct {
		name       string
		timeout    int
		command    string
		args       []string
		wantStdout string
		wantStderr string
		wantErr    bool
	}{
		{
			name:       "successful command",
			timeout:    5,
			command:    "echo",
			args:       []string{"hello world"},
			wantStdout: "hello world\n",
		},
		{
			name:       "command with arguments",
			timeout:    5,
			command:    "printf",
			args:       []string{"%s %s", "hello", "world"},
			wantStdout: "hello world",
		},
		{
			name:       "stderr output",
			timeout:    5,
			command:    "sh",
			args:       []string{"-c", "echo error-message >&2"},
			wantStderr: "error-message\n",
		},
		{
			name:    "invalid command",
			timeout: 5,
			command: "command-that-does-not-exist",
			wantErr: true,
		},
		{
			name:    "command timeout",
			timeout: 1,
			command: "sleep",
			args:    []string{"5"},
			wantErr: true,
		},
		{
			name:       "command without timeout",
			timeout:    0,
			command:    "echo",
			args:       []string{"no timeout"},
			wantStdout: "no timeout\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout-*")
			if err != nil {
				t.Fatal(err)
			}
			defer stdoutFile.Close()

			stderrFile, err := os.CreateTemp(t.TempDir(), "stderr-*")
			if err != nil {
				t.Fatal(err)
			}
			defer stderrFile.Close()

			start := time.Now()

			err = RunCommand(
				stdoutFile,
				stderrFile,
				tt.timeout,
				tt.command,
				tt.args...,
			)

			duration := time.Since(start)

			if tt.wantErr && err == nil {
				t.Fatal("expected an error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Read stdout.
			if _, err := stdoutFile.Seek(0, 0); err != nil {
				t.Fatal(err)
			}

			var stdout bytes.Buffer
			if _, err := stdout.ReadFrom(stdoutFile); err != nil {
				t.Fatal(err)
			}

			// Read stderr.
			if _, err := stderrFile.Seek(0, 0); err != nil {
				t.Fatal(err)
			}

			var stderr bytes.Buffer
			if _, err := stderr.ReadFrom(stderrFile); err != nil {
				t.Fatal(err)
			}

			// Check stdout.
			if got := stdout.String(); got != tt.wantStdout {
				t.Errorf(
					"stdout mismatch:\n got: %q\nwant: %q",
					got,
					tt.wantStdout,
				)
			}

			// Check stderr.
			if got := stderr.String(); got != tt.wantStderr {
				t.Errorf(
					"stderr mismatch:\n got: %q\nwant: %q",
					got,
					tt.wantStderr,
				)
			}

			// Check timeout behavior.
			if tt.name == "command timeout" {
				if duration >= 5*time.Second {
					t.Errorf("command did not time out quickly: %v", duration)
				}

				if duration < 1*time.Second {
					t.Errorf("command terminated unexpectedly early: %v", duration)
				}
			}
		})
	}
}
