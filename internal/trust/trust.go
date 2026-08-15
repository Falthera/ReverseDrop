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
package trust

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type TrustLevel string

const (
	TrustLevelUnknown   TrustLevel = "unknown"
	TrustLevelTrusted   TrustLevel = "trusted"
	TrustLevelBlocked   TrustLevel = "blocked"
)

type Store struct {
	mu      sync.RWMutex
	devices map[string]TrustLevel
	path    string
}

func NewStore(path string) *Store {
	if path == "" {
		dir, _ := os.UserConfigDir()
		if dir == "" {
			dir = filepath.Join(os.Getenv("HOME"), ".config")
		}
		path = filepath.Join(dir, "reversedrop", "trust.json")
	}
	s := &Store{
		devices: make(map[string]TrustLevel),
		path:    path,
	}
	s.load()
	return s
}

func (s *Store) Get(address string) TrustLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if level, ok := s.devices[address]; ok {
		return level
	}
	return TrustLevelUnknown
}

func (s *Store) Set(address string, level TrustLevel) {
	s.mu.Lock()
	s.devices[address] = level
	s.mu.Unlock()
	s.save()
}

func (s *Store) Remove(address string) {
	s.mu.Lock()
	delete(s.devices, address)
	s.mu.Unlock()
	s.save()
}

func (s *Store) IsAllowed(address string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	level, ok := s.devices[address]
	return !ok || level != TrustLevelBlocked
}

func (s *Store) List() map[string]TrustLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]TrustLevel, len(s.devices))
	for k, v := range s.devices {
		out[k] = v
	}
	return out
}

func (s *Store) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var devices map[string]TrustLevel
	if err := json.Unmarshal(data, &devices); err != nil {
		return
	}
	s.mu.Lock()
	s.devices = devices
	s.mu.Unlock()
}

func (s *Store) save() {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return
	}
	data, _ := json.MarshalIndent(s.devices, "", "  ")
	os.WriteFile(s.path, data, 0o600)
}
