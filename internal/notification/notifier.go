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
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Notifier interface {
	Send(title, message string) error
}

type noopNotifier struct{}

func (n *noopNotifier) Send(title, message string) error {
	return nil
}

func NewNotifier() Notifier {
	switch runtime.GOOS {
	case "darwin":
		return &darwinNotifier{}
	case "linux":
		return &linuxNotifier{}
	case "windows":
		return &windowsNotifier{}
	default:
		return &noopNotifier{}
	}
}

type darwinNotifier struct{}

func (n *darwinNotifier) Send(title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, strings.ReplaceAll(message, `"`, `\"`), strings.ReplaceAll(title, `"`, `\"`))
	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	return cmd.Run()
}

type linuxNotifier struct{}

func (n *linuxNotifier) Send(title, message string) error {
	cmd := exec.Command("notify-send", title, message)
	return cmd.Run()
}

type windowsNotifier struct{}

func (n *windowsNotifier) Send(title, message string) error {
	script := fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; $template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); $textNodes = $template.GetElementsByTagName("text"); $textNodes.Item(0).AppendChild($template.CreateTextNode("%s")) > $null; $textNodes.Item(1).AppendChild($template.CreateTextNode("%s")) > $null; [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("%s").Show([Windows.UI.Notifications.ToastNotificationManager]::new([Windows.UI.Notifications.ToastNotification]::new($template))`, title, message, title)
	ps := exec.Command("powershell", "-Command", script)
	var out bytes.Buffer
	ps.Stdout = &out
	ps.Stderr = &out
	return ps.Run()
}
