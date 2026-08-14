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
package parser

import (
	"testing"

	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
)

func TestClassifyAppleAdv(t *testing.T) {
	adv := ble.Advertisement{
		Address:          "AA:BB:CC:DD:EE:FF",
		RSSI:             -50,
		LocalName:        "MacBook",
		ManufacturerData: map[uint16][]byte{0x004C: {0x01, 0x02}},
		ServiceData:      map[string][]byte{"1234": {0x01, 0x02}},
		Timestamp:        0,
	}

	info, err := ClassifyAdv(adv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil AirDropInfo")
	}
	if !info.IsAirDrop {
		t.Fatal("expected IsAirDrop to be true")
	}
	if len(info.RawServiceData) != 1 {
		t.Fatalf("expected 1 service data entry, got %d", len(info.RawServiceData))
	}
}

func TestClassifyNonAppleAdv(t *testing.T) {
	adv := ble.Advertisement{
		Address:          "AA:BB:CC:DD:EE:FF",
		RSSI:             -50,
		ManufacturerData: map[uint16][]byte{0x1234: {0x01, 0x02}},
	}

	info, err := ClassifyAdv(adv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Fatalf("expected nil for non-Apple device, got %+v", info)
	}
}

func TestClassifyEmptyAdv(t *testing.T) {
	adv := ble.Advertisement{}
	info, err := ClassifyAdv(adv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info != nil {
		t.Fatalf("expected nil for empty advertisement, got %+v", info)
	}
}

func TestIsAppleManufacturer(t *testing.T) {
	cases := []struct {
		name       string
		manuf      map[uint16][]byte
		expectTrue bool
	}{
		{"apple", map[uint16][]byte{0x004C: {0x01}}, true},
		{"other", map[uint16][]byte{0x1234: {0x01}}, false},
		{"empty", map[uint16][]byte{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			adv := ble.Advertisement{ManufacturerData: c.manuf}
			got := isAppleManufacturer(adv)
			if got != c.expectTrue {
				t.Errorf("isAppleManufacturer() = %v, want %v", got, c.expectTrue)
			}
		})
	}
}
