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
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	fapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/systray"

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
	themeBtn         *widget.Button
	modeLabel        *widget.Label
	autoAcceptLabel  *widget.Label
	offlineLabel     *widget.Label
	selectedPeer     peer.Peer
	hasSelection     bool
	transferStart    time.Time
	lastProgressTime time.Time
	lastBytesSent    int64
	transferRecords  map[string]*transfer.TransferRecord
	transfersList    *widget.List
	selectedTransfer string
	selectedFiles    []string
	filesList        *widget.List
}

func newGUIApp() *guiApp {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	if cfg.DiscoveryMode == "" {
		cfg.DiscoveryMode = string(config.DiscoveryModeEveryone)
	}

	g := &guiApp{regAdapter: app.NewPeerRegistryAdapter(peer.NewRegistry()), cfg: cfg, transferRecords: make(map[string]*transfer.TransferRecord)}
	g.trustStore = trust.NewStore("")
	g.fyneApp = fapp.NewWithID("com.falthera.reversedrop")
	g.applyTheme()
	g.window = g.fyneApp.NewWindow("ReverseDrop")
	g.window.Resize(fyne.NewSize(900, 600))
	g.transferMgr = transfer.NewManager(transfer.DefaultPort, "")
	g.transferMgr.SetAutoAccept(cfg.AutoAcceptTrusted, g.trustStore)
	g.notifier = notification.NewNotifier()
	g.window.SetContent(g.buildUI())
	g.transferMgr.SetOnTransferUpdate(func(record *transfer.TransferRecord) {
		g.transferRecords[record.ID] = record
		if g.transfersList != nil {
			fyne.Do(func() {
				g.transfersList.Refresh()
			})
		}
	})
	g.registerLifecycle()
	go g.setupTray()
	return g
}

func (g *guiApp) applyTheme() {
	if g.cfg.Theme == "dark" {
		g.fyneApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		g.fyneApp.Settings().SetTheme(theme.LightTheme())
	}
}

func (g *guiApp) toggleTheme() {
	if g.cfg.Theme == "dark" {
		g.cfg.Theme = "light"
	} else {
		g.cfg.Theme = "dark"
	}
	g.applyTheme()
	_ = config.Save(g.cfg)
	g.updateThemeBtn()
}

func (g *guiApp) updateThemeBtn() {
	if g.cfg.Theme == "dark" {
		g.themeBtn.SetText("Theme: Dark")
	} else {
		g.themeBtn.SetText("Theme: Light")
	}
}

func (g *guiApp) setupTray() {
	if runtime.GOOS == "js" || runtime.GOOS == "ios" || runtime.GOOS == "android" {
		return
	}
	systray.Run(func() {
		systray.SetIcon(getTrayIcon())
		systray.SetTooltip("ReverseDrop")
		mShow := systray.AddMenuItem("Show", "Show ReverseDrop")
		mHide := systray.AddMenuItem("Hide", "Hide ReverseDrop")
		mSend := systray.AddMenuItem("Send File", "Send a file")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Quit ReverseDrop")

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					fyne.Do(func() {
						g.window.Show()
					})
				case <-mHide.ClickedCh:
					fyne.Do(func() {
						g.window.Hide()
					})
				case <-mSend.ClickedCh:
					fyne.Do(func() {
						g.window.Show()
						g.addFiles()
					})
				case <-mQuit.ClickedCh:
					systray.Quit()
					fyne.Do(func() {
						g.window.Close()
					})
				}
			}
		}()
	}, nil)
}

func getTrayIcon() []byte {
	return []byte{}
}

func (g *guiApp) registerLifecycle() {
	g.fyneApp.Lifecycle().SetOnEnteredBackground(func() {
		if g.mgr != nil {
			g.mgr.SetBackground(true)
		}
		if g.service != nil && !g.service.IsOnline() {
			g.offlineLabel.Show()
			g.statusLabel.SetText("In background (scanning paused) - Offline")
		} else {
			g.statusLabel.SetText("In background (scanning paused)")
		}
	})
	g.fyneApp.Lifecycle().SetOnExitedBackground(func() {
		if g.mgr != nil {
			g.mgr.SetBackground(false)
		}
		if g.mgr != nil && g.mgr != nil {
			if g.service != nil && !g.service.IsOnline() {
				g.offlineLabel.Show()
				g.statusLabel.SetText("Scanning... - Offline")
			} else {
				g.statusLabel.SetText("Scanning...")
				g.offlineLabel.Hide()
			}
		}
	})
}

