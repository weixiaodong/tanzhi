package config

import (
	"encoding/json"

	"github.com/spf13/viper"
)

var (
	GConfig = &Config{}
)

func Init(v *viper.Viper) (err error) {
	err = v.Unmarshal(GConfig)
	if err != nil {
		return err
	}

	return
}

type Config struct {
	Env       string    `mapstructure:"env"`
	Transport Transport `mapstructure:"transport"`
	Jobs      []Job     `mapstructure:"jobs"`
}

type Transport struct {
	HTTP HTTPConfig `mapstructure:"http"`
}

type HTTPConfig struct {
	Addr              string  `mapstructure:"addr"`
	ReadTimeout       float64 `mapstructure:"read_timeout"`
	ReadHeaderTimeout float64 `mapstructure:"read_header_timeout"`
	WriteTimeout      float64 `mapstructure:"write_timeout"`
	IdleTimeout       float64 `mapstructure:"idle_timeout"`
}

type Job struct {
	Name    string        `mapstructure:"name"`
	Expr    string        `mapstructure:"expr"`
	Command CommandConfig `mapstructure:"command"`
}

type CommandConfig struct {
	Type   string `mapstructure:"type", json:"type"`
	Method string `mapstructure:"method", json:"method"`
	Target string `mapstructure:"target", json:"target"`
}

func (c *CommandConfig) Encode() string {
	if c == nil {
		return ""
	}
	b, _ := json.Marshal(c)
	return string(b)
}
