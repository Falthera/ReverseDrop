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
	"fmt"
	"strings"

	"github.com/Falthera/ReverseDrop/internal/config"
	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
)

const (
	AppleCompanyID uint16 = 0x004C

	AirDropRecordType = 0x05
	AirDropVersion    = 0x01

	AirDropPayloadSize   = 18
	AirDropPaddingSize   = 8
	AirDropVersionOffset = 8

	AirDropAppleIDOffset = 9
	AirDropPhoneOffset   = 11
	AirDropEmailOffset   = 13
	AirDropEmail2Offset  = 15
	AirDropSuffixOffset  = 17

	HashSize = 2
)

type AirDropInfo struct {
	Hashes         []byte
	Version        uint8
	IsAirDrop      bool
	AppleModel     string
	Flags          string
	EMailHash      string
	PhoneHash      string
	DeviceName     string
	RawServiceData map[string][]byte
}

type AirDropRecord struct {
	Type        uint8
	Version     uint8
	ContactName string
	Email       string
	Phone       string
}

func ClassifyAdv(adv ble.Advertisement) (*AirDropInfo, error) {
	manufacturerData, ok := adv.ManufacturerData[AppleCompanyID]
	if !ok {
		return nil, nil
	}

	if len(manufacturerData) < 1 {
		return &AirDropInfo{IsAirDrop: false}, nil
	}

	recordType := manufacturerData[0]
	if recordType != AirDropRecordType {
		return &AirDropInfo{IsAirDrop: false}, nil
	}

	if len(manufacturerData) < 1+AirDropPayloadSize {
		return nil, fmt.Errorf("manufacturer data too short for AirDrop payload")
	}

	payload := manufacturerData[1 : 1+AirDropPayloadSize]
	if len(payload) != AirDropPayloadSize {
		return nil, fmt.Errorf("invalid AirDrop payload size")
	}

	info := &AirDropInfo{IsAirDrop: true}

	version := payload[AirDropVersionOffset]
	if version != AirDropVersion {
		return nil, fmt.Errorf("unsupported AirDrop version: %d", version)
	}
	info.Version = version

	info.Hashes = make([]byte, 0, HashSize*4)
	info.Hashes = append(info.Hashes, payload[AirDropAppleIDOffset:AirDropAppleIDOffset+HashSize]...)
	info.Hashes = append(info.Hashes, payload[AirDropPhoneOffset:AirDropPhoneOffset+HashSize]...)
	info.Hashes = append(info.Hashes, payload[AirDropEmailOffset:AirDropEmailOffset+HashSize]...)
	info.Hashes = append(info.Hashes, payload[AirDropEmail2Offset:AirDropEmail2Offset+HashSize]...)

	return info, nil
}

func isAppleManufacturer(adv ble.Advertisement) bool {
	_, ok := adv.ManufacturerData[AppleCompanyID]
	return ok
}

func GetDeviceID(adv ble.Advertisement) string {
	manufacturerData, ok := adv.ManufacturerData[AppleCompanyID]
	if !ok || len(manufacturerData) < 1+AirDropPayloadSize {
		return adv.Address
	}

	payload := manufacturerData[1 : 1+AirDropPayloadSize]
	hashes := make([]byte, 0, HashSize*4)
	hashes = append(hashes, payload[AirDropAppleIDOffset:AirDropAppleIDOffset+HashSize]...)
	hashes = append(hashes, payload[AirDropPhoneOffset:AirDropPhoneOffset+HashSize]...)
	hashes = append(hashes, payload[AirDropEmailOffset:AirDropEmailOffset+HashSize]...)
	hashes = append(hashes, payload[AirDropEmail2Offset:AirDropEmail2Offset+HashSize]...)

	hash := sha256.Sum256(hashes)
	return fmt.Sprintf("%x", hash[:8])
}

func CreateAirDropAdvertisement(appleID, phone, email, email2 string, mode config.DiscoveryMode) ble.Advertisement {
	hashes := make([]byte, 0, HashSize*4)

	if mode == config.DiscoveryModeContactsOnly {
		if appleID != "" {
			sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(appleID))))
			hashes = append(hashes, sum[:HashSize]...)
		} else {
			hashes = append(hashes, make([]byte, HashSize)...)
		}

		if phone != "" {
			sum := sha256.Sum256([]byte(strings.TrimSpace(phone)))
			hashes = append(hashes, sum[:HashSize]...)
		} else {
			hashes = append(hashes, make([]byte, HashSize)...)
		}

		if email != "" {
			sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
			hashes = append(hashes, sum[:HashSize]...)
		} else {
			hashes = append(hashes, make([]byte, HashSize)...)
		}

		if email2 != "" {
			sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email2))))
			hashes = append(hashes, sum[:HashSize]...)
		} else {
			hashes = append(hashes, make([]byte, HashSize)...)
		}
	} else {
		hashes = append(hashes, make([]byte, HashSize*4)...)
	}

	payload := make([]byte, AirDropPayloadSize)
	for i := range payload {
		payload[i] = 0x00
	}
	payload[AirDropVersionOffset] = AirDropVersion
	copy(payload[AirDropAppleIDOffset:], hashes[0:HashSize])
	copy(payload[AirDropPhoneOffset:], hashes[HashSize:HashSize*2])
	copy(payload[AirDropEmailOffset:], hashes[HashSize*2:HashSize*3])
	copy(payload[AirDropEmail2Offset:], hashes[HashSize*3:HashSize*4])

	manufacturerData := make([]byte, 0, 1+AirDropPayloadSize)
	manufacturerData = append(manufacturerData, AirDropRecordType)
	manufacturerData = append(manufacturerData, payload...)

	return ble.Advertisement{
		Address:          "",
		RSSI:             0,
		LocalName:        "",
		ManufacturerData: map[uint16][]byte{AppleCompanyID: manufacturerData},
		ServiceUUIDs:     []string{},
		ServiceData:      map[string][]byte{},
		Timestamp:        0,
	}
}
