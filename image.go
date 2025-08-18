package daytona

import (
	"fmt"
	"strings"
)

// Image provides a declarative builder for creating Docker images
// Similar to the TypeScript SDK's Image class
type Image struct {
	dockerfile []string
	envVars    map[string]string
}

// NewImage creates a new Image builder
func NewImage() *Image {
	return &Image{
		dockerfile: []string{},
		envVars:    make(map[string]string),
	}
}

// Base creates an image from an existing base image
func Base(image string) *Image {
	i := NewImage()
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("FROM %s", image))
	return i
}

// DebianSlim creates a Debian-based Python image
func DebianSlim(pythonVersion string) *Image {
	i := NewImage()
	if pythonVersion == "" {
		pythonVersion = "3.11"
	}
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("FROM python:%s-slim-bookworm", pythonVersion))
	
	// Update system and install basic tools
	i.dockerfile = append(i.dockerfile, 
		"RUN apt-get update && apt-get install -y --no-install-recommends \\",
		"    build-essential \\",
		"    curl \\",
		"    git \\",
		"    && rm -rf /var/lib/apt/lists/*")
	
	// Upgrade pip
	i.dockerfile = append(i.dockerfile, "RUN pip install --upgrade pip")
	
	return i
}

// UbuntuSlim creates an Ubuntu-based Python image
func UbuntuSlim(pythonVersion string) *Image {
	i := NewImage()
	if pythonVersion == "" {
		pythonVersion = "3.11"
	}
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("FROM python:%s-slim", pythonVersion))
	
	// Update system
	i.dockerfile = append(i.dockerfile, 
		"RUN apt-get update && apt-get install -y --no-install-recommends \\",
		"    build-essential \\",
		"    curl \\",
		"    git \\",
		"    && rm -rf /var/lib/apt/lists/*")
	
	// Upgrade pip
	i.dockerfile = append(i.dockerfile, "RUN pip install --upgrade pip")
	
	return i
}

// FromDockerfile creates an image from an existing Dockerfile content
func FromDockerfile(dockerfileContent string) *Image {
	i := NewImage()
	i.dockerfile = append(i.dockerfile, dockerfileContent)
	return i
}

// PipInstall adds Python packages to the image
func (i *Image) PipInstall(packages []string) *Image {
	if len(packages) > 0 {
		packagesStr := strings.Join(packages, " ")
		i.dockerfile = append(i.dockerfile, fmt.Sprintf("RUN pip install %s", packagesStr))
	}
	return i
}

// PipInstallFromRequirements installs packages from a requirements.txt file
func (i *Image) PipInstallFromRequirements(requirementsPath string) *Image {
	// First copy the requirements file
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("COPY %s /tmp/requirements.txt", requirementsPath))
	// Then install from it
	i.dockerfile = append(i.dockerfile, "RUN pip install -r /tmp/requirements.txt")
	return i
}

// AptInstall installs system packages using apt
func (i *Image) AptInstall(packages []string) *Image {
	if len(packages) > 0 {
		packagesStr := strings.Join(packages, " ")
		i.dockerfile = append(i.dockerfile, 
			fmt.Sprintf("RUN apt-get update && apt-get install -y %s && rm -rf /var/lib/apt/lists/*", packagesStr))
	}
	return i
}

// RunCommands executes shell commands during image build
func (i *Image) RunCommands(commands ...string) *Image {
	for _, cmd := range commands {
		i.dockerfile = append(i.dockerfile, fmt.Sprintf("RUN %s", cmd))
	}
	return i
}

// RunCommand executes a single shell command during image build
func (i *Image) RunCommand(command string) *Image {
	return i.RunCommands(command)
}

// Env sets environment variables
func (i *Image) Env(key, value string) *Image {
	i.envVars[key] = value
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("ENV %s=%s", key, value))
	return i
}

// EnvMap sets multiple environment variables from a map
func (i *Image) EnvMap(vars map[string]string) *Image {
	for key, value := range vars {
		i.Env(key, value)
	}
	return i
}

