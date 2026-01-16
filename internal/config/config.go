package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Qdrant   QdrantConfig   `mapstructure:"qdrant"`
	OpenAI   ModelConfig    `mapstructure:"openai"`
	DeepSeek ModelConfig    `mapstructure:"deepseek"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type QdrantConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	CollectionName string `mapstructure:"collection_name"`
}

type ModelConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	Model   string `mapstructure:"model"`
}

// LoadConfig 读取配置文件
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // 配置文件名 (不带扩展名)
	viper.SetConfigType("yaml")   // 文件类型
	viper.AddConfigPath(".")      // 查找路径：根目录

	// 这一步是为了支持环境变量覆盖 (例如在 Docker 中)
	// 比如设置环境变量 FACETAX_OPENAI_API_KEY 可以覆盖 yaml 里的值
	viper.SetEnvPrefix("FACETAX")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}
