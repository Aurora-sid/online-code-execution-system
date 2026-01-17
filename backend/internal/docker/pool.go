package docker

import (
	"context"
	"fmt"
	"sync"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type Pool struct {
	Sandbox *Sandbox
	// Map of language -> list of container IDs
	Available map[string][]string
	mu        sync.Mutex
}

func NewPool(s *Sandbox) *Pool {
	return &Pool{
		Sandbox:   s,
		Available: make(map[string][]string),
	}
}

// GetContainer returns a container ID for a language
func (p *Pool) GetContainer(ctx context.Context, language string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 1. Check if available
	if ids, ok := p.Available[language]; ok && len(ids) > 0 {
		id := ids[0]
		p.Available[language] = ids[1:]
		// Unpause if paused
		return id, nil
	}

	// 2. Create New
	return p.createContainer(ctx, language)
}

func (p *Pool) createContainer(ctx context.Context, language string) (string, error) {
	imageName := fmt.Sprintf("code-exec/%s", language)
	
	resp, err := p.Sandbox.cli.ContainerCreate(ctx, 
		&container.Config{
			Image: imageName,
			Cmd:   []string{"sleep", "infinity"},
			Tty:   false,
		}, 
		&container.HostConfig{
			Resources: container.Resources{
				Memory:   512 * 1024 * 1024, // 512MB (increased for Go compiler)
				NanoCPUs: 500000000,         // 0.5 CPU
			},
			NetworkMode: "none", // Security: No network
		}, 
		nil, nil, "")
	
	if err != nil {
		return "", err
	}

	if err := p.Sandbox.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

// ReturnContainer puts a container back or destroys it
func (p *Pool) ReturnContainer(ctx context.Context, containerID string) {
	// For MVP: Destroy it to ensure cleanliness
	// In future: p.Available[lang] = append(..., containerID) after cleanup
	p.Sandbox.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
}
