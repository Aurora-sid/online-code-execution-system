package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
	"unicode"

	"code-exec/config"
	"encoding/base64"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

// PrepareCode 创建要注入的代码的 tar 归档
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

// ExecuteInteractive 运行带有交互式标准输入支持的代码
// stdinChan 接收用户输入，outputChan 发送输出给用户
func (s *Sandbox) ExecuteInteractive(ctx context.Context, containerID string, language string, code string,
	stdinChan <-chan string, outputChan chan<- string) (int, error) {

	langCfg, err := config.GetLanguageConfig(language)
	if err != nil {
		return 0, err
	}

	filename := langCfg.Filename
	cmd := langCfg.RunCmd
	if len(langCfg.CompileCmd) > 0 {
		// 如果有编译命令，这里只取运行命令。注意：实际执行逻辑可能需要调整。
		// 在当前架构下，ExecuteInteractive 似乎假设是解释型语言或已处理好编译?
		// 之前的 getLanguageConfig 中：
		// java: "sh", "-c", "javac Main.java && java Main"
		// cpp: "sh", "-c", "g++ main.cpp -o main && ./main"
		// 这里 compile_cmd 好像被包含在了 run 逻辑里了（对于 sandbox 来说）。
		// 我们在 yaml 中定义了 run_cmd 和 compile_cmd。
		// 如果是单步执行（像旧代码那样把编译和运行串起来），我们应该检查 yaml 怎么定义的。
		// 旧代码 Java: "sh", "-c", "javac ... && java ..."
		// 我们的 yaml Java: compile_cmd: ["sh", "-c", "javac ... && java ..."]
		// 实际上旧代码把这一长串当做 cmd。
		// 所以如果 compile_cmd 存在，我们就用 compile_cmd（因为它包含了编译&&运行）。
		cmd = langCfg.CompileCmd
	}
	if len(cmd) == 0 {
		cmd = langCfg.RunCmd
	}

	// 注入代码
	log.Printf("[Sandbox] 正在注入代码到容器 %s (文件: %s, 大小: %d 字节)\n", containerID[:12], filename, len(code))
	tarBuf, err := PrepareCode(filename, code)
	if err != nil {
		log.Printf("[Sandbox] 准备代码失败: %v\n", err)
		return 0, err
	}

	err = s.cli.CopyToContainer(ctx, containerID, "/app", tarBuf, types.CopyToContainerOptions{})
	if err != nil {
		return 0, fmt.Errorf("inject error: %v", err)
	}

	// 创建执行并附加标准输入
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          false,
		WorkingDir:   "/app",
	}

	resp, err := s.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		log.Printf("[Sandbox] 创建执行实例失败: %v\n", err)
		return 0, fmt.Errorf("exec create error: %v", err)
	}
	log.Printf("[Sandbox] 执行实例已创建 %s (CMD: %v)\n", resp.ID[:12], cmd)

	// 开始执行并附加
	attachResp, err := s.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		log.Printf("[Sandbox] 附加执行实例失败: %v\n", err)
		return 0, fmt.Errorf("exec attach error: %v", err)
	}
	log.Printf("[Sandbox] 开始执行代码...\n")
	defer attachResp.Close()

	// Goroutine 读取标准输出并发送到 outputChan
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		// lastOutputTime := time.Now()
		// waitingInputSent := false

		for {
			// 设置读取超时
			attachResp.Conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

			n, err := attachResp.Reader.Read(buf)
			if n > 0 {
				// lastOutputTime = time.Now()
				// waitingInputSent = false // 有新输出，重置标记

				// 跳过 Docker 多路复用头部 (8 bytes)
				data := buf[:n]
				if len(data) > 8 && (data[0] == 1 || data[0] == 2) {
					data = data[8:]
				}
				output := sanitizeOutput(string(data))
				// 扫描 GUI 错误提示
				if guiTip := scanForGuiErrors(output); guiTip != "" {
					output += guiTip
				}

				if output != "" {
					select {
					case outputChan <- output:
					case <-ctx.Done():
						return
					}
				}
			}

			if err != nil {
				// 检查是否是超时错误
				if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
					// 超时了，检查是否应该发送"等待输入"信号
					/*
					   禁用自动发送等待输入信号，因为无法准确区分"编译/运行慢"和"真的再等待输入"。
					   这会导致 Go/Java 等需要编译的语言在启动时误弹出输入框。
					   改为完全依赖前端的手动"输入"按钮。
					*/
					// if !waitingInputSent && time.Since(lastOutputTime) > 3000*time.Millisecond {
					// 	// 发送等待输入信号
					// 	select {
					// 	case outputChan <- "\n__WAITING_INPUT__":
					// 		waitingInputSent = true
					// 	case <-ctx.Done():
					// 		return
					// 	}
					// }
					continue // 继续读取
				}

				if err != io.EOF {
					select {
					case outputChan <- fmt.Sprintf("\nError reading output: %v", err):
					case <-ctx.Done():
					}
				}
				return
			}
		}
	}()

	// Goroutine 从通道写入标准输入
	go func() {
		for {
			select {
			case input, ok := <-stdinChan:
				if !ok {
					// 通道已关闭，关闭标准输入
					attachResp.CloseWrite()
					return
				}
				// 写入输入到进程
				_, err := attachResp.Conn.Write([]byte(input))
				if err != nil {
					return
				}
			case <-ctx.Done():
				attachResp.CloseWrite()
				return
			case <-done:
				return
			}
		}
	}()

	// 等待完成或超时
	select {
	case <-done:
		// 进程已完成
	case <-ctx.Done():
		// 超时
		log.Printf("[Sandbox] 执行超时，正在终止...\n")
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		errorMsg := s.diagnoseTimeout(stopCtx, containerID)
		s.cli.ContainerStop(stopCtx, containerID, container.StopOptions{})
		log.Printf("[Sandbox] 容器已停止 (原因: %s)\n", errorMsg)
		return 124, fmt.Errorf("%s", errorMsg)
	}

	// 获取退出码
	inspectResp, err := s.cli.ContainerExecInspect(context.Background(), resp.ID)
	if err != nil {
		log.Printf("[Sandbox] 获取退出码失败: %v\n", err)
		return 0, fmt.Errorf("inspect error: %v", err)
	}

	// 检查是否有生成的图片
	if imgMsg := s.checkForImages(context.Background(), resp.ID); imgMsg != "" {
		select {
		case outputChan <- imgMsg:
		case <-ctx.Done():
		}
	}

	log.Printf("[Sandbox] 执行完成 (退出码: %d)\n", inspectResp.ExitCode)
	return inspectResp.ExitCode, nil
}

