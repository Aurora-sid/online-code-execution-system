package docker

import (
	"encoding/json"
	"log"
)

// ============================================================
//  Seccomp 白名单安全配置
//  基于 OCI 运行时规范，仅放行代码编译/运行所需的系统调用
//  未列入白名单的系统调用将返回 EPERM (Operation not permitted)
// ============================================================

// seccompProfile 定义 Seccomp Profile 的 JSON 结构
type seccompProfile struct {
	DefaultAction string         `json:"defaultAction"`
	Architectures []string       `json:"architectures"`
	Syscalls      []syscallGroup `json:"syscalls"`
}

type syscallGroup struct {
	Names  []string `json:"names"`
	Action string   `json:"action"`
}

// cachedSeccompJSON 缓存序列化后的 JSON（只生成一次）
var cachedSeccompJSON string

// getSeccompProfile 返回 Seccomp Profile 的 JSON 字符串
// 采用白名单模式：默认拒绝，仅放行约 110 个必要系统调用
func getSeccompProfile() string {
	if cachedSeccompJSON != "" {
		return cachedSeccompJSON
	}

	profile := seccompProfile{
		// 默认行为：拒绝（返回 EPERM）
		DefaultAction: "SCMP_ACT_ERRNO",
		// 支持的 CPU 架构
		Architectures: []string{
			"SCMP_ARCH_X86_64",
			"SCMP_ARCH_X86",
			"SCMP_ARCH_AARCH64",
		},
		Syscalls: []syscallGroup{
			{
				Action: "SCMP_ACT_ALLOW",
				Names: []string{
					// ============================================
					// 1. 进程管理（编译器/解释器必需）
					// ============================================
					"execve",       // 执行新程序（编译器、解释器启动）
					"execveat",     // execve 的扩展形式
					"clone",        // 创建子进程/线程
					"clone3",       // clone 的新版本
					"fork",         // 创建子进程
					"vfork",        // 创建子进程（共享内存）
					"wait4",        // 等待子进程结束
					"waitid",       // 等待子进程（扩展）
					"exit",         // 进程退出
					"exit_group",   // 线程组退出
					"getpid",       // 获取进程ID
					"getppid",      // 获取父进程ID
					"gettid",       // 获取线程ID
					"getuid",       // 获取用户ID
					"geteuid",      // 获取有效用户ID
					"getgid",       // 获取组ID
					"getegid",      // 获取有效组ID
					"getgroups",    // 获取附加组列表
					"setuid",       // 设置用户ID（容器内降权）
					"setgid",       // 设置组ID
					"setgroups",    // 设置附加组
					"setsid",       // 创建新会话
					"getpgrp",      // 获取进程组
					"setpgid",      // 设置进程组
					"getpgid",      // 获取进程组ID
					"getsid",       // 获取会话ID
					"prctl",        // 进程控制（运行时需要）
					"arch_prctl",   // 架构相关进程控制
					"set_tid_address", // 线程ID地址设置
					"set_robust_list", // 健壮互斥锁列表

					// ============================================
					// 2. 文件 I/O（读写源文件和编译产物）
					// ============================================
					"read",         // 读文件
					"write",        // 写文件
					"open",         // 打开文件
					"openat",       // 打开文件（相对路径）
					"openat2",      // 打开文件（新版）
					"close",        // 关闭文件
					"stat",         // 获取文件状态
					"fstat",        // 获取文件状态（fd）
					"lstat",        // 获取链接文件状态
					"newfstatat",   // 获取文件状态（新版）
					"statx",        // 获取文件扩展状态
					"lseek",        // 移动文件指针
					"pread64",      // 定位读取
					"pwrite64",     // 定位写入
					"readv",        // 聚集读取
					"writev",       // 分散写入
					"preadv",       // 定位聚集读取
					"pwritev",      // 定位分散写入
					"preadv2",      // 定位聚集读取v2
					"pwritev2",     // 定位分散写入v2
					"access",       // 检查文件权限
					"faccessat",    // 检查文件权限（相对路径）
					"faccessat2",   // 检查文件权限（新版）
					"dup",          // 复制文件描述符
					"dup2",         // 复制文件描述符到指定fd
					"dup3",         // 复制文件描述符（带flags）
					"pipe",         // 创建管道
					"pipe2",        // 创建管道（带flags）
					"select",       // I/O多路复用
					"pselect6",     // I/O多路复用（新版）
					"poll",         // I/O多路复用
					"ppoll",        // I/O多路复用（新版）
					"epoll_create", // epoll 创建
					"epoll_create1", // epoll 创建（带flags）
					"epoll_ctl",    // epoll 控制
					"epoll_wait",   // epoll 等待
					"epoll_pwait",  // epoll 等待（带信号掩码）
					"epoll_pwait2", // epoll 等待v2
					"eventfd",      // 事件通知
					"eventfd2",     // 事件通知（带flags）
					"readlink",     // 读取符号链接
					"readlinkat",   // 读取符号链接（相对路径）
					"fcntl",        // 文件控制
					"ioctl",        // 设备控制
					"flock",        // 文件锁定
					"fsync",        // 同步文件数据
					"fdatasync",    // 同步文件数据（不含元数据）
					"truncate",     // 截断文件
					"ftruncate",    // 截断文件（fd）
					"rename",       // 重命名文件
					"renameat",     // 重命名文件（相对路径）
					"renameat2",    // 重命名文件（新版）
					"unlink",       // 删除文件
					"unlinkat",     // 删除文件（相对路径）
					"symlink",      // 创建符号链接
					"symlinkat",    // 创建符号链接（相对路径）
					"link",         // 创建硬链接
					"linkat",       // 创建硬链接（相对路径）
					"chmod",        // 修改文件权限
					"fchmod",       // 修改文件权限（fd）
					"fchmodat",     // 修改文件权限（相对路径）
					"chown",        // 修改文件所有者
					"fchown",       // 修改文件所有者（fd）
					"fchownat",     // 修改文件所有者（相对路径）
					"umask",        // 设置文件创建掩码

					// ============================================
					// 3. 目录操作
					// ============================================
					"mkdir",        // 创建目录
					"mkdirat",      // 创建目录（相对路径）
					"rmdir",        // 删除目录
					"getcwd",       // 获取当前目录
					"chdir",        // 切换目录
					"fchdir",       // 切换目录（fd）
					"getdents",     // 读取目录项
					"getdents64",   // 读取目录项（64位）

					// ============================================
					// 4. 内存管理（运行时必需）
					// ============================================
					"mmap",         // 内存映射
					"munmap",       // 取消内存映射
					"mprotect",     // 修改内存保护属性
					"mremap",       // 重新映射内存
					"madvise",      // 内存使用建议
					"brk",          // 调整数据段大小
					"msync",        // 同步内存映射
					"mincore",      // 检查页面驻留情况
					"mlock",        // 锁定内存
					"mlock2",       // 锁定内存（新版）
					"munlock",      // 解锁内存
					"mlockall",     // 锁定所有内存
					"munlockall",   // 解锁所有内存
					"membarrier",   // 内存屏障

					// ============================================
					// 5. 信号处理
					// ============================================
					"rt_sigaction",    // 设置信号处理
					"rt_sigprocmask",  // 修改信号掩码
					"rt_sigreturn",    // 信号返回
					"rt_sigsuspend",   // 等待信号
					"rt_sigpending",   // 获取挂起信号
					"rt_sigtimedwait", // 定时等待信号
					"rt_sigqueueinfo", // 排队信号信息
					"rt_tgsigqueueinfo", // 线程组排队信号
					"sigaltstack",     // 替代信号栈
					"kill",            // 发送信号
					"tgkill",          // 向线程发送信号
					"tkill",           // 向线程发送信号（旧版）

					// ============================================
					// 6. 时间与定时器
					// ============================================
					"clock_gettime",   // 获取时钟时间
					"clock_getres",    // 获取时钟精度
					"clock_nanosleep", // 高精度睡眠
					"gettimeofday",    // 获取当前时间
					"nanosleep",       // 纳秒级睡眠
					"timer_create",    // 创建定时器
					"timer_settime",   // 设置定时器
					"timer_gettime",   // 获取定时器
					"timer_getoverrun", // 获取定时器溢出次数
					"timer_delete",    // 删除定时器
					"timerfd_create",  // 创建定时器fd
					"timerfd_settime", // 设置定时器fd
					"timerfd_gettime", // 获取定时器fd
					"alarm",           // 设置闹钟
					"times",           // 获取进程时间

					// ============================================
					// 7. 线程同步（多线程程序必需）
					// ============================================
					"futex",           // 快速用户空间互斥锁
					"get_robust_list", // 获取健壮互斥锁列表
					"sched_yield",     // 让出CPU
					"sched_getaffinity",  // 获取CPU亲和性
					"sched_setaffinity",  // 设置CPU亲和性
					"sched_getscheduler", // 获取调度策略
					"sched_getparam",     // 获取调度参数
					"sched_setscheduler", // 设置调度策略
					"sched_setparam",     // 设置调度参数

					// ============================================
					// 8. 系统信息查询（只读，安全）
					// ============================================
					"uname",        // 获取系统信息
					"getrlimit",    // 获取资源限制
					"setrlimit",    // 设置资源限制（容器内）
					"prlimit64",    // 获取/设置资源限制（新版）
					"getrusage",    // 获取资源使用统计
					"sysinfo",      // 获取系统信息
					"getrandom",    // 获取随机数

					// ============================================
					// 9. 其他必需调用
					// ============================================
					"sendfile",     // 文件间数据传输（编译器用）
					"copy_file_range", // 文件范围复制
					"splice",       // 管道间数据传输
					"tee",          // 管道数据复制
					"vmsplice",     // 向管道写入用户空间数据
					"signalfd",     // 信号文件描述符
					"signalfd4",    // 信号文件描述符（新版）
					"inotify_init",  // 文件监控初始化
					"inotify_init1", // 文件监控初始化（新版）
					"inotify_add_watch", // 添加监控
					"inotify_rm_watch",  // 移除监控
					"capget",       // 获取进程能力
					"capset",       // 设置进程能力
					"rseq",         // 可重启序列（Go运行时需要）
					"close_range",  // 批量关闭文件描述符
					"statfs",       // 获取文件系统状态
					"fstatfs",      // 获取文件系统状态（fd）

					// ============================================
					// 10. 容器运行基础设施
					// ============================================
					"seccomp",          // Seccomp 自身操作
					"personality",      // 设置执行域（某些语言需要）
					"syslog",           // 系统日志（受限）
				},
			},
		},
	}

	data, err := json.Marshal(profile)
	if err != nil {
		log.Printf("[Seccomp] Profile 序列化失败: %v，将使用 Docker 默认策略\n", err)
		return ""
	}

	cachedSeccompJSON = string(data)
	log.Printf("[Seccomp] 白名单 Profile 已加载 (允许 %d 个系统调用)\n", len(profile.Syscalls[0].Names))
	return cachedSeccompJSON
}
