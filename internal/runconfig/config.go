package runconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogFile             string `yaml:"log_file"`
	Start               string `yaml:"start"`
	End                 string `yaml:"end"`
	EventName           string `yaml:"event_name"`
	AssistWindowSeconds int    `yaml:"assist_window_seconds"`
	TeamsEvent          bool   `yaml:"teams_event"`
	DataDir             string `yaml:"data_dir"`
	HTMLDir             string `yaml:"html_dir"`
}

func (config *Config) ApplyDefaults() {
	if config.AssistWindowSeconds <= 0 {
		config.AssistWindowSeconds = 30
	}
	if strings.TrimSpace(config.DataDir) == "" {
		config.DataDir = "data"
	}
	if strings.TrimSpace(config.HTMLDir) == "" {
		config.HTMLDir = "html"
	}
}

func (config Config) Validate() error {
	if strings.TrimSpace(config.LogFile) == "" {
		return fmt.Errorf("log_file is required")
	}
	if err := validateClock("start", config.Start); err != nil {
		return err
	}
	if err := validateClock("end", config.End); err != nil {
		return err
	}
	if strings.TrimSpace(config.EventName) == "" {
		return fmt.Errorf("event_name is required")
	}
	if config.AssistWindowSeconds <= 0 {
		return fmt.Errorf("assist_window_seconds must be greater than zero")
	}
	if strings.TrimSpace(config.DataDir) == "" {
		return fmt.Errorf("data_dir is required")
	}
	if strings.TrimSpace(config.HTMLDir) == "" {
		return fmt.Errorf("html_dir is required")
	}

	return nil
}

func Load(reader io.Reader) (Config, error) {
	var config Config
	if err := yaml.NewDecoder(reader).Decode(&config); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	config.ApplyDefaults()
	if err := config.Validate(); err != nil {
		return Config{}, err
	}

	return config, nil
}

func LoadFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	return Load(file)
}

func Write(writer io.Writer, config Config) error {
	config.ApplyDefaults()
	if err := config.Validate(); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(2)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("close config encoder: %w", err)
	}

	return nil
}

func ResolvePath(configPath string, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}

	return filepath.Clean(filepath.Join(filepath.Dir(configPath), value))
}

func validateClock(field string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if _, err := time.Parse("15:04:05", value); err != nil {
		return fmt.Errorf("%s must use HH:MM:SS: %w", field, err)
	}

	return nil
}
