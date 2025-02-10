package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string
	}
	App struct {
		RepoBaseDir string `mapstructure:"repo_base_dir"`
		LogLevel    string `mapstructure:"log_level"`
	}
	Git struct {
		Username      string `mapstructure:"username"`
		Token         string `mapstructure:"token"`
		DefaultBranch string `mapstructure:"default_branch"`
		RemoteURL     string `mapstructure:"remote_url"`
	}
	Webhook struct {
		Secret string
	}
}

func LoadConfig() (*Config, error) {
	// Load .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, loading from system environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	// Viper Configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("cmd/config")

	// Read the config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		log.Println("No external config file found, using environment variables only")
	}

	// Bind environment variables manually
	viper.AutomaticEnv()

	viper.BindEnv("git.username", "GIT_USERNAME")
	viper.BindEnv("git.token", "GIT_TOKEN")
	viper.BindEnv("git.default_branch", "GIT_DEFAULT_BRANCH")
	viper.BindEnv("git.remote_url", "GIT_REMOTE_URL")
	viper.BindEnv("webhook.secret", "WEBHOOK_SECRET")
	viper.BindEnv("app.log_level", "LOG_LEVEL")
	viper.BindEnv("app.repo_base_dir", "REPO_BASE_DIR")

	// Unmarshal configuration into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Helper functions
func (cfg *Config) LogLevel() string {
	return getString("log_level", "info")
}

func (cfg *Config) Environment() string {
	return getString("environment", "development")
}

func (cfg *Config) Get(key string, defaultVal string) string {
	val := viper.GetString(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getString(key string, defaultVal string) string {
	if val, exists := viper.Get(key).(string); exists && val != "" {
		return val
	}
	return defaultVal
}
