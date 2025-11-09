package config

import (
	"net/url"
	"strconv"
	"strings"

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
	DataSources DataSourcesConfig
	Auth        AuthConfig
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

// DataSourcesConfig v0.5 multi-source collection configuration
type DataSourcesConfig struct {
	YahooFetchInterval      int
	ScraperFetchInterval    int
	TwitterFetchInterval    int
	SentimentUpdateInterval int

	TwitterAPIKey       string
	TwitterAPISecret    string
	TwitterAccessToken  string
	TwitterAccessSecret string

	SymbolUniverse string // comma-separated symbols (e.g. THYAO,AKBNK,ASELS)
}

// AuthConfig v1.0 authentication configuration
type AuthConfig struct {
	JWTSecret string // JWT signing secret
	APIKey    string // Master API key for authentication
}

// parseDatabaseURL parses DATABASE_URL and returns DatabaseConfig
// Supports both DATABASE_URL and individual DB_* variables
func parseDatabaseURL() DatabaseConfig {
	dbURL := viper.GetString("DATABASE_URL")
	if dbURL != "" {
		parsedURL, err := url.Parse(dbURL)
		if err == nil {
			host := parsedURL.Hostname()
			port := parsedURL.Port()
			if port == "" {
				port = "5432"
			}
			user := parsedURL.User.Username()
			password, _ := parsedURL.User.Password()
			dbName := strings.TrimPrefix(parsedURL.Path, "/")
			sslMode := "require"
			if parsedURL.Query().Get("sslmode") != "" {
				sslMode = parsedURL.Query().Get("sslmode")
			}

			return DatabaseConfig{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
				DBName:   dbName,
				SSLMode:  sslMode,
			}
		}
	}

	// Fallback to individual variables
	return DatabaseConfig{
		Host:     viper.GetString("DB_HOST"),
		Port:     viper.GetString("DB_PORT"),
		User:     viper.GetString("DB_USER"),
		Password: viper.GetString("DB_PASSWORD"),
		DBName:   viper.GetString("DB_NAME"),
		SSLMode:  viper.GetString("DB_SSLMODE"),
	}
}

// parseRedisURL parses REDIS_URL and returns RedisConfig
// Supports both REDIS_URL and individual REDIS_* variables
func parseRedisURL() RedisConfig {
	redisURL := viper.GetString("REDIS_URL")
	if redisURL != "" {
		parsedURL, err := url.Parse(redisURL)
		if err == nil {
			host := parsedURL.Hostname()
			port := parsedURL.Port()
			if port == "" {
				port = "6379"
			}
			password, _ := parsedURL.User.Password()
			db := 0
			if parsedURL.Path != "" {
				if dbNum, err := strconv.Atoi(strings.TrimPrefix(parsedURL.Path, "/")); err == nil {
					db = dbNum
				}
			}

			return RedisConfig{
				Host:     host,
				Port:     port,
				Password: password,
				DB:       db,
			}
		}
	}

	// Fallback to individual variables
	return RedisConfig{
		Host:     viper.GetString("REDIS_HOST"),
		Port:     viper.GetString("REDIS_PORT"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	}
}

// Load .env dosyasından veya environment variables'dan konfigürasyonu yükler
func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // Environment variables'ları otomatik oku

	// .env dosyası yoksa hata verme, sadece environment variables kullan
	_ = viper.ReadInConfig() // Ignore error if .env file doesn't exist

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("PORT"),
			Env:  viper.GetString("ENV"),
		},
		Database: parseDatabaseURL(),
		Redis:    parseRedisURL(),
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
		DataSources: DataSourcesConfig{
			YahooFetchInterval:      viper.GetInt("YAHOO_FETCH_INTERVAL"),
			ScraperFetchInterval:    viper.GetInt("SCRAPER_FETCH_INTERVAL"),
			TwitterFetchInterval:    viper.GetInt("TWITTER_FETCH_INTERVAL"),
			SentimentUpdateInterval: viper.GetInt("SENTIMENT_UPDATE_INTERVAL"),

			TwitterAPIKey:       viper.GetString("TWITTER_API_KEY"),
			TwitterAPISecret:    viper.GetString("TWITTER_API_SECRET"),
			TwitterAccessToken:  viper.GetString("TWITTER_ACCESS_TOKEN"),
			TwitterAccessSecret: viper.GetString("TWITTER_ACCESS_SECRET"),
			SymbolUniverse:      viper.GetString("SYMBOL_UNIVERSE"),
		},
		Auth: AuthConfig{
			JWTSecret: viper.GetString("JWT_SECRET"),
			APIKey:    viper.GetString("API_KEY"),
		},
	}

	return config, nil
}
