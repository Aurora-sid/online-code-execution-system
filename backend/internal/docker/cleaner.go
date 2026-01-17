package docker

// Cleaner is responsible for cleaning up artifacts
// Since we are destroying containers in ReturnContainer, we just need to handling file system temp files if any.
// But we inject code via tar stream to container memory (or fs).
// If we execute in container, the data is in container.
// If we destroy container, data is gone.
// So Cleaner logic is trivial: ensure container is removed.
// Which is handled by Pool.ReturnContainer.

// We can just keep this file as a placeholder for future host-side cleanup
// e.g. if we saved temp files on host before injecting. 
// But Sandbox.Execute uses bytes buffer. No host files.
// So nothing to clean on host!

// We can implement a "Reaper" here that checks for zombie containers.


import (
    "context"
    "time"

    "github.com/docker/docker/api/types/filters"
)

func StartReaper(s *Sandbox) {
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for range ticker.C {
            // Find containers older than X and remove them
             pruneArgs := filters.NewArgs()
             pruneArgs.Add("until", "1h")
             s.cli.ContainersPrune(context.Background(), pruneArgs)
        }
    }()
}
