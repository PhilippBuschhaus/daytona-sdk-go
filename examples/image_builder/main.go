package main

import (
	"context"
	"fmt"
	"log"
	"time"

	daytona "github.com/PhilippBuschhaus/daytona-sdk-go"
)

func main() {
	fmt.Println("=== Declarative Image Builder Examples ===\n")

	// Create SDK client
	client, err := daytona.NewClient(&daytona.Config{})
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Example 1: Simple Python data science image
	fmt.Println("Example 1: Data Science Image")
	fmt.Println("------------------------------")

	dataScience := daytona.PythonDataScience("3.11")
	fmt.Printf("Generated Dockerfile:\n%s\n", dataScience.Build())

	// Example 2: Custom web application image
	fmt.Println("\nExample 2: Web Application Image")
	fmt.Println("---------------------------------")

	webApp := daytona.DebianSlim("3.11").
		PipInstall([]string{"fastapi", "uvicorn[standard]", "sqlalchemy", "alembic"}).
		AptInstall([]string{"postgresql-client"}).
		Env("APP_ENV", "production").
		Env("PORT", "8000").
		Workdir("/app").
		Copy(".", "/app").
		Expose(8000).
		Cmd([]string{"uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"})

	fmt.Printf("Generated Dockerfile:\n%s\n", webApp.Build())

	// Example 3: Multi-stage Go build
	fmt.Println("\nExample 3: Go Application Image")
	fmt.Println("--------------------------------")

	goApp := daytona.Base("golang:1.21 AS builder").
		Workdir("/build").
		Copy("go.mod", ".").
		Copy("go.sum", ".").
		RunCommand("go mod download").
		Copy(".", ".").
		RunCommand("go build -o app .").
		// Note: In a real multi-stage build, you'd need a second FROM
		// This is just demonstrating the builder pattern
		Workdir("/app")

	fmt.Printf("Generated Dockerfile:\n%s\n", goApp.Build())

	// Example 4: Node.js with custom packages
	fmt.Println("\nExample 4: Node.js Application")
	fmt.Println("-------------------------------")

	nodeApp := daytona.NodeJS("20").
		Copy("package*.json", "./").
		RunCommand("npm ci --only=production").
		Copy(".", ".").
		Expose(3000).
		Cmd([]string{"node", "server.js"})

	fmt.Printf("Generated Dockerfile:\n%s\n", nodeApp.Build())

	// Example 5: Actually create a sandbox with a custom image
	fmt.Println("\nExample 5: Creating Sandbox with Custom Image")
	fmt.Println("----------------------------------------------")

	// Build a custom image for machine learning
	mlImage := daytona.DebianSlim("3.11").
		Label("project", "ml-sandbox").
		Label("created-by", "go-sdk").
		AptInstall([]string{"gcc", "g++", "gfortran"}).
		PipInstall([]string{
			"tensorflow",
			"torch",
			"transformers",
			"datasets",
			"jupyter",
			"ipykernel",
		}).
		Env("PYTHONUNBUFFERED", "1").
		Workdir("/workspace").
		RunCommand("jupyter notebook --generate-config").
		RunCommand("echo \"c.NotebookApp.token = ''\" >> ~/.jupyter/jupyter_notebook_config.py").
		Expose(8888)

	fmt.Println("Creating sandbox with ML image...")
	createReq := &daytona.CreateSandboxRequest{
		Target:            daytona.StringPtr("eu"),
		DockerfileContent: daytona.StringPtr(mlImage.Build()),
		Labels: map[string]string{
			"example": "image_builder",
			"type":    "ml",
		},
	}

	sandbox, err := client.CreateSandbox(ctx, createReq)
	if err != nil {
		log.Printf("Failed to create sandbox: %v\n", err)
		fmt.Println("Note: Sandbox creation failed, but the image builder examples above show the generated Dockerfiles.")
		return
	}

	sandboxID := sandbox.GetId()
	fmt.Printf("✓ Created sandbox: %s\n", sandboxID)

	// Wait for sandbox to be ready
	fmt.Println("Waiting for sandbox to be ready...")
	sandbox, err = client.WaitForSandboxReady(ctx, sandboxID, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to wait for sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox is ready!")

		// Test that our ML packages are installed
		fmt.Println("\nVerifying ML packages are installed...")
		execReq := &daytona.ExecuteCommandRequest{
			Command: "python -c 'import tensorflow as tf; import torch; print(f\"TensorFlow: {tf.__version__}\"); print(f\"PyTorch: {torch.__version__}\")'",
			Timeout: 30.0,
		}

		result, err := client.ExecuteCommand(ctx, sandboxID, execReq)
		if err != nil {
			log.Printf("Failed to verify packages: %v\n", err)
		} else {
			fmt.Printf("Package versions:\n%s\n", result.Result)
		}
	}

	// Clean up
	fmt.Println("\nCleaning up...")
	err = client.DeleteSandbox(ctx, sandboxID)
	if err != nil {
		log.Printf("Failed to delete sandbox: %v\n", err)
	} else {
		fmt.Println("✓ Sandbox deleted")
	}

	fmt.Println("\n=== Example Complete ===")
}
