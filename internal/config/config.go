package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	LLMKey     string
	UseMockLLM bool
	HTTPAddr   string
	LogLevel   string
	Production bool
}

func Load() *Config {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("USE_LLM_MOCK", true)
	viper.SetDefault("PRODUCTION", false)
	viper.SetDefault("HTTP_ADDR", ":8080")
	viper.SetDefault("LOG_LEVEL", "info")

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Loaded .env config from: %s", viper.ConfigFileUsed())
	} else {
		log.Printf("No .env file loaded, assuming environment variables are set externally")
	}

	return &Config{
		LLMKey:     viper.GetString("LLM_KEY"),
		UseMockLLM: viper.GetBool("USE_LLM_MOCK"),
		HTTPAddr:   viper.GetString("HTTP_ADDR"),
		LogLevel:   viper.GetString("LOG_LEVEL"),
		Production: viper.GetBool("PRODUCTION"),
	}
}