func (g *guiApp) buildUI() fyne.CanvasObject {
	g.statusLabel = widget.NewLabel("Ready")
	g.statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	g.offlineLabel = widget.NewLabelWithStyle("Offline - no network available", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	g.offlineLabel.Hide()

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

	g.themeBtn = widget.NewButton("Theme: Light", g.toggleTheme)
	g.updateThemeBtn()

	settingsRow := container.NewHBox(g.modeLabel, widget.NewLabel("  |  "), g.autoAcceptLabel, widget.NewLabel("  |  "), g.themeBtn)

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
	g.addBtn := widget.NewButton("Add Files", g.addFiles)
	g.sendBtn = widget.NewButton("Send Files", g.sendFiles)
	g.sendBtn.Disable()

	g.filesList = widget.NewList(
		func() int { return len(g.selectedFiles) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("file"),
				widget.NewButton("Remove", nil),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(g.selectedFiles) {
				return
			}
			box := o.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			btn := box.Objects[1].(*widget.Button)
			label.SetText(filepath.Base(g.selectedFiles[i]))
			btn.OnTapped = func() { g.removeFile(i) }
		},
	)
	g.filesList.Hide()

	toolbar := container.NewHBox(g.scanBtn, g.stopBtn, g.addBtn, g.sendBtn)

	filesHeader := widget.NewLabelWithStyle("Selected Files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	if len(g.selectedFiles) == 0 {
		filesHeader.Hide()
		g.filesList.Hide()
	}

	header := container.NewVBox(
		widget.NewLabelWithStyle("ReverseDrop", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		g.offlineLabel,
		g.statusLabel,
		g.progressLabel,
		g.progressBar,
		settingsRow,
		toolbar,
		filesHeader,
		g.filesList,
	)

	g.transfersList = widget.NewList(
		func() int { return len(g.transferRecords) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("file name"),
				widget.NewLabel("status"),
				widget.NewProgressBar(),
				container.NewHBox(),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(g.transferRecords) {
				return
			}
			var records []*transfer.TransferRecord
			for _, r := range g.transferRecords {
				records = append(records, r)
			}
			if i >= len(records) {
				return
			}
			record := records[i]
			box := o.(*fyne.Container)
			nameLabel := box.Objects[0].(*widget.Label)
			statusLabel := box.Objects[1].(*widget.Label)
			progressBar := box.Objects[2].(*widget.ProgressBar)
			actions := box.Objects[3].(*fyne.Container)

			nameLabel.SetText(record.FileName)
			statusLabel.SetText(fmt.Sprintf("%s - %s", string(record.Status), formatBytes(record.BytesReceived)))
			if record.TotalBytes > 0 {
				progressBar.SetValue(float64(record.BytesReceived) / float64(record.TotalBytes))
			} else {
				progressBar.SetValue(0)
			}
			progressBar.Show()

			actions.Objects = nil
			switch record.Status {
			case transfer.StatusActive, transfer.StatusPending:
				actions.Add(widget.NewButton("Pause", func() { g.pauseTransfer(record.ID) }))
			case transfer.StatusPaused:
				actions.Add(widget.NewButton("Resume", func() { g.resumeTransfer(record.ID) }))
			case transfer.StatusFailed:
				actions.Add(widget.NewButton("Retry", func() { g.retryTransfer(record.ID) }))
			}
			actions.Refresh()
		},
	)

	split := container.NewVSplit(g.peersList, g.transfersList)
	split.SetOffset(0.6)

	content := container.NewBorder(header, nil, nil, nil, split)

	return content
}

func (g *guiApp) addFiles() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		path := reader.URI().Path()
		reader.Close()
		g.selectedFiles = append(g.selectedFiles, path)
		g.refreshFilesList()
	}, g.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".*"}))
	fd.Show()
}

func (g *guiApp) removeFile(index int) {
	if index < 0 || index >= len(g.selectedFiles) {
		return
	}
	g.selectedFiles = append(g.selectedFiles[:index], g.selectedFiles[index+1:]...)
	g.refreshFilesList()
}

func (g *guiApp) refreshFilesList() {
	if len(g.selectedFiles) == 0 {
		g.filesList.Hide()
		g.sendBtn.Disable()
	} else {
		g.filesList.Show()
		if g.hasSelection {
			g.sendBtn.Enable()
		}
	}
	g.filesList.Refresh()
}

func (g *guiApp) sendFiles() {
	if !g.hasSelection {
		g.statusLabel.SetText("Select a peer first")
		return
	}
	if len(g.selectedFiles) == 0 {
		g.statusLabel.SetText("Add files first")
		return
	}
	peer := g.selectedPeer
	paths := make([]string, len(g.selectedFiles))
	copy(paths, g.selectedFiles)
	g.startTransfer(paths, peer)
}

