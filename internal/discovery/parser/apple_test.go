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
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/Falthera/ReverseDrop/internal/config"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
)

func buildAirDropPayload(appleID, phone, email, email2 string) []byte {
	payload := make([]byte, AirDropPayloadSize)
	for i := range payload {
		payload[i] = 0x00
	}
	payload[AirDropVersionOffset] = AirDropVersion

	if appleID != "" {
		sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(appleID))))
		copy(payload[AirDropAppleIDOffset:], sum[:HashSize])
	}
	if phone != "" {
		sum := sha256.Sum256([]byte(strings.TrimSpace(phone)))
		copy(payload[AirDropPhoneOffset:], sum[:HashSize])
	}
	if email != "" {
		sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
		copy(payload[AirDropEmailOffset:], sum[:HashSize])
	}
	if email2 != "" {
		sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email2))))
		copy(payload[AirDropEmail2Offset:], sum[:HashSize])
	}

	return payload
}

func buildAirDropAdv(appleID, phone, email, email2 string) ble.Advertisement {
	payload := buildAirDropPayload(appleID, phone, email, email2)
	manufacturerData := make([]byte, 0, 1+len(payload))
	manufacturerData = append(manufacturerData, AirDropRecordType)
	manufacturerData = append(manufacturerData, payload...)

	return ble.Advertisement{
		Address:          "AA:BB:CC:DD:EE:FF",
		RSSI:             -50,
		ManufacturerData: map[uint16][]byte{AppleCompanyID: manufacturerData},
		ServiceData:      map[string][]byte{},
		Timestamp:        0,
	}
}

func TestClassifyAirDropAdv(t *testing.T) {
	appleID := "test@example.com"
	phone := "+1234567890"
	email := "test@example.com"
	email2 := "test2@example.com"

	adv := buildAirDropAdv(appleID, phone, email, email2)

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
	if info.Version != AirDropVersion {
		t.Fatalf("expected version %d, got %d", AirDropVersion, info.Version)
	}
	if len(info.Hashes) != HashSize*4 {
		t.Fatalf("expected %d hash bytes, got %d", HashSize*4, len(info.Hashes))
	}

	payload := buildAirDropPayload(appleID, phone, email, email2)
	expectedHashes := make([]byte, 0, HashSize*4)
	expectedHashes = append(expectedHashes, payload[AirDropAppleIDOffset:AirDropAppleIDOffset+HashSize]...)
	expectedHashes = append(expectedHashes, payload[AirDropPhoneOffset:AirDropPhoneOffset+HashSize]...)
	expectedHashes = append(expectedHashes, payload[AirDropEmailOffset:AirDropEmailOffset+HashSize]...)
	expectedHashes = append(expectedHashes, payload[AirDropEmail2Offset:AirDropEmail2Offset+HashSize]...)

	for i, b := range expectedHashes {
		if info.Hashes[i] != b {
			t.Fatalf("hash mismatch at byte %d: got %02x, want %02x", i, info.Hashes[i], b)
		}
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

func TestClassifyAppleNonAirDropAdv(t *testing.T) {
	adv := ble.Advertisement{
		Address:          "AA:BB:CC:DD:EE:FF",
		RSSI:             -50,
		ManufacturerData: map[uint16][]byte{0x004C: {0x01, 0x02}},
	}

	info, err := ClassifyAdv(adv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil AirDropInfo")
	}
	if info.IsAirDrop {
		t.Fatal("expected IsAirDrop to be false for non-AirDrop Apple data")
	}
}

func TestClassifyShortManufacturerData(t *testing.T) {
	adv := ble.Advertisement{
		Address:          "AA:BB:CC:DD:EE:FF",
		RSSI:             -50,
		ManufacturerData: map[uint16][]byte{0x004C: {0x05}},
	}

	_, err := ClassifyAdv(adv)
	if err == nil {
		t.Fatal("expected error for short manufacturer data")
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

func TestGetDeviceID(t *testing.T) {
	appleID := "test@example.com"
	phone := "+1234567890"
	email := "test@example.com"
	email2 := "test2@example.com"

	adv := buildAirDropAdv(appleID, phone, email, email2)
	deviceID := GetDeviceID(adv)

	if deviceID == adv.Address {
		t.Fatal("expected device ID to be derived from hashes, not address")
	}
	if len(deviceID) != 16 {
		t.Fatalf("expected device ID length 16, got %d", len(deviceID))
	}

	nonAppleAdv := ble.Advertisement{
		Address:          "AA:BB:CC:DD:EE:FF",
		ManufacturerData: map[uint16][]byte{},
	}
	deviceID = GetDeviceID(nonAppleAdv)
	if deviceID != nonAppleAdv.Address {
		t.Fatalf("expected device ID to be address for non-Apple device, got %s", deviceID)
	}
}

func TestCreateAirDropAdvertisement(t *testing.T) {
	appleID := "test@example.com"
	phone := "+1234567890"
	email := "test@example.com"
	email2 := "test2@example.com"

	adv := CreateAirDropAdvertisement(appleID, phone, email, email2, config.DiscoveryModeContactsOnly)

	if adv.ManufacturerData == nil {
		t.Fatal("expected ManufacturerData to be set")
	}
	data, ok := adv.ManufacturerData[AppleCompanyID]
	if !ok {
		t.Fatal("expected Apple company ID in ManufacturerData")
	}
	if len(data) != 1+AirDropPayloadSize {
		t.Fatalf("expected manufacturer data length %d, got %d", 1+AirDropPayloadSize, len(data))
	}
	if data[0] != AirDropRecordType {
		t.Fatalf("expected record type %02x, got %02x", AirDropRecordType, data[0])
	}
	payload := data[1:]
	if payload[AirDropVersionOffset] != AirDropVersion {
		t.Fatalf("expected version %d, got %d", AirDropVersion, payload[AirDropVersionOffset])
	}

	info, err := ClassifyAdv(adv)
	if err != nil {
		t.Fatalf("unexpected error parsing created advertisement: %v", err)
	}
	if !info.IsAirDrop {
		t.Fatal("expected created advertisement to be classified as AirDrop")
	}
	if info.Version != AirDropVersion {
		t.Fatalf("expected version %d, got %d", AirDropVersion, info.Version)
	}
}

func TestCreateAirDropAdvertisementEmptyStrings(t *testing.T) {
	adv := CreateAirDropAdvertisement("", "", "", "", config.DiscoveryModeContactsOnly)

	if adv.ManufacturerData == nil {
		t.Fatal("expected ManufacturerData to be set")
	}
	data, ok := adv.ManufacturerData[AppleCompanyID]
	if !ok {
		t.Fatal("expected Apple company ID in ManufacturerData")
	}
	if len(data) != 1+AirDropPayloadSize {
		t.Fatalf("expected manufacturer data length %d, got %d", 1+AirDropPayloadSize, len(data))
	}

	info, err := ClassifyAdv(adv)
	if err != nil {
		t.Fatalf("unexpected error parsing created advertisement: %v", err)
	}
	if !info.IsAirDrop {
		t.Fatal("expected created advertisement to be classified as AirDrop")
	}
	for i, b := range info.Hashes {
		if b != 0x00 {
			t.Fatalf("expected zero hash at byte %d, got %02x", i, b)
		}
	}
}
