package config

import (
	"fmt"
	"html/template"
	"time"

	"github.com/creasty/defaults"
	"github.com/spf13/viper"
)

type DBConfig struct {
	Host     string        `json:"host" mapstructure:"host"`
	Port     int           `json:"port" mapstructure:"port"`
	Username string        `json:"username" mapstructure:"username"`
	Password string        `json:"password" mapstructure:"password"`
	DBName   string        `json:"db_name" mapstructure:"db_name"`
	SSLMode  string        `json:"ssl_mode" mapstructure:"ssl_mode"`
	Timeout  time.Duration `json:"timeout" mapstructure:"timeout"`
}

type Configuration struct {
	App        *AppConfig       `json:"server" mapstructure:"server"`
	FileServer *FileServer      `json:"file_server" mapstructure:"file_server"`
	Generator  *GeneratorConfig `json:"generator" mapstructure:"generator"`
	DB         *DBConfig        `json:"db" mapstructure:"db"`
}

type AppConfig struct {
	Port     int      `json:"port" mapstructure:"port"`
	BasePath string   `json:"base_path" mapstructure:"base_path"`
	Timeout  *Timeout `json:"timeouts" mapstructure:"timeouts"`
}

type GeneratorConfig struct {
	Timeout      time.Duration `json:"timeout" mapstructure:"timeout"`
	ConvertURL   string        `json:"convert_url" mapstructure:"convert_url"`
	MergeURL     string        `json:"merge_url" mapstructure:"merge_url"`
	Template     string        `json:"template" mapstructure:"template"`
	TemplateFile *template.Template
}

type FileServer struct {
	URLPrefix string `json:"url_prefix" mapstructure:"url_prefix"`
}

type Timeout struct {
	Read       time.Duration `json:"read" mapstructure:"read"`
	Write      time.Duration `json:"write" mapstructure:"write"`
	ReadHeader time.Duration `json:"read_header" mapstructure:"read_header"`
	Shutdown   time.Duration `json:"shutdown" mapstructure:"shutdown"`
}

// New on success returns Configuration object and nil
// fails when there are missing fields in configuration data
func New() (*Configuration, error) {
	configFile := "config.yaml"
	viper.SetConfigFile(configFile)
	var err error
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Configuration{}

	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	cfg.Generator.TemplateFile, err = template.ParseFiles(cfg.Generator.Template)
	if err != nil {
		return nil, fmt.Errorf("config error: parsing cover html template: %s", err.Error())
	}
	return cfg, nil
}