func (g *guiApp) startScan() {
	if g.service != nil && !g.service.IsOnline() {
		g.offlineLabel.Show()
		g.statusLabel.SetText("Offline - no network available")
		return
	}

	caps := app.NewCapabilitySet()
	capReporter := platform.NewDefaultReporter()
	avail, err := capReporter.BluetoothAvailable()
	status, detail := g.capabilityStatus(avail, err)
	caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: status, Detail: detail})
	caps.Set(app.CapabilityInfo{Name: app.CapabilityNetworkDiscovery, Status: app.CapabilityAvailable, Detail: "Network discovery available"})

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
	g.offlineLabel.Hide()
	_ = g.notifier.SendWithCategory(config.NotificationCategoryTransferStarted, "ReverseDrop", "Scanning for nearby devices...")

	go func() {
		for evt := range g.regAdapter.PeerRegistry.Events() {
			switch evt.Type {
			case peer.EventPeerUpserted, peer.EventPeerUpdated:
				_ = g.notifier.SendWithCategory(config.NotificationCategoryTransferProgress, "ReverseDrop", fmt.Sprintf("Peer discovered: %s (%s)", evt.Peer.DeviceName, evt.Peer.Address))
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

func (g *guiApp) startTransfer(paths []string, peer peer.Peer) {
	g.sendBtn.Disable()
	g.progressBar.Show()
	g.progressLabel.Show()
	g.progressBar.SetValue(0)
	g.transferStart = time.Now()
	g.lastProgressTime = time.Now()
	g.lastBytesSent = 0

	if len(paths) == 0 {
		g.statusLabel.SetText("No files to send")
		g.resetTransferUI()
		return
	}

	totalSize := int64(0)
	fileNames := make([]string, len(paths))
	for i, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			g.statusLabel.SetText("Error: " + err.Error())
			g.resetTransferUI()
			return
		}
		totalSize += stat.Size()
		fileNames[i] = path
	}

	addr := fmt.Sprintf("%s:%d", peer.Address, transfer.DefaultPort)
	req := transfer.TransferRequest{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		FileNames:  fileNames,
		FileSizes:  make([]int64, len(paths)),
		SenderName: "ReverseDrop User",
	}
	for i, path := range paths {
		stat, _ := os.Stat(path)
		req.FileSizes[i] = stat.Size()
	}
	if len(paths) == 1 {
		req.FileName = filepath.Base(paths[0])
	}

	record := &transfer.TransferRecord{
		ID:        req.ID,
		FileName:  fmt.Sprintf("%d files", len(paths)),
		Status:    transfer.StatusActive,
		PeerAddr:  addr,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		TotalBytes: totalSize,
	}
	g.transferRecords[req.ID] = record

	go func() {
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
				fileInfo := fmt.Sprintf("%d files", len(paths))
				if len(paths) == 1 {
					fileInfo = filepath.Base(paths[0])
				}
				if speedStr != "" {
					g.progressLabel.SetText(fmt.Sprintf("%s: %s %.1f%% (%s, %s)", state, fileInfo, pct, speedStr, etaStr))
				} else {
					g.progressLabel.SetText(fmt.Sprintf("%s: %s %.1f%%", state, fileInfo, pct))
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
			record.Status = transfer.StatusFailed
			record.Error = err.Error()
			record.UpdatedAt = time.Now()
			_ = g.notifier.SendWithCategory(config.NotificationCategoryTransferFailed, "ReverseDrop", "Transfer failed: "+err.Error())
		} else if resp != nil && resp.Accepted {
			g.statusLabel.SetText("Transfer complete")
			record.Status = transfer.StatusCompleted
			record.BytesReceived = record.TotalBytes
			record.UpdatedAt = time.Now()
			fileName := filepath.Base(paths[0])
			if len(paths) > 1 {
				fileName = fmt.Sprintf("%d files", len(paths))
			}
			_ = g.notifier.SendWithCategory(config.NotificationCategoryTransferCompleted, "ReverseDrop", "Transfer complete: "+fileName)
		} else if resp != nil {
			g.statusLabel.SetText("Transfer rejected: " + resp.Error)
			record.Status = transfer.StatusFailed
			record.Error = resp.Error
			record.UpdatedAt = time.Now()
			_ = g.notifier.SendWithCategory(config.NotificationCategoryTransferFailed, "ReverseDrop", "Transfer rejected: "+resp.Error)
		}
		g.transfersList.Refresh()
		g.resetTransferUI()
	}()
}

func (g *guiApp) pauseTransfer(id string) {
	if err := g.transferMgr.PauseTransfer(id); err != nil {
		g.statusLabel.SetText("Error: " + err.Error())
		return
	}
	if r, ok := g.transferRecords[id]; ok {
		r.Status = transfer.StatusPaused
		r.UpdatedAt = time.Now()
	}
	g.transfersList.Refresh()
	g.statusLabel.SetText("Transfer paused")
}

func (g *guiApp) resumeTransfer(id string) {
	if err := g.transferMgr.ResumeTransfer(id); err != nil {
		g.statusLabel.SetText("Error: " + err.Error())
		return
	}
	if r, ok := g.transferRecords[id]; ok {
		r.Status = transfer.StatusActive
		r.Error = ""
		r.UpdatedAt = time.Now()
	}
	g.transfersList.Refresh()
	g.statusLabel.SetText("Transfer resumed")
	go func() {
		state, err := transfer.LoadTransferState(id)
		if err != nil {
			return
		}
		addr := state["addr"].(string)
		path := state["path"].(string)
		fileSize := int64(state["file_size"].(float64))
		fileName := state["file_name"].(string)
		startOffset := int64(state["sent"].(float64))

		f, err := os.Open(path)
		if err != nil {
			return
		}
		defer f.Close()

		files := []transfer.FileMeta{{
			FileType:    "public.data",
			FileName:    fileName,
			FileBomPath: "./" + fileName,
		}}
		cpioData, err := transfer.WriteCpioArchive(files, filepath.Dir(path))
		if err != nil {
			return
		}
		archive, err := transfer.CompressDVZip(cpioData)
		if err != nil {
			return
		}

		cert, err := transfer.GenerateCert()
		if err != nil {
			return
		}
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
			NextProtos:         []string{"airdrop"},
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return
		}
		defer conn.Close()

		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					conn.SetDeadline(time.Now().Add(30 * time.Second))
				}
			}
		}()

		req := transfer.TransferRequest{
			ID:         id,
			FileName:   fileName,
			FileSize:   fileSize,
			SenderName: "ReverseDrop User",
		}
		transfer.SendDiscover(conn, req)
		transfer.SendAsk(conn, req)
		sent, err := transfer.SendChunkedUploadFrom(conn, req, archive, 4*1024*1024, startOffset, func(state string, sent, total int64) {
			if r, ok := g.transferRecords[id]; ok {
				r.BytesReceived = sent
				r.UpdatedAt = time.Now()
			}
		})
		close(done)
		if err != nil {
			if r, ok := g.transferRecords[id]; ok {
				r.Status = transfer.StatusFailed
				r.Error = err.Error()
				r.UpdatedAt = time.Now()
			}
		} else {
			transfer.DeleteTransferState(id)
			if r, ok := g.transferRecords[id]; ok {
				r.Status = transfer.StatusCompleted
				r.BytesReceived = r.TotalBytes
				r.UpdatedAt = time.Now()
			}
		}
		g.transfersList.Refresh()
	}()
}

