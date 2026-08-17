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
package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type NotificationCategory string

const (
	NotificationCategoryTransferStarted  NotificationCategory = "transfer_started"
	NotificationCategoryTransferProgress NotificationCategory = "transfer_progress"
	NotificationCategoryTransferCompleted NotificationCategory = "transfer_completed"
	NotificationCategoryTransferFailed   NotificationCategory = "transfer_failed"
)

type NotificationPreferences struct {
	Enabled        bool                          `json:"enabled"`
	CategorySounds map[NotificationCategory]bool `json:"category_sounds,omitempty"`
	Vibration      bool                          `json:"vibration"`
}

type Notifier interface {
	SendWithCategory(category NotificationCategory, title, message string) error
	Send(title, message string) error
	SetEnabled(enabled bool)
	SetCategorySound(category NotificationCategory, enabled bool)
	SetVibration(enabled bool)
	LoadPreferences()
	SavePreferences()
}

type noopNotifier struct{}

func (n *noopNotifier) SendWithCategory(category NotificationCategory, title, message string) error {
	return nil
}

func (n *noopNotifier) Send(title, message string) error {
	return nil
}

func (n *noopNotifier) SetEnabled(enabled bool) {}

func (n *noopNotifier) SetCategorySound(category NotificationCategory, enabled bool) {}

func (n *noopNotifier) SetVibration(enabled bool) {}

func (n *noopNotifier) LoadPreferences() {}

func (n *noopNotifier) SavePreferences() {}

type notifier struct {
	prefs     *NotificationPreferences
	prefsPath string
	mu        sync.RWMutex
	impl      Notifier
}

func NewNotifier() Notifier {
	cfgDir, _ := os.UserConfigDir()
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	prefsPath := filepath.Join(cfgDir, "reversedrop", "notification_prefs.json")

	var impl Notifier
	switch runtime.GOOS {
	case "darwin":
		impl = &darwinNotifier{}
	case "linux":
		impl = &linuxNotifier{}
	case "windows":
		impl = &windowsNotifier{}
	default:
		impl = &noopNotifier{}
	}

	n := &notifier{
		prefs:     &NotificationPreferences{Enabled: true, Vibration: true, CategorySounds: map[NotificationCategory]bool{NotificationCategoryTransferCompleted: true, NotificationCategoryTransferFailed: true}},
		prefsPath: prefsPath,
		impl:      impl,
	}
	n.LoadPreferences()
	return n
}

func (n *notifier) SendWithCategory(category NotificationCategory, title, message string) error {
	n.mu.RLock()
	enabled := n.prefs.Enabled
	vibration := n.prefs.Vibration
	soundEnabled := n.prefs.CategorySounds[category]
	n.mu.RUnlock()

	if !enabled {
		return nil
	}

	err := n.impl.Send(title, message)
	if err != nil {
		return err
	}

	if soundEnabled {
		n.playSound()
	}
	if vibration {
		n.triggerVibration()
	}
	return nil
}

func (n *notifier) Send(title, message string) error {
	return n.SendWithCategory(NotificationCategoryTransferProgress, title, message)
}

func (n *notifier) SetEnabled(enabled bool) {
	n.mu.Lock()
	n.prefs.Enabled = enabled
	n.mu.Unlock()
	n.SavePreferences()
}

func (n *notifier) SetCategorySound(category NotificationCategory, enabled bool) {
	n.mu.Lock()
	if n.prefs.CategorySounds == nil {
		n.prefs.CategorySounds = make(map[NotificationCategory]bool)
	}
	n.prefs.CategorySounds[category] = enabled
	n.mu.Unlock()
	n.SavePreferences()
}

func (n *notifier) SetVibration(enabled bool) {
	n.mu.Lock()
	n.prefs.Vibration = enabled
	n.mu.Unlock()
	n.SavePreferences()
}

func (n *notifier) LoadPreferences() {
	data, err := os.ReadFile(n.prefsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		slog.Warn("failed to read notification prefs", "error", err)
		return
	}
	var prefs NotificationPreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		slog.Warn("failed to parse notification prefs", "error", err)
		return
	}
	n.mu.Lock()
	n.prefs = &prefs
	n.mu.Unlock()
}

func (n *notifier) SavePreferences() {
	n.mu.RLock()
	prefs := n.prefs
	n.mu.RUnlock()

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		slog.Warn("failed to marshal notification prefs", "error", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(n.prefsPath), 0o700); err != nil {
		slog.Warn("failed to create notification prefs dir", "error", err)
		return
	}
	if err := os.WriteFile(n.prefsPath, data, 0o600); err != nil {
		slog.Warn("failed to write notification prefs", "error", err)
	}
}

func (n *notifier) playSound() {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Start()
	case "linux":
		_ = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/message.oga").Start()
	case "windows":
		_ = exec.Command("powershell", "-Command", "[console]::beep(1000,200)").Start()
	}
}

func (n *notifier) triggerVibration() {
	switch runtime.GOOS {
	case "android":
		_ = exec.Command("termux-vibrate", "-d", "200").Start()
	}
}

type darwinNotifier struct {
	noopNotifier
}

func (n *darwinNotifier) Send(title, message string) error {
	slog.Debug("sending desktop notification", "os", "darwin", "title", title, "message", message)
	script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, strings.ReplaceAll(message, `"`, `\"`), strings.ReplaceAll(title, `"`, `\"`))
	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send desktop notification: %w", err)
	}
	return nil
}

type linuxNotifier struct {
	noopNotifier
}

func (n *linuxNotifier) Send(title, message string) error {
	slog.Debug("sending desktop notification", "os", "linux", "title", title, "message", message)
	cmd := exec.Command("notify-send", title, message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send desktop notification: %w", err)
	}
	return nil
}

type windowsNotifier struct {
	noopNotifier
}

func (n *windowsNotifier) Send(title, message string) error {
	slog.Debug("sending desktop notification", "os", "windows", "title", title, "message", message)
	script := fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; $template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); $textNodes = $template.GetElementsByTagName("text"); $textNodes.Item(0).AppendChild($template.CreateTextNode("%s")) > $null; $textNodes.Item(1).AppendChild($template.CreateTextNode("%s")) > $null; [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("%s").Show([Windows.UI.Notifications.ToastNotification]::new([Windows.UI.Notifications.ToastNotification]::new($template))`, powershellEscape(title), powershellEscape(message), powershellEscape(title))
	ps := exec.Command("powershell", "-Command", script)
	var out bytes.Buffer
	ps.Stdout = &out
	ps.Stderr = &out
	if err := ps.Run(); err != nil {
		return fmt.Errorf("failed to send desktop notification: %w", err)
	}
	return nil
}

func powershellEscape(s string) string {
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, "\"", "`\"")
	s = strings.ReplaceAll(s, "$", "`$")
	return s
}
