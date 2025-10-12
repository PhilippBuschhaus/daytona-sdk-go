package daytona

import (
	"context"
	"fmt"

	api "github.com/PhilippBuschhaus/daytona-sdk-go/apiclient"
)

// ExecuteCommandRequest wraps command execution parameters
type ExecuteCommandRequest struct {
	Command string
	Cwd     string
	Timeout float32
}

// ExecuteCommand executes a command in a sandbox
func (c *Client) ExecuteCommand(ctx context.Context, sandboxID string, req *ExecuteCommandRequest) (*api.ExecuteResponse, error) {
	if sandboxID == "" {
		return nil, fmt.Errorf("sandbox ID is required")
	}
	if req == nil || req.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Create the API request
	execReq := api.ExecuteRequest{
		Command: req.Command,
	}

	if req.Cwd != "" {
		execReq.Cwd = &req.Cwd
	}
	if req.Timeout > 0 {
		execReq.Timeout = &req.Timeout
	}

	// Execute the command
	resp, _, err := c.apiClient.ToolboxAPI.ExecuteCommand(ctx, sandboxID).ExecuteRequest(execReq).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return resp, nil
}

// CreateSession creates a new session for interactive commands
func (c *Client) CreateSession(ctx context.Context, sandboxID string, sessionID string) error {
	if sandboxID == "" || sessionID == "" {
		return fmt.Errorf("sandbox ID and session ID are required")
	}

	req := api.CreateSessionRequest{
		SessionId: sessionID,
	}

	_, err := c.apiClient.ToolboxAPI.CreateSession(ctx, sandboxID).CreateSessionRequest(req).Execute()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// ExecuteSessionCommand executes a command in an existing session
func (c *Client) ExecuteSessionCommand(ctx context.Context, sandboxID, sessionID, command string) (*api.SessionExecuteResponse, error) {
	if sandboxID == "" || sessionID == "" {
		return nil, fmt.Errorf("sandbox ID and session ID are required")
	}
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}

	req := api.SessionExecuteRequest{
		Command: command,
	}

	resp, _, err := c.apiClient.ToolboxAPI.ExecuteSessionCommand(ctx, sandboxID, sessionID).SessionExecuteRequest(req).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to execute session command: %w", err)
	}

	return resp, nil
}

// DeleteSession deletes a session
func (c *Client) DeleteSession(ctx context.Context, sandboxID, sessionID string) error {
	if sandboxID == "" || sessionID == "" {
		return fmt.Errorf("sandbox ID and session ID are required")
	}

	_, err := c.apiClient.ToolboxAPI.DeleteSession(ctx, sandboxID, sessionID).Execute()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