func (g *guiApp) retryTransfer(id string) {
	if r, ok := g.transferRecords[id]; ok {
		r.Status = transfer.StatusActive
		r.Error = ""
		r.UpdatedAt = time.Now()
	}
	g.transfersList.Refresh()
	g.statusLabel.SetText("Retrying transfer...")
	go func() {
		resp, err := transfer.ResumeFileTransfer(context.Background(), id, func(state string, sent, total int64) {
			if r, ok := g.transferRecords[id]; ok {
				r.BytesReceived = sent
				r.UpdatedAt = time.Now()
			}
		})
		if err != nil {
			g.statusLabel.SetText("Retry failed: " + err.Error())
			if r, ok := g.transferRecords[id]; ok {
				r.Status = transfer.StatusFailed
				r.Error = err.Error()
				r.UpdatedAt = time.Now()
			}
		} else if resp != nil && resp.Accepted {
			g.statusLabel.SetText("Retry complete")
			transfer.DeleteTransferState(id)
			if r, ok := g.transferRecords[id]; ok {
				r.Status = transfer.StatusCompleted
				r.BytesReceived = r.TotalBytes
				r.UpdatedAt = time.Now()
			}
		}
		g.transfersList.Refresh()
	}()
}

func (g *guiApp) resetTransferUI() {
	g.sendBtn.Enable()
	g.progressBar.SetValue(0)
	g.progressBar.Hide()
	g.progressLabel.Hide()
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

func main() {
	g := newGUIApp()
	g.window.ShowAndRun()
}
