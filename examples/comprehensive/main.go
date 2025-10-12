package main

import (
	"context"
	"fmt"
	"log"
	"time"

	daytona "github.com/PhilippBuschhaus/daytona-sdk-go"
	api "github.com/PhilippBuschhaus/daytona-sdk-go/apiclient"
)

func main() {
	fmt.Println("=== Daytona Go SDK Comprehensive Example ===\n")

	// Create SDK client with configuration
	// API key is loaded from DAYTONA_API_KEY environment variable
	config := &daytona.Config{
		Timeout: 30 * time.Second,
	}

	client, err := daytona.NewClient(config)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()
	fmt.Println("✓ SDK Client initialized\n")

	// =====================================
	// SANDBOX OPERATIONS
	// =====================================
	fmt.Println("=== Sandbox Operations ===")

	// List existing sandboxes
	fmt.Println("Listing sandboxes...")
	sandboxes, err := client.ListSandboxes(ctx)
	if err != nil {
		log.Printf("Failed to list sandboxes: %v\n", err)
	} else {
		fmt.Printf("Found %d existing sandboxes\n", len(sandboxes))
	}

	// Create a new sandbox
	fmt.Println("\nCreating a new sandbox...")
	createReq := &daytona.CreateSandboxRequest{
		Target: daytona.StringPtr("eu"), // Required: deployment region

		// Optional: Environment variables
		Env: map[string]string{
			"NODE_ENV":    "development",
			"DEBUG":       "true",
			"API_VERSION": "v1",
		},

		// Optional: Metadata labels
		Labels: map[string]string{
			"project":     "sdk-demo",
			"environment": "development",
		},

		// Optional: Auto-stop configuration
		AutoStopInterval: daytona.IntPtr(30), // Stop after 30 minutes of inactivity

		// Other optional fields:
		// User:     daytona.StringPtr("daytona"),           // Defaults to account default
		// Snapshot: daytona.StringPtr("daytonaio/sandbox:0.4.3"), // Defaults to latest
		// Public:   daytona.BoolPtr(false),                 // Defaults to false
	}

	sandbox, err := client.CreateSandbox(ctx, createReq)
	if err != nil {
		log.Printf("Failed to create sandbox: %v\n", err)
		// For demo purposes, we'll continue with a fake ID
		sandbox = &api.Sandbox{
			Id: "demo-sandbox-123",
		}
	} else {
		fmt.Printf("✓ Created sandbox: %s\n", sandbox.GetId())

		// Wait for sandbox to be ready
		fmt.Println("Waiting for sandbox to be ready...")
		sandbox, err = client.WaitForSandboxReady(ctx, sandbox.GetId(), 5*time.Minute)
		if err != nil {
			log.Printf("Failed to wait for sandbox: %v\n", err)
		} else {
			fmt.Println("✓ Sandbox is ready!")
		}
	}

	sandboxID := sandbox.GetId()

	// =====================================
	// FILE OPERATIONS
	// =====================================
	fmt.Println("\n=== File Operations ===")

	// Create directory structure
	fmt.Println("Creating directory structure...")
	dirs := []string{
		"/tmp/project",
		"/tmp/project/src",
		"/tmp/project/tests",
		"/tmp/project/docs",
	}

	for _, dir := range dirs {
		err = client.CreateFolder(ctx, sandboxID, dir)
		if err != nil {
			log.Printf("Failed to create %s: %v\n", dir, err)
		}
	}

	// Write files
	fmt.Println("Writing project files...")
	files := map[string]string{
		"/tmp/project/README.md": `# Demo Project
This project was created using the Daytona Go SDK.

## Features
- File management
- Command execution
- Session handling
`,
		"/tmp/project/src/main.py": `#!/usr/bin/env python3
import json

def main():
    data = {
        "message": "Hello from Daytona!",
        "sdk": "Go SDK",
        "version": "1.0.0"
    }
    print(json.dumps(data, indent=2))

if __name__ == "__main__":
    main()
`,
		"/tmp/project/Makefile": `all: test

test:
	@echo "Running tests..."
	python3 src/main.py

clean:
	@echo "Cleaning up..."
	rm -rf __pycache__
`,
	}

	for path, content := range files {
		err = client.WriteFile(ctx, sandboxID, path, []byte(content))
		if err != nil {
			log.Printf("Failed to write %s: %v\n", path, err)
		}
	}
	fmt.Println("✓ Files created")

	// List files
	fmt.Println("\nListing project files...")
	projectFiles, err := client.ListFiles(ctx, sandboxID, "/tmp/project")
	if err != nil {
		log.Printf("Failed to list files: %v\n", err)
	} else {
		for _, f := range projectFiles {
			fmt.Printf("  - %s", f.Name)
			if f.IsDir {
				fmt.Print(" [DIR]")
			} else {
				fmt.Printf(" [%.0f bytes]", f.Size)
			}
			fmt.Println()
		}
	}

	// =====================================
	// COMMAND EXECUTION
	// =====================================
	fmt.Println("\n=== Command Execution ===")

	// Execute various commands
	commands := []struct {
		desc    string
		command string
		cwd     string
	}{
		{"Check Python version", "python3 --version", ""},
		{"List tmp", "ls -la", "/tmp"},
		{"Run Python script", "python3 src/main.py", "/tmp/project"},
		{"Run Makefile", "make test", "/tmp/project"},
		{"Check system info", "uname -a", ""},
		{"Check disk usage", "df -h", ""},
	}

	for _, cmd := range commands {
		fmt.Printf("\n%s:\n", cmd.desc)
		result, err := client.ExecuteCommand(ctx, sandboxID, &daytona.ExecuteCommandRequest{
			Command: cmd.command,
			Cwd:     cmd.cwd,
			Timeout: 10.0,
		})
		if err != nil {
			log.Printf("  Failed: %v\n", err)
		} else {
			fmt.Printf("  Output: %s\n", result.Result)
			if result.ExitCode != 0 {
				fmt.Printf("  Exit code: %.0f\n", result.ExitCode)
			}
		}
	}

	// =====================================
	// SESSION MANAGEMENT
	// =====================================
	fmt.Println("\n=== Session Management ===")

	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	fmt.Printf("Creating session %s...\n", sessionID)

	err = client.CreateSession(ctx, sandboxID, sessionID)
	if err != nil {
		log.Printf("Failed to create session: %v\n", err)
	} else {
		fmt.Println("✓ Session created")

		// Execute commands in session
		sessionCmds := []string{
			"cd /tmp/project",
			"export PROJECT_NAME='DaytonaDemo'",
			"echo \"Project: $PROJECT_NAME\"",
			"ls -la",
		}

		for _, cmd := range sessionCmds {
			fmt.Printf("  Executing: %s\n", cmd)
			_, err = client.ExecuteSessionCommand(ctx, sandboxID, sessionID, cmd)
			if err != nil {
				log.Printf("    Failed: %v\n", err)
			}
		}

		// Clean up session
		err = client.DeleteSession(ctx, sandboxID, sessionID)
		if err != nil {
			log.Printf("Failed to delete session: %v\n", err)
		} else {
			fmt.Println("✓ Session deleted")
		}
	}

	// =====================================
	// SEARCH OPERATIONS
	// =====================================
	fmt.Println("\n=== Search Operations ===")

	fmt.Println("Searching for 'Daytona' in project files...")
	matches, err := client.FindInFiles(ctx, sandboxID, "Daytona", "/tmp/project")
	if err != nil {
		log.Printf("Failed to search: %v\n", err)
	} else {
		fmt.Printf("Found %d matches:\n", len(matches))
		for _, m := range matches {
			fmt.Printf("  - %s (line %.0f): %s\n", m.File, m.Line, m.Content)
		}
	}

	// =====================================
	// SANDBOX LIFECYCLE
	// =====================================
	fmt.Println("\n=== Sandbox Lifecycle ===")

	// Stop sandbox
	fmt.Println("Stopping sandbox...")
	err = client.StopSandbox(ctx, sandboxID)
	if err != nil {
		log.Printf("Failed to stop sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox stopped")
	}

	// Start sandbox again
	fmt.Println("Starting sandbox again...")
	_, err = client.StartSandbox(ctx, sandboxID)
	if err != nil {
		log.Printf("Failed to start sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox started")
	}

	// Archive sandbox
	fmt.Println("Archiving sandbox...")
	err = client.ArchiveSandbox(ctx, sandboxID)
	if err != nil {
		log.Printf("Failed to archive sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox archived")
	}

	// Delete sandbox
	fmt.Println("Deleting sandbox...")
	err = client.DeleteSandbox(ctx, sandboxID) // force delete
	if err != nil {
		log.Printf("Failed to delete sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox deleted")
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("Successfully demonstrated all SDK features!")
}
