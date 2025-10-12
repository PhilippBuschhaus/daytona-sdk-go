package main

import (
	"context"
	"fmt"
	"log"
	"time"

	daytona "github.com/PhilippBuschhaus/daytona-sdk-go"
)

func main() {
	// Create SDK client
	// API key is loaded from DAYTONA_API_KEY environment variable
	client, err := daytona.NewClient(&daytona.Config{})
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	fmt.Println("=== Daytona SDK Basic Example ===\n")

	// List existing sandboxes
	fmt.Println("Listing sandboxes...")
	sandboxes, err := client.ListSandboxes(ctx)
	if err != nil {
		log.Printf("Failed to list sandboxes: %v\n", err)
	} else {
		fmt.Printf("Found %d sandboxes\n", len(sandboxes))
		for _, s := range sandboxes {
			fmt.Printf("  - %s (User: %s, State: %v)\n", s.GetId(), s.GetUser(), s.State)
		}
	}

	// Create a new sandbox
	fmt.Println("\nCreating a new sandbox...")
	// Most fields are optional - only Target is typically required
	createReq := &daytona.CreateSandboxRequest{
		Target: daytona.StringPtr("eu"), // Required: deployment region (eu, us, etc.)

		// Optional fields (uncomment to use):
		// User:     daytona.StringPtr("daytona"),           // Defaults to account default
		// Snapshot: daytona.StringPtr("daytonaio/sandbox:0.4.3"), // Defaults to latest
		// Public:   daytona.BoolPtr(false),                 // Defaults to false
		// Labels: map[string]string{                        // Optional metadata
		// 	"created_by": "go_sdk",
		// 	"example":    "basic",
		// },
	}

	sandbox, err := client.CreateSandbox(ctx, createReq)
	if err != nil {
		log.Fatal("Failed to create sandbox:", err)
	}
	fmt.Printf("✓ Created sandbox: %s\n", sandbox.GetId())

	// Wait for sandbox to be ready
	fmt.Println("\nWaiting for sandbox to be ready...")
	sandbox, err = client.WaitForSandboxReady(ctx, sandbox.GetId(), 5*time.Minute)
	if err != nil {
		log.Printf("Failed to wait for sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox is ready!")
	}

	// Execute a command
	fmt.Println("\nExecuting command in sandbox...")
	execReq := &daytona.ExecuteCommandRequest{
		Command: "echo 'Hello from Daytona SDK!'",
	}

	execResp, err := client.ExecuteCommand(ctx, sandbox.GetId(), execReq)
	if err != nil {
		log.Printf("Failed to execute command: %v\n", err)
	} else {
		fmt.Printf("Command output: %s\n", execResp.Result)
		fmt.Printf("Exit code: %.0f\n", execResp.ExitCode)
	}

	// Create a file
	fmt.Println("\nCreating a file...")
	content := []byte("Hello, World!\nThis file was created by the Daytona Go SDK.")
	err = client.WriteFile(ctx, sandbox.GetId(), "/tmp/hello.txt", content)
	if err != nil {
		log.Printf("Failed to write file: %v\n", err)
	} else {
		fmt.Println("✓ File created")
	}

	// Read the file back
	fmt.Println("\nReading file...")
	data, err := client.ReadFile(ctx, sandbox.GetId(), "/tmp/hello.txt")
	if err != nil {
		log.Printf("Failed to read file: %v\n", err)
	} else {
		fmt.Printf("File content:\n%s\n", string(data))
	}

	// List files
	fmt.Println("\nListing files in /tmp...")
	files, err := client.ListFiles(ctx, sandbox.GetId(), "/tmp")
	if err != nil {
		log.Printf("Failed to list files: %v\n", err)
	} else {
		for _, f := range files {
			fmt.Printf("  - %s", f.Name)
			if f.IsDir {
				fmt.Print(" (dir)")
			}
			fmt.Printf(" [%.0f bytes]", f.Size)
			fmt.Println()
		}
	}

	// Clean up
	fmt.Println("\nCleaning up...")
	err = client.DeleteSandbox(ctx, sandbox.GetId())
	if err != nil {
		log.Printf("Failed to delete sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox deleted")
	}

	fmt.Println("\n=== Example Complete ===")
}
