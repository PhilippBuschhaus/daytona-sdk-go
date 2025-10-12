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

	fmt.Println("=== File Operations Example ===\n")

	// Create a new sandbox for file operations
	fmt.Println("Creating sandbox for file operations...")
	createReq := &daytona.CreateSandboxRequest{
		Target: daytona.StringPtr("eu"), // Required: deployment region

		// Optional: Labels for organization
		Labels: map[string]string{
			"example": "file_operations",
		},
	}

	sandbox, err := client.CreateSandbox(ctx, createReq)
	if err != nil {
		log.Fatal("Failed to create sandbox:", err)
	}
	sandboxID := sandbox.GetId()
	fmt.Printf("✓ Created sandbox: %s\n", sandboxID)

	// Wait for sandbox to be ready
	fmt.Println("Waiting for sandbox to be ready...")
	sandbox, err = client.WaitForSandboxReady(ctx, sandboxID, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to wait for sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox is ready!\n")
	}

	// Create a directory
	fmt.Println("Creating directory...")
	err = client.CreateFolder(ctx, sandboxID, "/tmp/test")
	if err != nil {
		log.Printf("Failed to create directory: %v\n", err)
	} else {
		fmt.Println("✓ Directory created")
	}

	// Write a file
	fmt.Println("\nWriting file...")
	content := []byte(`# Test File
This is a test file created by the Daytona Go SDK.

## Features
- File operations
- Directory management
- Content search
`)
	err = client.WriteFile(ctx, sandboxID, "/tmp/test/README.md", content)
	if err != nil {
		log.Printf("Failed to write file: %v\n", err)
	} else {
		fmt.Println("✓ File written")
	}

	// Read the file back
	fmt.Println("\nReading file...")
	data, err := client.ReadFile(ctx, sandboxID, "/tmp/test/README.md")
	if err != nil {
		log.Printf("Failed to read file: %v\n", err)
	} else {
		fmt.Printf("File content (%d bytes):\n%s\n", len(data), string(data))
	}

	// Get file info
	fmt.Println("\nGetting file info...")
	info, err := client.GetFileInfo(ctx, sandboxID, "/tmp/test/README.md")
	if err != nil {
		log.Printf("Failed to get file info: %v\n", err)
	} else {
		fmt.Printf("File: %s\n", info.Name)
		fmt.Printf("  Size: %.0f bytes\n", info.Size)
		fmt.Printf("  Is Directory: %v\n", info.IsDir)
		fmt.Printf("  Modified: %s\n", info.ModTime)
		fmt.Printf("  Permissions: %s\n", info.Permissions)
	}

	// List files
	fmt.Println("\nListing files in /tmp/test...")
	files, err := client.ListFiles(ctx, sandboxID, "/tmp/test")
	if err != nil {
		log.Printf("Failed to list files: %v\n", err)
	} else {
		for _, f := range files {
			fmt.Printf("  - %s", f.Name)
			if f.IsDir {
				fmt.Print(" [DIR]")
			} else {
				fmt.Printf(" [%.0f bytes]", f.Size)
			}
			fmt.Println()
		}
	}

	// Move/rename file
	fmt.Println("\nMoving file...")
	err = client.MoveFile(ctx, sandboxID, "/tmp/test/README.md", "/tmp/test/documentation.md")
	if err != nil {
		log.Printf("Failed to move file: %v\n", err)
	} else {
		fmt.Println("✓ File moved")
	}

	// Search in files
	fmt.Println("\nSearching for 'SDK' in files...")
	matches, err := client.FindInFiles(ctx, sandboxID, "SDK", "/tmp/test")
	if err != nil {
		log.Printf("Failed to search in files: %v\n", err)
	} else {
		fmt.Printf("Found %d matches\n", len(matches))
		for _, m := range matches {
			fmt.Printf("  - %s: line %.0f - %s\n", m.File, m.Line, m.Content)
		}
	}

	// Delete file
	fmt.Println("\nDeleting file...")
	err = client.DeleteFile(ctx, sandboxID, "/tmp/test/documentation.md")
	if err != nil {
		log.Printf("Failed to delete file: %v\n", err)
	} else {
		fmt.Println("✓ File deleted")
	}

	// Clean up sandbox
	fmt.Println("\nCleaning up...")
	err = client.DeleteSandbox(ctx, sandboxID)
	if err != nil {
		log.Printf("Failed to delete sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox deleted")
	}

	fmt.Println("\n=== Example Complete ===")
}
