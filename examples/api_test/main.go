package main

import (
	"context"
	"fmt"
	"log"
	"os"

	api "github.com/PhilippBuschhaus/daytona-sdk-go/apiclient"
)

func main() {
	// Load API key from environment variable
	apiKey := os.Getenv("DAYTONA_API_KEY")
	if apiKey == "" {
		log.Fatal("DAYTONA_API_KEY environment variable is required")
	}

	// Create API configuration
	config := api.NewConfiguration()
	config.Host = "api.daytona.io"
	config.Scheme = "https"
	config.AddDefaultHeader("Authorization", "Bearer "+apiKey)

	// Create API client
	client := api.NewAPIClient(config)
	ctx := context.Background()

	fmt.Println("=== Daytona API Client Test ===\n")
	fmt.Println("✓ API client created successfully")

	// Try to list sandboxes
	fmt.Println("\nAttempting to list sandboxes...")
	sandboxAPI := client.SandboxAPI
	sandboxes, resp, err := sandboxAPI.ListSandboxes(ctx).Execute()
	if err != nil {
		if resp != nil {
			fmt.Printf("API Error - Status: %d\n", resp.StatusCode)
		}
		log.Printf("Failed to list sandboxes: %v\n", err)
	} else {
		fmt.Printf("✓ Successfully connected to API\n")
		fmt.Printf("✓ Found %d sandboxes\n", len(sandboxes))

		for i, sandbox := range sandboxes {
			fmt.Printf("\nSandbox %d:\n", i+1)
			fmt.Printf("  ID: %s\n", sandbox.GetId())
			fmt.Printf("  User: %s\n", sandbox.GetUser())
			fmt.Printf("  Target: %s\n", sandbox.GetTarget())
			fmt.Printf("  CPU: %.1f\n", sandbox.GetCpu())
			fmt.Printf("  Memory: %.1f\n", sandbox.GetMemory())
			fmt.Printf("  Public: %v\n", sandbox.GetPublic())
		}
	}

	// Try creating a simple sandbox
	fmt.Println("\nAttempting to create a sandbox...")

	// Helper function to get pointer to a value
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int32) *int32 { return &i }
	boolPtr := func(b bool) *bool { return &b }

	envVars := map[string]string{
		"TEST": "true",
	}
	labels := map[string]string{
		"created_by": "go_sdk",
		"test":       "true",
	}

	createReq := api.CreateSandbox{
		User:   strPtr("sdk-test"),
		Target: strPtr("ubuntu:latest"),
		Cpu:    intPtr(1),
		Memory: intPtr(2048),
		Disk:   intPtr(10240),
		Public: boolPtr(false),
		Env:    &envVars,
		Labels: &labels,
	}

	newSandbox, resp, err := sandboxAPI.CreateSandbox(ctx).CreateSandbox(createReq).Execute()
	if err != nil {
		if resp != nil {
			fmt.Printf("API Error - Status: %d\n", resp.StatusCode)
		}
		log.Printf("Failed to create sandbox: %v\n", err)
	} else {
		fmt.Printf("✓ Sandbox created successfully!\n")
		fmt.Printf("  ID: %s\n", newSandbox.GetId())
		fmt.Printf("  User: %s\n", newSandbox.GetUser())

		// Clean up - delete the sandbox
		fmt.Printf("\nDeleting sandbox %s...\n", newSandbox.GetId())
		_, resp, err := sandboxAPI.DeleteSandbox(ctx, newSandbox.GetId()).Execute()
		if err != nil {
			if resp != nil {
				fmt.Printf("API Error - Status: %d\n", resp.StatusCode)
			}
			log.Printf("Failed to delete sandbox: %v\n", err)
		} else {
			fmt.Println("✓ Sandbox deleted successfully")
		}
	}

	fmt.Println("\n=== Test Complete ===")
}
