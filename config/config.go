package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

const pathToConfig = "../../config/config.yaml"

type Config struct {
	Env        string `yaml:"env"`
	HTTPServer struct {
		Address string `yaml:"address"`
		Timeout string `yaml:"timeout"`
	} `yaml:"http_server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`
	EmailConfirm EmailConfirm `yaml:"email_confirm"`
}

type EmailConfirm struct {
		EmailFrom string `yaml:"email_from"`
		PasswordFrom string `yaml:"password_from"`
		SmtpHost string `yaml:"smtp_host"`
		SmtpPort string `yaml:"smtp_port"`
} 

var config Config

func LoadConfig() (Config, error) {
	data, err := os.ReadFile(pathToConfig)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func GetConfigEmail() EmailConfirm {
	return config.EmailConfirm
}