// EnvVars sets multiple environment variables from key-value pairs
// Usage: EnvVars("KEY1", "value1", "KEY2", "value2", ...)
func (i *Image) EnvVars(kvPairs ...string) *Image {
	if len(kvPairs)%2 != 0 {
		// Ignore if not even number of arguments
		return i
	}
	for j := 0; j < len(kvPairs); j += 2 {
		i.Env(kvPairs[j], kvPairs[j+1])
	}
	return i
}

// Workdir sets the working directory
func (i *Image) Workdir(path string) *Image {
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("WORKDIR %s", path))
	return i
}

// Copy copies files or directories into the image
func (i *Image) Copy(src, dest string) *Image {
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("COPY %s %s", src, dest))
	return i
}

// Add adds files or URLs to the image (similar to COPY but with URL support)
func (i *Image) Add(src, dest string) *Image {
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("ADD %s %s", src, dest))
	return i
}

// File creates a file with the specified content in the image
func (i *Image) File(path, content string) *Image {
	// Escape the content for shell - replace single quotes and backslashes
	escapedContent := strings.ReplaceAll(content, "\\", "\\\\")
	escapedContent = strings.ReplaceAll(escapedContent, "'", "'\\''")
	
	// Create parent directory if needed and write file
	dir := ""
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		dir = path[:idx]
	}
	
	if dir != "" && dir != "/" {
		i.dockerfile = append(i.dockerfile, fmt.Sprintf("RUN mkdir -p %s", dir))
	}
	
	// Use echo with -e flag to handle escape sequences and write to file
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("RUN echo '%s' > %s", escapedContent, path))
	
	return i
}

// User sets the user for subsequent commands
func (i *Image) User(user string) *Image {
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("USER %s", user))
	return i
}

// Expose exposes a port
func (i *Image) Expose(port int) *Image {
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("EXPOSE %d", port))
	return i
}

// Label adds metadata to the image
func (i *Image) Label(key, value string) *Image {
	i.dockerfile = append(i.dockerfile, fmt.Sprintf("LABEL %s=\"%s\"", key, value))
	return i
}

// Entrypoint sets the entrypoint for the container
func (i *Image) Entrypoint(command []string) *Image {
	if len(command) > 0 {
		cmdStr := strings.Join(command, "\", \"")
		i.dockerfile = append(i.dockerfile, fmt.Sprintf("ENTRYPOINT [\"%s\"]", cmdStr))
	}
	return i
}

// Cmd sets the default command for the container
func (i *Image) Cmd(command []string) *Image {
	if len(command) > 0 {
		cmdStr := strings.Join(command, "\", \"")
		i.dockerfile = append(i.dockerfile, fmt.Sprintf("CMD [\"%s\"]", cmdStr))
	}
	return i
}

// Build returns the generated Dockerfile content
func (i *Image) Build() string {
	return strings.Join(i.dockerfile, "\n") + "\n"
}

// String returns the generated Dockerfile content (alias for Build)
func (i *Image) String() string {
	return i.Build()
}

// Common preset builders

// PythonDataScience creates an image with common data science packages
func PythonDataScience(pythonVersion string) *Image {
	return DebianSlim(pythonVersion).
		AptInstall([]string{"gcc", "g++", "gfortran"}).
		PipInstall([]string{
			"numpy",
			"pandas",
			"scipy",
			"scikit-learn",
			"matplotlib",
			"seaborn",
			"jupyter",
			"ipython",
		}).
		Workdir("/workspace")
}

// PythonWeb creates an image with common web development packages
func PythonWeb(pythonVersion string) *Image {
	return DebianSlim(pythonVersion).
		PipInstall([]string{
			"flask",
			"fastapi",
			"uvicorn",
			"requests",
			"sqlalchemy",
			"alembic",
			"redis",
			"celery",
		}).
		Workdir("/app").
		Expose(8000)
}

// NodeJS creates a Node.js image
func NodeJS(nodeVersion string) *Image {
	if nodeVersion == "" {
		nodeVersion = "20"
	}
	return Base(fmt.Sprintf("node:%s-slim", nodeVersion)).
		Workdir("/app")
}

// Go creates a Go development image
func Go(goVersion string) *Image {
	if goVersion == "" {
		goVersion = "1.21"
	}
	return Base(fmt.Sprintf("golang:%s", goVersion)).
		Workdir("/workspace")
}