// Execute 在指定容器中运行代码（非交互式，用于向后兼容）
func (s *Sandbox) Execute(ctx context.Context, containerID string, language string, code string, stdin string) (string, int, error) {
	langCfg, err := config.GetLanguageConfig(language)
	if err != nil {
		return "", 0, err
	}

	filename := langCfg.Filename
	cmd := langCfg.RunCmd
	if len(langCfg.CompileCmd) > 0 {
		cmd = langCfg.CompileCmd
	}
	if len(cmd) == 0 {
		cmd = langCfg.RunCmd
	}

	// 注入代码
	tarBuf, err := PrepareCode(filename, code)
	if err != nil {
		return "", 0, err
	}

	err = s.cli.CopyToContainer(ctx, containerID, "/app", tarBuf, types.CopyToContainerOptions{})
	if err != nil {
		return "", 0, fmt.Errorf("inject error: %v", err)
	}

	// 创建执行
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  stdin != "",
		WorkingDir:   "/app",
	}

	resp, err := s.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", 0, fmt.Errorf("exec create error: %v", err)
	}

	// 开始执行
	attachResp, err := s.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return "", 0, fmt.Errorf("exec attach error: %v", err)
	}
	defer attachResp.Close()

	// 如果提供了标准输入则写入
	outputChan := make(chan string)
	errChan := make(chan error)

	go func() {
		if stdin != "" {
			_, err := attachResp.Conn.Write([]byte(stdin + "\n"))
			if err != nil {
				errChan <- fmt.Errorf("stdin write error: %v", err)
				return
			}
			attachResp.CloseWrite()
		}

		var outBuf bytes.Buffer
		_, err := stdcopy.StdCopy(&outBuf, &outBuf, attachResp.Reader)
		if err != nil && err != io.EOF {
			errChan <- err
			return
		}
		outputChan <- outBuf.String()
	}()

	select {
	case out := <-outputChan:
		inspectResp, err := s.cli.ContainerExecInspect(ctx, resp.ID)
		if err != nil {
			return sanitizeOutput(out), 0, fmt.Errorf("inspect error: %v", err)
		}

		sanitizedOut := sanitizeOutput(out)
		if leakMsg := scanSensitiveData(sanitizedOut); leakMsg != "" {
			return leakMsg, 1, nil
		}

		return sanitizedOut, inspectResp.ExitCode, nil
	case err := <-errChan:
		return "", 0, err
	case <-ctx.Done():
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		errorMsg := s.diagnoseTimeout(stopCtx, containerID)
		s.cli.ContainerStop(stopCtx, containerID, container.StopOptions{})
		return "", 124, fmt.Errorf("%s", errorMsg)
	}
}

