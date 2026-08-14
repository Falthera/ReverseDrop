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
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Falthera/ReverseDrop/internal/app"
	"github.com/Falthera/ReverseDrop/internal/discovery"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/platform"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
)

var (
	timeoutFlag = flag.Duration("timeout", 30*time.Second, "Scan timeout duration")
	targetFlag  = flag.String("target", "", "Target device address")
	verboseFlag = flag.Bool("verbose", false, "Enable verbose (DEBUG) logging")
)

func main() {
	flag.Parse()

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	if *verboseFlag {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	slog.SetDefault(slog.New(handler))

	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pl := platform.OS()
	slog.Info("platform detected", "os", pl.OS, "arch", pl.Arch, "go", pl.Version)

	caps := app.NewCapabilitySet()
	capReporter := platform.NewDefaultReporter()
	if avail, detail := capReporter.BluetoothAvailable(); avail {
		caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: app.CapabilityAvailable, Detail: detail})
	} else {
		caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: app.CapabilityUnavailable, Detail: detail})
	}

	if info, ok := caps.Get(app.CapabilityBluetooth); ok && info.Status == app.CapabilityAvailable {
		return runScan(ctx, caps)
	}

	slog.Warn("bluetooth unavailable on this platform")
	fmt.Println("Bluetooth unavailable on this platform")
	return nil
}

func runScan(ctx context.Context, caps *app.CapabilitySet) error {
	var scanner ble.Scanner
	scanner = ble.NewFakeScanner(nil, 500*time.Millisecond)

	reg := peer.NewRegistry()
	regAdapter := app.NewPeerRegistryAdapter(reg)
	_, err := app.NewService(scanner)
	if err != nil {
		return err
	}

	mgr := discovery.NewManager(scanner, regAdapter, caps)
	if err := mgr.Start(); err != nil {
		return err
	}
	defer mgr.Stop()

	scanCtx, cancel := context.WithTimeout(ctx, *timeoutFlag)
	defer cancel()

	slog.Info("scanning for devices", "timeout", *timeoutFlag)

	fmt.Println("Scanning for nearby devices...")
	fmt.Println("Press Ctrl+C to stop")

	go func() {
		for evt := range reg.Events() {
			switch evt.Type {
			case peer.EventPeerUpserted, peer.EventPeerUpdated:
				p := evt.Peer
				fmt.Printf("Found: %s (%s) RSSI: %d dBm\n", p.DeviceName, p.Address, p.RSSI)
			case peer.EventPeerRemoved:
				p := evt.Peer
				fmt.Printf("Lost: %s (%s)\n", p.DeviceName, p.Address)
			}
		}
	}()

	<-scanCtx.Done()
	fmt.Println("Scan complete.")
	return nil
}
