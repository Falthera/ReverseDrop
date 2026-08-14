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
	"bytes"
	"context"
	"flag"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Falthera/ReverseDrop/internal/app"
	"github.com/Falthera/ReverseDrop/internal/discovery"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
	"github.com/Falthera/ReverseDrop/internal/platform"
	"github.com/Falthera/ReverseDrop/internal/protocol/peer"
)

func TestMainFunction_RunsWithoutHardware(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"reversedrop", "--timeout", "100ms"}

	var timeoutFlagVal time.Duration
	var verboseFlagVal bool
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	timeoutFlag = &timeoutFlagVal
	verboseFlag = &verboseFlagVal

	_ = run()
}

func TestRunScan_FakeDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
	slog.SetDefault(slog.New(handler))

	caps := app.NewCapabilitySet()
	capReporter := platform.NewDefaultReporter()
	if avail, detail := capReporter.BluetoothAvailable(); avail {
		caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: app.CapabilityAvailable, Detail: detail})
	} else {
		caps.Set(app.CapabilityInfo{Name: app.CapabilityBluetooth, Status: app.CapabilityUnavailable, Detail: detail})
	}

	if info, ok := caps.Get(app.CapabilityBluetooth); ok && info.Status == app.CapabilityAvailable {
		events := []ble.Advertisement{
			{Address: "AA:BB:CC:DD:EE:FF", RSSI: -50, LocalName: "TestDevice", Timestamp: time.Now().Unix()},
		}
		scanner := ble.NewFakeScanner(events, 10*time.Millisecond)
		reg := peer.NewRegistry()
		regAdapter := app.NewPeerRegistryAdapter(reg)
		svc, err := app.NewService(scanner)
		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}
		_ = svc

		mgr := discovery.NewManager(scanner, regAdapter, caps)
		if err := mgr.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		time.Sleep(30 * time.Millisecond)
		mgr.Stop()
	}

	_ = buf
}