// diagnoseTimeout 检查容器状态和统计信息以确定超时原因
func (s *Sandbox) diagnoseTimeout(ctx context.Context, containerID string) string {
	stats, err := s.cli.ContainerStats(ctx, containerID, false)
	if err == nil {
		defer stats.Body.Close()

		var v types.StatsJSON
		if err := json.NewDecoder(stats.Body).Decode(&v); err == nil {
			if v.PidsStats.Current >= 45 {
				return "执行超时：进程数量增长异常（Fork炸弹特征），系统已拦截"
			}

			if v.MemoryStats.Usage > 0 && v.MemoryStats.Limit > 0 {
				usagePercent := float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100
				if usagePercent > 90 {
					return fmt.Sprintf("执行超时：内存占用激增（%.1f%%），判定为内存炸弹攻击", usagePercent)
				}
			}

			if v.MemoryStats.Failcnt > 0 {
				return "执行超时：物理内存耗尽（OOM Killed），系统自动终止进程"
			}
		}
	}

	inspect, err := s.cli.ContainerInspect(ctx, containerID)
	if err == nil {
		if inspect.State.OOMKilled {
			return "执行超时：内存耗尽被系统终止(OOM Killed)"
		}
	}

	return "执行超时：代码运行时间超过25秒限制"
}

// scanSensitiveData 检查输出是否包含敏感系统文件内容
func scanSensitiveData(output string) string {
	if strings.Contains(output, "root:x:0:0:") || strings.Contains(output, "daemon:x:") {
		return "安全警告：检测到敏感系统文件(/etc/passwd)读取尝试，输出已被拦截"
	}
	if strings.Contains(output, ":$6$") || strings.Contains(output, ":$5$") {
		return "安全警告：检测到敏感系统文件(/etc/shadow)读取尝试，输出已被拦截"
	}
	// 扫描 GUI 相关错误并给出提示
	if guiTip := scanForGuiErrors(output); guiTip != "" {
		return output + guiTip
	}
	return ""
}

// scanForGuiErrors 扫描输出中的 GUI 相关错误并返回提示
func scanForGuiErrors(output string) string {
	keywords := []string{
		"TclError: no display name",
		"Gtk-WARNING",
		"qt.qpa.xcb",
		"UserWarning: Matplotlib is currently using agg",
		"unable to open display",
		"cannot connect to X server",
	}

	for _, kw := range keywords {
		if strings.Contains(output, kw) {
			return "\n\n[System Tip] 检测到您尝试显示图形界面 (GUI) 或使用 plt.show()。\n" +
				"本系统运行在服务器端无头环境 (Headless)，不支持弹窗显示。\n" +
				"请将图片保存为文件，例如：plt.savefig('output.png')，系统会自动为您展示保存的图片。"
		}
	}
	return ""
}

// checkForImages 检查容器内是否生成了图片文件，如果有则读取并转换为 Base64
func (s *Sandbox) checkForImages(ctx context.Context, containerID string) string {
	// 简单起见，目前只检查 output.png
	// 将文件 cat 出来
	execConfig := types.ExecConfig{
		Cmd:          []string{"cat", "/app/output.png"},
		AttachStdout: true,
		AttachStderr: false,
	}

	resp, err := s.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return ""
	}

	attachResp, err := s.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return ""
	}
	defer attachResp.Close()

	var outBuf bytes.Buffer
	// 读取输出（注意处理 Docker header）
	// 由于 stdcopy 需要 demultiplex，我们简单读取所有字节
	// 由于我们用的是 ContainerExecAttach，它返回的是原始流（带 Header）
	// 为了简单，我们使用 stdcopy 来剥离 stdout
	_, err = stdcopy.StdCopy(&outBuf, io.Discard, attachResp.Reader)
	if err != nil {
		return ""
	}

	// 检查是否有内容且没有"No such file" (cat 错误通常在 stderr，但我们没attach stderr，所以 stdout 应该是纯文件内容)
	// 如果文件不存在，cat 可能会从 stdout 输出吗？通常是在 stderr。
	// 检查 exit code 可能更稳妥，但 Create/Attach 流程不容易同步获取 ExitCode。
	// 我们假设如果 Content-Length > 0 且开头看起来像 PNG 文件头，就是图片。
	data := outBuf.Bytes()
	if len(data) == 0 {
		return ""
	}

	// 简单的 PNG 头检查 (89 50 4E 47 0D 0A 1A 0A)
	if len(data) > 8 && string(data[:4]) == "\x89PNG" {
		log.Printf("[Sandbox] 检测到生成图片 output.png (大小: %d 字节)\n", len(data))
		encoded := base64.StdEncoding.EncodeToString(data)
		return fmt.Sprintf("\n\n[System] 检测到图片生成 (output.png):\n<<<<IMAGE_START>>>>%s<<<<IMAGE_END>>>>\n", encoded)
	}

	return ""
}

// sanitizeOutput 从输出中移除不可打印的控制字符
func sanitizeOutput(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r == '\r' {
			return r
		}
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}
