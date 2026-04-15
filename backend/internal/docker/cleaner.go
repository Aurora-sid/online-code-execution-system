package docker

// Cleaner 负责清理工件
// 由于我们在 ReturnContainer 中销毁容器，我们只需要处理文件系统临时文件（如果有）。
// 但我们通过 tar 流将代码注入到容器内存（或文件系统）中。
// 如果我们在容器中执行，数据就在容器中。
// 如果我们销毁容器，数据就消失了。
// 所以 Cleaner 的逻辑很简单：确保容器被移除。
// 这由 Pool.ReturnContainer 处理。

// 我们可以保留此文件作为未来宿主机侧清理的占位符
// 例如，如果我们注入前在宿主机保存了临时文件。
// 但 Sandbox.Execute 使用字节缓冲区。没有宿主机文件。
// 所以宿主机上没有什么需要清理的！

// 我们可以在这里实现一个 "Reaper" 来检查僵尸容器。
// 避免随着项目的长期在线或者代码沙箱未妥善回收导致的主机硬盘或者节点计算资源干涸。

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/filters"
)

func StartReaper(s *Sandbox) {
	// 每小时检查一次容器
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			// 查找早于 X 的容器并将其移除
			pruneArgs := filters.NewArgs()
			pruneArgs.Add("until", "1h")
			s.cli.ContainersPrune(context.Background(), pruneArgs)
		}
	}()
}
