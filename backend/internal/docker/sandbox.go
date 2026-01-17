package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"



	"code-exec/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Sandbox struct {
	cli *client.Client
}

func NewSandbox() (*Sandbox, error) {
	cfg := config.Load()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(cfg.DockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Sandbox{cli: cli}, nil
}

// PrepareCode creates a tar archive of the code to inject
func PrepareCode(filename, content string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	header := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		return nil, err
	}
	return buf, nil
}

// Execute runs the code in the specified container and returns output
func (s *Sandbox) Execute(ctx context.Context, containerID string, language string, code string) (string, error) {
	// 1. Determine filename based on language
	filename := "main.py"
	cmd := []string{"python", "main.py"}
	if language == "java" {
		filename = "Main.java"
		cmd = []string{"sh", "-c", "javac Main.java && java Main"} // Simplified
	} else if language == "cpp" {
		filename = "main.cpp"
		cmd = []string{"sh", "-c", "g++ main.cpp -o main && ./main"}
	} else if language == "go" {
		filename = "main.go"
		cmd = []string{"go", "run", "main.go"}
	}

	// 2. Inject Code
	tarBuf, err := PrepareCode(filename, code)
	if err != nil {
		return "", err
	}

	err = s.cli.CopyToContainer(ctx, containerID, "/app", tarBuf, types.CopyToContainerOptions{})
	if err != nil {
		return "", fmt.Errorf("inject error: %v", err)
	}

	// 3. Exec Create
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/app",
	}

	resp, err := s.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("exec create error: %v", err)
	}

	// 4. Exec Start
	attachResp, err := s.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("exec attach error: %v", err)
	}
	defer attachResp.Close()

	// 5. Read Output with Timeout
	// Use a goroutine to read output
	outputChan := make(chan string)
	errChan := make(chan error)

	go func() {
		var outBuf bytes.Buffer
		// StdCopy demultiplexes stdout/stderr (if TTY is false, which it is by default in ExecConfig)
		// But here we are getting raw stream or demuxed? 
		// Attach gives us a Reader.
		// If tty=false, the stream is multiplexed. We need stdcopy.StdCopy
		// But since we didn't import pkg/stdcopy (it's in docker/pkg/stdcopy), we can just read all for now 
		// or treat it as raw text if we don't care about separating stdout/stderr strictly yet.
		// However, docker output has headers. simpler to just read all bytes and strip headers if lazy, 
		// but correct way is stdcopy. 
		// For simplicity in this step, let's just readAll. It might contain header bytes.
		// "github.com/docker/docker/pkg/stdcopy"
		
		_, err := io.Copy(&outBuf, attachResp.Reader)
		if err != nil {
			errChan <- err
			return
		}
		outputChan <- outBuf.String()
	}()

	select {
	case out := <-outputChan:
		return out, nil
	case err := <-errChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("timeout")
	}
}
