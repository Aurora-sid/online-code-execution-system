package model
// 主要定义数据库模型结构体，使用 GORM 标签进行 ORM 映射
/*使用 Go 业内非常知名的 ORM 框架 GORM 将该系统中相关联的核心持久化
数据表进行了与对象相互对应的映射映射定义。在运行时候负责建表以及定义
相应的底层数据的约束属性与限制等元数据。*/
import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;size:50"`
	Password  string `gorm:"size:255"`                           // 已哈希
	Role      string `gorm:"size:20;default:'user'" json:"role"` // admin 或 user
	CreatedAt time.Time
}

type Submission struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Language  string `gorm:"size:20"`
	Code      string `gorm:"type:text"`
	Input     string `gorm:"type:text"` // 标准输入数据
	Status    string `gorm:"size:20"`   // Pending, Running, Success, Failed, Timeout
	Output    string `gorm:"type:text"`
	CreatedAt time.Time
}
/*用户做出的每一次在线测试请求流水存储实体。
可以根据 UserID 进行主键外联锁定。内容包含代码文本块
请求时刻的环境语种、最后是否判定（挂死或者成功或者时间到）
、产出结果保留副本等。*/

// Language 表示支持的编程语言
type Language struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Value        string `gorm:"uniqueIndex;size:20" json:"value"` // 如: cpp, python, java
	Label        string `gorm:"size:50" json:"label"`             // 如: C++ (g++17)
	Icon         string `gorm:"size:100" json:"icon"`             // 图标文件名
	DisplayOrder int    `gorm:"default:0" json:"displayOrder"`    // 排序顺序
	Enabled      bool   `gorm:"default:true" json:"enabled"`      // 是否启用
}
