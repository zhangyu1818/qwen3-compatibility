package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	DashScope DashScopeConfig `mapstructure:"dashscope"`
	Upload    UploadConfig    `mapstructure:"upload"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DashScopeConfig struct {
	Timeout int `mapstructure:"timeout"`
}

type UploadConfig struct {
	MaxFileSize  int64    `mapstructure:"max_file_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
}

func Load() (*Config, error) {
	config := &Config{}

	// Initialize viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("$HOME/.config/qwen3-compatibility")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Bind to environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("QWEN_COMPAT")

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return config, nil
}

func InitializeFlags(cmd *cobra.Command) {
	// Server flags
	cmd.Flags().String("host", "0.0.0.0", "Server host")
	cmd.Flags().String("port", "9000", "Server port")
}

// BindFlags binds command flags to viper
// This must be called after flags are parsed
func BindFlags(cmd *cobra.Command) {
	// Bind flags to viper
	_ = viper.BindPFlag("server.host", cmd.Flags().Lookup("host"))
	_ = viper.BindPFlag("server.port", cmd.Flags().Lookup("port"))
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "9000")
	viper.SetDefault("dashscope.timeout", 30)
	viper.SetDefault("upload.max_file_size", 100*1024*1024) // 100MB
	viper.SetDefault("upload.allowed_types", []string{
		"audio/aac", "audio/amr", "audio/flac", "audio/mp3", "audio/mpeg",
		"audio/mp4", "audio/x-m4a", "audio/ogg", "audio/opus",
		"audio/wav", "audio/wave", "audio/x-wav", "audio/webm",
		"audio/x-ms-wma", "video/x-msvideo", "video/x-flv",
		"video/x-matroska", "video/quicktime", "video/mp4",
		"video/mpeg", "video/webm", "video/x-ms-wmv",
	})
}

func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	return nil
}

func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}
