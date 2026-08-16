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
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	fapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/Falthera/ReverseDrop/internal/app"
	"github.com/Falthera/ReverseDrop/internal/config"
	"github.com/Falthera/ReverseDrop/internal/discovery"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/discovery/parser"
	"github.com/Falthera/ReverseDrop/internal/notification"
	"github.com/Falthera/ReverseDrop/internal/platform"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
	"github.com/Falthera/ReverseDrop/internal/transfer"
	"github.com/Falthera/ReverseDrop/internal/trust"
)

type guiApp struct {
	fyneApp          fyne.App
	window           fyne.Window
	cfg              *config.Config
	trustStore       *trust.Store
	service          *app.Service
	regAdapter       *app.PeerRegistryAdapter
	mgr              *discovery.Manager
	transferMgr      *transfer.Manager
	notifier         notification.Notifier
	peersList        *widget.List
	statusLabel      *widget.Label
	progressBar      *widget.ProgressBar
	progressLabel    *widget.Label
	scanBtn          *widget.Button
	stopBtn          *widget.Button
	sendBtn          *widget.Button
	modeLabel        *widget.Label
	autoAcceptLabel  *widget.Label
	selectedPeer     peer.Peer
	hasSelection     bool
	transferStart    time.Time
	lastProgressTime time.Time
	lastBytesSent    int64
}

func newGUIApp() *guiApp {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	if cfg.DiscoveryMode == "" {
		cfg.DiscoveryMode = string(config.DiscoveryModeEveryone)
	}

	g := &guiApp{regAdapter: app.NewPeerRegistryAdapter(peer.NewRegistry()), cfg: cfg}
	g.trustStore = trust.NewStore("")
	g.fyneApp = fapp.NewWithID("com.falthera.reversedrop")
	g.fyneApp.Settings().SetTheme(fyne.CurrentApp().Settings().Theme())
	g.window = g.fyneApp.NewWindow("ReverseDrop")
	g.window.Resize(fyne.NewSize(900, 600))
	g.transferMgr = transfer.NewManager(transfer.DefaultPort, "")
	g.transferMgr.SetAutoAccept(cfg.AutoAcceptTrusted, g.trustStore)
	g.notifier = notification.NewNotifier()
	g.window.SetContent(g.buildUI())
	g.registerLifecycle()
	return g
}

func (g *guiApp) registerLifecycle() {
	g.fyneApp.Lifecycle().SetOnEnteredBackground(func() {
		if g.mgr != nil {
			g.mgr.SetBackground(true)
		}
		g.statusLabel.SetText("In background (scanning paused)")
	})
	g.fyneApp.Lifecycle().SetOnExitedBackground(func() {
		if g.mgr != nil {
			g.mgr.SetBackground(false)
		}
		if g.mgr != nil && g.mgr != nil {
			g.statusLabel.SetText("Scanning...")
		}
	})
}

