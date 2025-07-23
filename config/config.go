package config

import "github.com/spf13/viper"

type Config struct {
	NomeServico   string `mapstructure:"NOME_SERVICO"`
	Porta         string `mapstructure:"HTTP_PORTA"`
	Host          string `mapstructure:"HTTP_HOST"`
	WeatherApiKey string `mapstructure:"WEATHER_API_KEY"`
	ServiceBURL   string `mapstructure:"SERVICE_B_URL"`
	ZipkinURL     string `mapstructure:"ZIPKIN_URL"`
}

func LoadConfig() (*Config, error) {
	var cfg *Config

	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	if cfg.ZipkinURL == "" {
		cfg.ZipkinURL = "http://localhost:9411/api/v2/spans"
	}
	if cfg.ServiceBURL == "" {
		cfg.ServiceBURL = "http://localhost:8081"
	}

	return cfg, err
}

func (cfg *Config) ShowConfig() {
	if cfg == nil {
		return
	}
	println("Service Name:", cfg.NomeServico)
	println("Service Port:", cfg.Porta)
	println("Service Host:", cfg.Host)

	println("Service B URL:", cfg.ServiceBURL)
	println("Zipkin URL:", cfg.ZipkinURL)
}
