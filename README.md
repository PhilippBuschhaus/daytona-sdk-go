# Daytona Go SDK (Unofficial)

⚠️ **UNOFFICIAL SDK - IN DEVELOPMENT** ⚠️

This is an unofficial Go SDK for the [Daytona](https://www.daytona.io) platform. It is currently in active development and not affiliated with or endorsed by Daytona Platforms Inc.

## Overview

The Daytona Go SDK provides a Go interface for interacting with the Daytona API, enabling developers to manage sandboxes, execute commands, handle files, and more programmatically.

## Features

- 🏗️ **Sandbox Management** - Create, start, stop, archive, and delete sandboxes
- 📁 **File Operations** - Upload, download, create folders, and manage files
- 💻 **Command Execution** - Run commands and manage sessions
- 🔧 **Custom Images** - Build custom sandbox environments with pre-installed packages
- 🔄 **Session Management** - Maintain stateful command execution contexts

## Installation

```bash
go get github.com/PhilippBuschhaus/daytona-sdk-go
```


## Configuration

The SDK automatically loads environment variables from `.env` files using [godotenv](https://github.com/joho/godotenv). 

Create a `.env` file in your project root or add into the configuration struct:

```bash
DAYTONA_API_KEY=your_api_key_here
DAYTONA_API_URL=https://app.daytona.io/api  # Optional, defaults to this URL
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    daytona "github.com/PhilippBuschhaus/daytona-sdk-go"
)

func main() {
    // Create client (automatically loads .env file)
    client, err := daytona.NewClient(&daytona.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Create a sandbox
    sandbox, err := client.CreateSandbox(ctx, &daytona.CreateSandboxRequest{
        Target: daytona.StringPtr("eu"), 
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created sandbox: %s\n", sandbox.GetId())
    
    // Execute a command
    result, err := client.ExecuteCommand(ctx, sandbox.GetId(), &daytona.ExecuteCommandRequest{
        Command: "echo 'Hello from Daytona!'",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Output: %s\n", result.Result)
    
    // Clean up
    err = client.DeleteSandbox(ctx, sandbox.GetId(), true)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Examples

The SDK includes comprehensive examples in the `examples/` directory demonstrating:

- Basic sandbox operations and command execution
- File upload, download, and manipulation
- Custom Docker image building with pre-installed packages
- Data generation and processing workflows
- Session management for stateful operations
- And much more...

Run any example:
```bash
# Examples automatically load the .env file
go run examples/basic/main.go
go run examples/comprehensive/main.go
```

## Declarative Image Builder

The SDK includes a powerful declarative builder for creating custom Docker images, similar to the TypeScript SDK:

### Basic Usage

```go
// Use preset builders
image := daytona.PythonDataScience("3.11")  // Includes numpy, pandas, scikit-learn, etc.

// Or build custom images with method chaining
image := daytona.DebianSlim("3.11").
    AptInstall([]string{"postgresql-client"}).
    PipInstall([]string{"fastapi", "uvicorn", "sqlalchemy"}).
    EnvVars(                           // Set multiple env vars at once
        "APP_ENV", "production",
        "PORT", "8000",
        "WORKERS", "4",
    ).
    Workdir("/app").
    Expose(8000)

// Create sandbox with the custom image
sandbox, err := client.CreateSandbox(ctx, &daytona.CreateSandboxRequest{
    Target: daytona.StringPtr("eu"),
    DockerfileContent: daytona.StringPtr(image.Build()),
})
```

### Available Builders

- `DebianSlim(pythonVersion)` - Debian-based Python image
- `UbuntuSlim(pythonVersion)` - Ubuntu-based Python image  
- `Base(imageName)` - Start from any base image
- `FromDockerfile(content)` - Use existing Dockerfile
- `PythonDataScience(version)` - Pre-configured for data science
- `PythonWeb(version)` - Pre-configured for web development
- `NodeJS(version)` - Node.js development environment
- `Go(version)` - Go development environment

### Builder Methods

All methods return the Image instance for chaining:

- `PipInstall(packages)` - Install Python packages
- `AptInstall(packages)` - Install system packages
- `RunCommand(cmd)` - Run shell commands
- `Env(key, value)` - Set a single environment variable
- `EnvVars(k1, v1, k2, v2, ...)` - Set multiple environment variables as pairs
- `EnvMap(map[string]string)` - Set environment variables from a map
- `Workdir(path)` - Set working directory
- `Copy(src, dest)` - Copy files into image
- `Expose(port)` - Expose container ports
- `Label(key, value)` - Add metadata labels
- `User(username)` - Set user for commands
- `Entrypoint(cmd)` - Set container entrypoint
- `Cmd(cmd)` - Set default command
- `Build()` - Get the generated Dockerfile

## API Reference

### Client Creation
```go
client, err := daytona.NewClient(&daytona.Config{
    APIKey:  "your_api_key",     // Optional, uses DAYTONA_API_KEY env var
    BaseURL: "https://...",      // Optional, uses DAYTONA_API_URL env var
    Timeout: 30 * time.Second,   // Optional, defaults to 30s
})
```

### Sandbox Operations
- `CreateSandbox(ctx, request)` - Create a new sandbox
- `GetSandbox(ctx, sandboxID)` - Get sandbox details
- `ListSandboxes(ctx)` - List all sandboxes
- `StartSandbox(ctx, sandboxID)` - Start a sandbox
- `StopSandbox(ctx, sandboxID)` - Stop a sandbox
- `ArchiveSandbox(ctx, sandboxID)` - Archive a sandbox
- `DeleteSandbox(ctx, sandboxID, force)` - Delete a sandbox
- `WaitForSandboxReady(ctx, sandboxID, timeout)` - Wait for sandbox to be ready

### File Operations
- `WriteFile(ctx, sandboxID, path, content)` - Write a file
- `ReadFile(ctx, sandboxID, path)` - Read a file
- `CreateFolder(ctx, sandboxID, path)` - Create a directory
- `DeleteFile(ctx, sandboxID, path)` - Delete a file or directory
- `MoveFile(ctx, sandboxID, source, dest)` - Move/rename a file
- `ListFiles(ctx, sandboxID, path)` - List files in a directory
- `GetFileInfo(ctx, sandboxID, path)` - Get file information
- `FindInFiles(ctx, sandboxID, pattern, path)` - Search in files

### Command Execution
- `ExecuteCommand(ctx, sandboxID, request)` - Execute a command
- `CreateSession(ctx, sandboxID, sessionID)` - Create a session
- `ExecuteSessionCommand(ctx, sandboxID, sessionID, command)` - Execute in session
- `DeleteSession(ctx, sandboxID, sessionID)` - Delete a session

## Development Status

This SDK is under active development. Current limitations:

- Not all Daytona API endpoints are implemented
- API may change as the SDK evolves
- Limited testing coverage
- No official support from Daytona

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

The SDK includes the Daytona API client bundled within the `apiclient` directory

### Updating the API Client

To update the bundled API client to the latest version from the Daytona repository:

```bash
./scripts/update-apiclient.sh
```

This script will fetch the latest API client from GitHub and update the local copy.

## Disclaimer

This is an unofficial SDK and is not affiliated with, endorsed by, or supported by Daytona Platforms Inc. Use at your own risk.

## Resources

- [Official Daytona Documentation](https://www.daytona.io/docs)
- [Daytona API Reference](https://www.daytona.io/docs/api)
- [Official TypeScript SDK](https://github.com/daytonaio/daytona/tree/main/libs/sdk-typescript)