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
	"fmt"
	"strings"

	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
)

const (
	AppleCompanyID uint16 = 0x004C

	AirDropRecordType = 0x05
	AirDropVersion    = 0x00

	RecordTypeContact = 0x01
	RecordTypeEmail   = 0x02
	RecordTypePhone   = 0x03
)

type AirDropInfo struct {
	AppleModel     string
	RecordID       string
	IsAirDrop      bool
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
	if !isAppleManufacturer(adv) {
		return nil, nil
	}
	info := &AirDropInfo{RawServiceData: make(map[string][]byte)}
	for k, v := range adv.ServiceData {
		info.RawServiceData[k] = v
	}
	records, err := parseAirDropRecords(adv)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		if len(adv.ServiceData) == 0 {
			return info, nil
		}
		info.IsAirDrop = true
		return info, nil
	}
	info.IsAirDrop = true
	for _, rec := range records {
		switch rec.Type {
		case RecordTypeContact:
			info.RecordID = rec.ContactName
			info.AppleModel = rec.ContactName
		case RecordTypeEmail:
			if info.RecordID == "" {
				info.RecordID = rec.Email
			}
		case RecordTypePhone:
			if info.RecordID == "" {
				info.RecordID = rec.Phone
			}
		}
	}
	return info, nil
}

func parseAirDropRecords(adv ble.Advertisement) ([]AirDropRecord, error) {
	var records []AirDropRecord
	for _, data := range adv.ServiceData {
		if len(data) < 2 {
			continue
		}
		recordType := data[0]
		version := data[1]
		if recordType != AirDropRecordType || version != AirDropVersion {
			continue
		}
		payload := data[2:]
		rec, err := decodeAirDropRecord(payload)
		if err != nil {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

func decodeAirDropRecord(payload []byte) (AirDropRecord, error) {
	if len(payload) < 2 {
		return AirDropRecord{}, fmt.Errorf("payload too short")
	}
	recordType := payload[0]
	length := payload[1]
	if int(length) > len(payload)-2 {
		return AirDropRecord{}, fmt.Errorf("invalid record length")
	}
	value := payload[2 : 2+int(length)]
	switch recordType {
	case RecordTypeContact:
		return AirDropRecord{Type: recordType, ContactName: strings.TrimSpace(string(value))}, nil
	case RecordTypeEmail:
		return AirDropRecord{Type: recordType, Email: strings.TrimSpace(string(value))}, nil
	case RecordTypePhone:
		return AirDropRecord{Type: recordType, Phone: strings.TrimSpace(string(value))}, nil
	default:
		return AirDropRecord{Type: recordType}, nil
	}
}

func isAppleManufacturer(adv ble.Advertisement) bool {
	for id := range adv.ManufacturerData {
		if id == AppleCompanyID {
			return true
		}
	}
	return false
}

func GetDeviceID(adv ble.Advertisement) string {
	for _, data := range adv.ManufacturerData {
		if len(data) >= 8 {
			return fmt.Sprintf("%x", data[:8])
		}
	}
	return adv.Address
}
