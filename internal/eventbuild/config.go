package eventbuild

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type EventSettings struct {
	EventName           string `yaml:"event_name" json:"event_name"`
	AssistWindowSeconds int    `yaml:"assist_window_seconds" json:"assist_window_seconds"`
}

func (settings *EventSettings) ApplyDefaults() {
	if settings.AssistWindowSeconds <= 0 {
		settings.AssistWindowSeconds = 30
	}
}

type TeamConfig struct {
	Teams     []ConfiguredTeam   `yaml:"teams"`
	Ungrouped []ConfiguredMember `yaml:"ungrouped"`
	Ignored   []ConfiguredMember `yaml:"ignored"`
}

type ConfiguredTeam struct {
	Name    string             `yaml:"name"`
	Members []ConfiguredMember `yaml:"members"`
}

type ConfiguredMember struct {
	PlayerID      string `yaml:"player_id"`
	PreferredName string `yaml:"preferred_name"`
	DisplayName   string `yaml:"display_name,omitempty"`
}

func LoadEventSettings(reader io.Reader) (EventSettings, error) {
	var settings EventSettings
	if err := yaml.NewDecoder(reader).Decode(&settings); err != nil {
		return EventSettings{}, fmt.Errorf("decode event settings: %w", err)
	}

	settings.ApplyDefaults()
	return settings, nil
}

func LoadTeamConfig(reader io.Reader) (TeamConfig, error) {
	var config TeamConfig
	if err := yaml.NewDecoder(reader).Decode(&config); err != nil {
		return TeamConfig{}, fmt.Errorf("decode team config: %w", err)
	}

	return config, nil
}
