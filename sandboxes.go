package daytona

import (
	"context"
	"fmt"
	"time"

	api "github.com/PhilippBuschhaus/daytona-sdk-go/apiclient"
)

// CreateSandboxRequest wraps the API CreateSandbox model with helper methods
type CreateSandboxRequest struct {
	Snapshot            *string
	User                *string
	Env                 map[string]string
	Labels              map[string]string
	Public              *bool
	Class               *string
	Target              *string
	CPU                 *int32
	GPU                 *int32
	Memory              *int32
	Disk                *int32
	AutoStopInterval    *int32
	AutoArchiveInterval *int32
	AutoDeleteInterval  *int32
	DockerfileContent   *string // For custom Dockerfile builds
}

// ToAPI converts to the API model
func (r *CreateSandboxRequest) ToAPI() api.CreateSandbox {
	req := api.CreateSandbox{}

	if r.Snapshot != nil {
		req.Snapshot = r.Snapshot
	}
	if r.User != nil {
		req.User = r.User
	}
	if r.Env != nil {
		req.Env = &r.Env
	}
	if r.Labels != nil {
		req.Labels = &r.Labels
	}
	if r.Public != nil {
		req.Public = r.Public
	}
	if r.Class != nil {
		req.Class = r.Class
	}
	if r.Target != nil {
		req.Target = r.Target
	}
	if r.CPU != nil {
		req.Cpu = r.CPU
	}
	if r.GPU != nil {
		req.Gpu = r.GPU
	}
	if r.Memory != nil {
		req.Memory = r.Memory
	}
	if r.Disk != nil {
		req.Disk = r.Disk
	}
	if r.AutoStopInterval != nil {
		req.AutoStopInterval = r.AutoStopInterval
	}
	if r.AutoArchiveInterval != nil {
		req.AutoArchiveInterval = r.AutoArchiveInterval
	}
	if r.AutoDeleteInterval != nil {
		req.AutoDeleteInterval = r.AutoDeleteInterval
	}
	if r.DockerfileContent != nil {
		buildInfo := api.NewCreateBuildInfo(*r.DockerfileContent)
		req.BuildInfo = buildInfo
	}

	return req
}

// CreateSandbox creates a new sandbox
func (c *Client) CreateSandbox(ctx context.Context, req *CreateSandboxRequest) (*api.Sandbox, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	apiReq := req.ToAPI()

	resp, _, err := c.apiClient.SandboxAPI.CreateSandbox(ctx).CreateSandbox(apiReq).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}

	return resp, nil
}

// GetSandbox retrieves a sandbox by ID
func (c *Client) GetSandbox(ctx context.Context, sandboxID string) (*api.Sandbox, error) {
	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID is required")
	}

	resp, _, err := c.apiClient.SandboxAPI.GetSandbox(ctx, sandboxID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox: %w", err)
	}

	return resp, nil
}

// ListSandboxes lists all sandboxes
func (c *Client) ListSandboxes(ctx context.Context) ([]api.Sandbox, error) {
	resp, _, err := c.apiClient.SandboxAPI.ListSandboxes(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list sandboxes: %w", err)
	}

	return resp, nil
}

// StartSandbox starts a stopped sandbox
func (c *Client) StartSandbox(ctx context.Context, sandboxID string) (*api.Sandbox, error) {
	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID is required")
	}

	resp, _, err := c.apiClient.SandboxAPI.StartSandbox(ctx, sandboxID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to start sandbox: %w", err)
	}

	return resp, nil
}

// StopSandbox stops a running sandbox
func (c *Client) StopSandbox(ctx context.Context, sandboxID string) error {
	if sandboxID == "" {
		return fmt.Errorf("sandbox ID is required")
	}

	_, _, err := c.apiClient.SandboxAPI.StopSandbox(ctx, sandboxID).Execute()
	if err != nil {
		return fmt.Errorf("failed to stop sandbox: %w", err)
	}

	return nil
}

// DeleteSandbox deletes a sandbox
func (c *Client) DeleteSandbox(ctx context.Context, sandboxID string) error {
	if sandboxID == "" {
		return fmt.Errorf("sandbox ID is required")
	}

	req := c.apiClient.SandboxAPI.DeleteSandbox(ctx, sandboxID)

	_, _, err := req.Execute()
	if err != nil {
		return fmt.Errorf("failed to delete sandbox: %w", err)
	}

	return nil
}

// ArchiveSandbox archives a sandbox
func (c *Client) ArchiveSandbox(ctx context.Context, sandboxID string) error {
	if sandboxID == "" {
		return fmt.Errorf("sandbox ID is required")
	}

	_, _, err := c.apiClient.SandboxAPI.ArchiveSandbox(ctx, sandboxID).Execute()
	if err != nil {
		return fmt.Errorf("failed to archive sandbox: %w", err)
	}

	return nil
}

// WaitForSandboxReady waits for a sandbox to be in running state
func (c *Client) WaitForSandboxReady(ctx context.Context, sandboxID string, timeout time.Duration) (*api.Sandbox, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for sandbox to be ready")
			}

			sandbox, err := c.GetSandbox(ctx, sandboxID)
			if err != nil {
				return nil, err
			}

			if sandbox.State != nil {
				state := *sandbox.State
				if state == api.SANDBOXSTATE_STARTED {
					return sandbox, nil
				}
				if state == api.SANDBOXSTATE_DESTROYED || state == api.SANDBOXSTATE_ERROR || state == api.SANDBOXSTATE_BUILD_FAILED {
					return nil, fmt.Errorf("sandbox failed to start: %s", state)
				}
			}
		}
	}
}

// Helper functions for creating pointers
func StringPtr(s string) *string { return &s }
func IntPtr(i int32) *int32      { return &i }
func BoolPtr(b bool) *bool       { return &b }