func (g *guiApp) buildUI() fyne.CanvasObject {
	g.statusLabel = widget.NewLabel("Ready")
	g.statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	g.progressBar = widget.NewProgressBar()
	g.progressBar.Min = 0
	g.progressBar.Max = 1
	g.progressBar.Hide()
	g.progressLabel = widget.NewLabel("")
	g.progressLabel.Hide()

	discoveryModeText := "Everyone"
	if g.cfg.DiscoveryMode == string(config.DiscoveryModeContactsOnly) {
		discoveryModeText = "Contacts Only"
	}
	g.modeLabel = widget.NewLabel(fmt.Sprintf("Discovery Mode: %s", discoveryModeText))
	autoAcceptText := "Off"
	if g.cfg.AutoAcceptTrusted {
		autoAcceptText = "On"
	}
	g.autoAcceptLabel = widget.NewLabel(fmt.Sprintf("Auto-Accept Trusted: %s", autoAcceptText))

	settingsRow := container.NewHBox(g.modeLabel, widget.NewLabel("  |  "), g.autoAcceptLabel)

	g.peersList = widget.NewList(
		func() int { return len(g.regAdapter.PeerRegistry.List()) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("device name"),
				widget.NewLabel("address"),
				widget.NewLabel("status"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			peers := g.regAdapter.PeerRegistry.List()
			if i < 0 || i >= len(peers) {
				return
			}
			p := peers[i]
			box := o.(*fyne.Container)
			name := box.Objects[0].(*widget.Label)
			addr := box.Objects[1].(*widget.Label)
			status := box.Objects[2].(*widget.Label)
			name.SetText(p.DeviceName)
			addr.SetText(p.Address)
			status.SetText(string(p.State))
		},
	)
	g.peersList.OnSelected = func(id widget.ListItemID) {
		peers := g.regAdapter.PeerRegistry.List()
		if id < 0 || id >= len(peers) {
			g.hasSelection = false
			g.selectedPeer = peer.Peer{}
			g.statusLabel.SetText("Ready")
			return
		}
		p := peers[id]
		g.selectedPeer = p
		g.hasSelection = true
		g.statusLabel.SetText(fmt.Sprintf("Selected: %s (%s)", p.DeviceName, p.Address))
		g.sendBtn.Enable()
	}

	g.scanBtn = widget.NewButton("Scan", g.startScan)
	g.stopBtn = widget.NewButton("Stop", g.stopScan)
	g.stopBtn.Disable()
	g.sendBtn = widget.NewButton("Send File", g.sendFile)
	g.sendBtn.Disable()

	toolbar := container.NewHBox(g.scanBtn, g.stopBtn, g.sendBtn)

	header := container.NewVBox(
		widget.NewLabelWithStyle("ReverseDrop", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		g.statusLabel,
		g.progressLabel,
		g.progressBar,
		settingsRow,
		toolbar,
	)

	content := container.NewBorder(header, nil, nil, nil, g.peersList)

	return content
}

func (g *guiApp) startScan() {
	caps := app.NewCapabilitySet()
	capReporter := platform.NewDefaultReporter()
	avail, err := capReporter.BluetoothAvailable()
	status, detail := g.capabilityStatus(avail, err)
	caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: status, Detail: detail})

	var scanner ble.Scanner
	scanner = ble.NewFakeScanner(nil, 500*time.Millisecond)

	svc, err := app.NewService(scanner)
	if err != nil {
		g.statusLabel.SetText("Error: " + err.Error())
		return
	}
	g.service = svc
	g.regAdapter = svc.Registry()

	mgr := discovery.NewManager(scanner, g.regAdapter, caps)
	if err := mgr.Start(); err != nil {
		g.statusLabel.SetText("Error: " + err.Error())
		return
	}

	if btInfo, ok := caps.Get(app.CapabilityBluetooth); ok && btInfo.Status != app.CapabilityAvailable {
		g.showBluetoothError(btInfo)
	}

	g.mgr = mgr
	g.scanBtn.Disable()
	g.stopBtn.Enable()
	g.statusLabel.SetText("Scanning...")
	_ = g.notifier.Send("ReverseDrop", "Scanning for nearby devices...")

	go func() {
		for evt := range g.regAdapter.PeerRegistry.Events() {
			switch evt.Type {
			case peer.EventPeerUpserted, peer.EventPeerUpdated:
				_ = g.notifier.Send("ReverseDrop", fmt.Sprintf("Peer discovered: %s (%s)", evt.Peer.DeviceName, evt.Peer.Address))
				g.peersList.Refresh()
			case peer.EventPeerRemoved:
				g.peersList.Refresh()
			}
		}
	}()
}

func (g *guiApp) capabilityStatus(avail bool, err error) (app.CapabilityStatus, string) {
	status := app.CapabilityAvailable
	detail := "Bluetooth is available"
	if err != nil {
		detail = err.Error()
		if bluetoothErr, ok := err.(*platform.BluetoothError); ok {
			switch bluetoothErr.Category {
			case platform.BluetoothErrorPermissionDenied:
				status = app.CapabilityPermissionDenied
			case platform.BluetoothErrorUnavailable:
				status = app.CapabilityUnavailable
			case platform.BluetoothErrorDisabled:
				status = app.CapabilityUnavailable
			default:
				status = app.CapabilityUnavailable
			}
		} else {
			status = app.CapabilityUnavailable
		}
	}
	if !avail && status == app.CapabilityAvailable {
		status = app.CapabilityUnavailable
	}
	return status, detail
}

