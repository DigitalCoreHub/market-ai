package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Log         LogConfig
	News        NewsConfig
	AI          AIConfig
	Leaderboard LeaderboardConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type LogConfig struct {
	Level string
}

// NewsConfig v0.3 news aggregation configuration
type NewsConfig struct {
	APIKey         string
	UpdateInterval int    // minutes
	CacheTTL       int    // minutes
	Feeds          string // comma-separated RSS feeds
}

// AIConfig v0.3 AI integration configuration
type AIConfig struct {
	OpenAIKey    string
	AnthropicKey string
	GPTModel     string
	ClaudeModel  string
	Temperature  float64
	MaxTokens    int

	// v0.4 providers
	GoogleKey     string
	GoogleModel   string
	DeepSeekKey   string
	DeepSeekModel string
	GroqKey       string
	GroqModel     string
	MistralKey    string
	MistralModel  string
	XAIKey        string
	XAIModel      string
	GPT4MiniModel string

	// Cost optimization flags
	BudgetMode          bool
	EnablePremiumModels bool
}

// LeaderboardConfig v0.4 leaderboard update interval
type LeaderboardConfig struct {
	UpdateInterval int // seconds
}

// Load .env dosyasından konfigürasyonu yükler
func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("PORT"),
			Env:  viper.GetString("ENV"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
		News: NewsConfig{
			APIKey:         viper.GetString("NEWS_API_KEY"),
			UpdateInterval: viper.GetInt("NEWS_UPDATE_INTERVAL"),
			CacheTTL:       viper.GetInt("NEWS_CACHE_TTL"),
			Feeds:          viper.GetString("RSS_FEEDS"),
		},
		AI: AIConfig{
			OpenAIKey:    viper.GetString("OPENAI_API_KEY"),
			AnthropicKey: viper.GetString("ANTHROPIC_API_KEY"),
			GPTModel:     viper.GetString("AI_MODEL_GPT"),
			ClaudeModel:  viper.GetString("AI_MODEL_CLAUDE"),
			Temperature:  viper.GetFloat64("AI_TEMPERATURE"),
			MaxTokens:    viper.GetInt("AI_MAX_TOKENS"),

			GoogleKey:     viper.GetString("GOOGLE_API_KEY"),
			GoogleModel:   viper.GetString("AI_MODEL_GEMINI"),
			DeepSeekKey:   viper.GetString("DEEPSEEK_API_KEY"),
			DeepSeekModel: viper.GetString("AI_MODEL_DEEPSEEK"),
			GroqKey:       viper.GetString("GROQ_API_KEY"),
			GroqModel:     viper.GetString("AI_MODEL_LLAMA"),
			MistralKey:    viper.GetString("MISTRAL_API_KEY"),
			MistralModel:  viper.GetString("AI_MODEL_MIXTRAL"),
			XAIKey:        viper.GetString("XAI_API_KEY"),
			XAIModel:      viper.GetString("AI_MODEL_GROK"),
			GPT4MiniModel: viper.GetString("AI_MODEL_GPT4_MINI"),

			BudgetMode:          viper.GetBool("BUDGET_MODE"),
			EnablePremiumModels: viper.GetBool("ENABLE_PREMIUM_MODELS"),
		},
		Leaderboard: LeaderboardConfig{
			UpdateInterval: viper.GetInt("LEADERBOARD_UPDATE_INTERVAL"),
		},
	}

	return config, nil
}
