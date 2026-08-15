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
	"time"

	"fyne.io/fyne/v2"
	fapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/Falthera/ReverseDrop/internal/app"
	"github.com/Falthera/ReverseDrop/internal/discovery"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/platform"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
	"github.com/Falthera/ReverseDrop/internal/transfer"
)

type guiApp struct {
	fyneApp       fyne.App
	window        fyne.Window
	service       *app.Service
	regAdapter    *app.PeerRegistryAdapter
	mgr           *discovery.Manager
	transferMgr   *transfer.Manager
	peersList     *widget.List
	statusLabel   *widget.Label
	progressBar   *widget.ProgressBar
	progressLabel *widget.Label
	scanBtn       *widget.Button
	stopBtn       *widget.Button
	sendBtn       *widget.Button
}

func newGUIApp() *guiApp {
	g := &guiApp{regAdapter: app.NewPeerRegistryAdapter(peer.NewRegistry())}
	g.fyneApp = fapp.NewWithID("com.falthera.reversedrop")
	g.window = g.fyneApp.NewWindow("ReverseDrop")
	g.window.Resize(fyne.NewSize(900, 600))
	g.transferMgr = transfer.NewManager(transfer.DefaultPort, "")
	g.window.SetContent(g.buildUI())
	return g
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
			return
		}
		g.statusLabel.SetText(fmt.Sprintf("Selected: %s (%s)", peers[id].DeviceName, peers[id].Address))
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
		toolbar,
	)

	content := container.NewBorder(header, nil, nil, nil, g.peersList)

	return content
}

func (g *guiApp) startScan() {
	caps := app.NewCapabilitySet()
	capReporter := platform.NewDefaultReporter()
	if avail, detail := capReporter.BluetoothAvailable(); avail {
		caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: app.CapabilityAvailable, Detail: detail})
	} else {
		caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: app.CapabilityUnavailable, Detail: detail})
	}

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
	g.mgr = mgr

	g.scanBtn.Disable()
	g.stopBtn.Enable()
	g.statusLabel.SetText("Scanning...")
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
	peers := g.regAdapter.PeerRegistry.List()
	if len(peers) == 0 {
		g.statusLabel.SetText("No peers available")
		return
	}
	selected := g.peersList.Selected()
	if selected < 0 || selected >= len(peers) {
		g.statusLabel.SetText("Select a peer first")
		return
	}
	peer := peers[selected]
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

func (g *guiApp) startTransfer(path string, peer peer.Peer) {
	g.sendBtn.Disable()
	g.progressBar.Show()
	g.progressLabel.Show()
	g.progressBar.SetValue(0)
	g.progressLabel.SetText(fmt.Sprintf("Sending %s...", filepath.Base(path)))

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
		resp, err := transfer.SendFile(context.Background(), addr, req, func(sent, total int64) {
			if total > 0 {
				g.progressBar.SetValue(float64(sent) / float64(total))
			}
			g.progressLabel.SetText(fmt.Sprintf("Sending %s... %d/%d", filepath.Base(path), sent, total))
		})
		if err != nil {
			g.statusLabel.SetText("Transfer failed: " + err.Error())
		} else if resp != nil && resp.Accepted {
			g.statusLabel.SetText("Transfer complete")
		} else if resp != nil {
			g.statusLabel.SetText("Transfer rejected: " + resp.Error)
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
