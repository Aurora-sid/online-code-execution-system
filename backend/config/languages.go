package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type LanguageConfig struct {
	ID         string   `yaml:"id"`
	Image      string   `yaml:"image"`
	Filename   string   `yaml:"filename"`
	CompileCmd []string `yaml:"compile_cmd"`
	RunCmd     []string `yaml:"run_cmd"`
}

type LanguagesConfig struct {
	Languages []LanguageConfig `yaml:"languages"`
}

// Global variable to hold loaded languages
var Languages map[string]LanguageConfig

func LoadLanguages(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取语言配置失败: %w", err)
	}

	var cfg LanguagesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("解析语言配置失败: %w", err)
	}

	Languages = make(map[string]LanguageConfig)
	for _, lang := range cfg.Languages {
		Languages[lang.ID] = lang
	}

	return nil
}

// GetLanguageConfig 返回特定语言的配置
func GetLanguageConfig(id string) (LanguageConfig, error) {
	if cfg, ok := Languages[id]; ok {
		return cfg, nil
	}
	return LanguageConfig{}, fmt.Errorf("不支持的语言: %s", id)
}