func (g *guiApp) showBluetoothError(info app.CapabilityInfo) {
	dialog.ShowError(fmt.Errorf("%s", info.Detail), g.window)
}

func (g *guiApp) stopScan() {
	if g.mgr != nil {
		g.mgr.Stop()
		g.mgr = nil
	}
	g.scanBtn.Enable()
	g.stopBtn.Disable()
	g.statusLabel.SetText("Scan stopped")
	g.sendBtn.Disable()
}

func (g *guiApp) sendFile() {
	if !g.hasSelection {
		g.statusLabel.SetText("Select a peer first")
		return
	}
	peer := g.selectedPeer
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		path := reader.URI().Path()
		reader.Close()
		g.startTransfer(path, peer)
	}, g.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".*"}))
	fd.Show()
}

func formatBytes(b int64) string {
	f := float64(b)
	if f < 1024 {
		return fmt.Sprintf("%.0f B", f)
	}
	f /= 1024
	if f < 1024 {
		return fmt.Sprintf("%.1f KB", f)
	}
	f /= 1024
	if f < 1024 {
		return fmt.Sprintf("%.1f MB", f)
	}
	f /= 1024
	return fmt.Sprintf("%.1f GB", f)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

func (g *guiApp) startTransfer(path string, peer peer.Peer) {
	g.sendBtn.Disable()
	g.progressBar.Show()
	g.progressLabel.Show()
	g.progressBar.SetValue(0)
	g.transferStart = time.Now()
	g.lastProgressTime = time.Now()
	g.lastBytesSent = 0

	go func() {
		stat, err := os.Stat(path)
		if err != nil {
			g.statusLabel.SetText("Error: " + err.Error())
			g.resetTransferUI()
			return
		}
		addr := fmt.Sprintf("%s:%d", peer.Address, transfer.DefaultPort)
		req := transfer.TransferRequest{
			ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
			FileName:   filepath.Base(path),
			FileSize:   stat.Size(),
			SenderName: "ReverseDrop User",
		}
		resp, err := transfer.SendFile(context.Background(), addr, req, func(state string, sent, total int64) {
			now := time.Now()
			if total > 0 {
				pct := float64(sent) / float64(total) * 100
				elapsed := now.Sub(g.transferStart).Seconds()
				speedStr := ""
				etaStr := ""
				if elapsed > 0 && sent > 0 {
					speed := float64(sent) / elapsed
					remaining := total - sent
					if speed > 0 {
						eta := time.Duration(float64(remaining)/speed) * time.Second
						speedStr = formatBytes(int64(speed)) + "/s"
						etaStr = "ETA: " + formatDuration(eta)
					}
				}
				if speedStr != "" {
					g.progressLabel.SetText(fmt.Sprintf("%s: %.1f%% (%s, %s)", state, pct, speedStr, etaStr))
				} else {
					g.progressLabel.SetText(fmt.Sprintf("%s: %.1f%%", state, pct))
				}
				g.progressBar.SetValue(float64(sent) / float64(total))
			} else {
				g.progressLabel.SetText(state)
				g.progressBar.SetValue(0)
			}
			g.lastProgressTime = now
			g.lastBytesSent = sent
		})
		if err != nil {
			g.statusLabel.SetText("Transfer failed: " + err.Error())
			_ = g.notifier.Send("ReverseDrop", "Transfer failed: "+err.Error())
		} else if resp != nil && resp.Accepted {
			g.statusLabel.SetText("Transfer complete")
			_ = g.notifier.Send("ReverseDrop", "Transfer complete: "+filepath.Base(path))
		} else if resp != nil {
			g.statusLabel.SetText("Transfer rejected: " + resp.Error)
			_ = g.notifier.Send("ReverseDrop", "Transfer rejected: "+resp.Error)
		}
		g.resetTransferUI()
	}()
}

func (g *guiApp) resetTransferUI() {
	g.sendBtn.Enable()
	g.progressBar.SetValue(0)
	g.progressBar.Hide()
	g.progressLabel.Hide()
}

func main() {
	g := newGUIApp()
	g.window.ShowAndRun()
}
