package config

/*该文件用于处理针对不同编程语言执行环境的动态配置。
系统不再将每种编程语言（如 C, Python, Go）应该如何编译和执行硬编码在后台代码中，
而是提供从外部 YAML 配置文件动态读取的功能，极大地增添了扩展性和灵活性。*/
import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3" //  Go 语言中用于解析和生成 YAML 格式数据的第三方库
)

type LanguageConfig struct {
	// 在languages.yaml文件中，每种编程语言的配置项包括：
	ID         string   `yaml:"id"`
	Image      string   `yaml:"image"`
	Filename   string   `yaml:"filename"`
	CompileCmd []string `yaml:"compile_cmd"`
	RunCmd     []string `yaml:"run_cmd"`
}

// LanguagesConfig 是一个包含多个 LanguageConfig 的结构体，
// 用于从 YAML 文件中加载所有支持的编程语言配置。
type LanguagesConfig struct {
	Languages []LanguageConfig `yaml:"languages"`
}

// Languages 是一个全局变量，存储了所有加载的编程语言配置，
// 以语言 ID 作为键，方便在运行时根据请求的语言类型快速获取对应的配置。
var Languages map[string]LanguageConfig // language散列表
// 全局变量声明在任何函数体之外，供整个包使用

// LoadLanguages 从指定路径的 YAML 文件中加载编程语言配置，
// 解析后存储在全局变量 Languages 中，供系统在运行时使用。
func LoadLanguages(path string) error {
	data, err := os.ReadFile(path) // 从指定路径读取 YAML 文件内容，如果读取失败则返回错误
	if err != nil {
		return fmt.Errorf("读取语言配置失败: %w", err)
	}

	var cfg LanguagesConfig // LanguagesConfig 结构体变量 cfg，存储从 YAML 文件解析后的数据
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// 正常来说，err应该为空，如果不为空则说明解析失败，返回错误
		return fmt.Errorf("解析语言配置失败: %w", err)
	}

	//使用make初始化一片空间并赋值给languages
	Languages = make(map[string]LanguageConfig)
	for _, lang := range cfg.Languages {
		Languages[lang.ID] = lang
	}

	return nil
}

// 返回特定语言的配置
// func + 函数名 + (参数列表) + (返回值列表)
func GetLanguageConfig(id string) (LanguageConfig, error) {
	if cfg, ok := Languages[id]; ok { // 根据语言 ID 从全局变量 Languages 中获取对应的配置，如果存在则返回，否则返回错误
		return cfg, nil
	}
	return LanguageConfig{}, fmt.Errorf("不支持的语言: %s", id)
}
