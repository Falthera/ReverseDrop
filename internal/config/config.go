// ReverseDrop - Copyright (C) ReverseDrop Contributors
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type NotificationCategory string

const (
	NotificationCategoryTransferStarted  NotificationCategory = "transfer_started"
	NotificationCategoryTransferProgress NotificationCategory = "transfer_progress"
	NotificationCategoryTransferCompleted NotificationCategory = "transfer_completed"
	NotificationCategoryTransferFailed   NotificationCategory = "transfer_failed"
)

type NotificationPreferences struct {
	Enabled        bool                    `json:"enabled"`
	CategorySounds map[NotificationCategory]bool `json:"category_sounds,omitempty"`
	Vibration      bool                    `json:"vibration"`
}

type Config struct {
	ScanTimeoutSeconds   int      `json:"scan_timeout_seconds"`
	TrustedDevices      []string `json:"trusted_devices,omitempty"`
	Theme               string   `json:"theme,omitempty"`
	AutoConnect         bool     `json:"auto_connect"`
	DiscoveryMode       string   `json:"discovery_mode"`
	AutoAcceptTrusted   bool     `json:"auto_accept_trusted"`
	MaxParallelTransfers int      `json:"max_parallel_transfers"`
	ResumeEnabled       bool     `json:"resume_enabled"`
	ChunkSize           int      `json:"chunk_size"`
	Notifications       *NotificationPreferences `json:"notifications,omitempty"`
}

func Default() *Config {
	return &Config{
		ScanTimeoutSeconds:   30,
		AutoConnect:         false,
		DiscoveryMode:       string(DiscoveryModeEveryone),
		AutoAcceptTrusted:   false,
		MaxParallelTransfers: 3,
		ResumeEnabled:       true,
		ChunkSize:           4 * 1024 * 1024,
		Theme:               "light",
		Notifications: &NotificationPreferences{
			Enabled:        true,
			CategorySounds: map[NotificationCategory]bool{
				NotificationCategoryTransferCompleted: true,
				NotificationCategoryTransferFailed:    true,
			},
			Vibration: true,
		},
	}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "reversedrop", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.MaxParallelTransfers <= 0 {
		cfg.MaxParallelTransfers = 3
	}
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 4 * 1024 * 1024
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
