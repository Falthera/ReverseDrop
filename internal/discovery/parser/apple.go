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

	"github.com/Falthera/ReverseDrop/internal/discovery/ble"
)

const (
	AppleCompanyID uint16 = 0x004C
)

type AirDropInfo struct {
	AppleModel     string
	RecordID       string
	IsAirDrop      bool
	RawServiceData map[string][]byte
}

func ClassifyAdv(adv ble.Advertisement) (*AirDropInfo, error) {
	if !isAppleManufacturer(adv) {
		return nil, nil
	}
	info := &AirDropInfo{RawServiceData: make(map[string][]byte)}
	for k, v := range adv.ServiceData {
		info.RawServiceData[k] = v
	}
	if len(adv.ServiceData) == 0 {
		return info, nil
	}
	info.IsAirDrop = true
	for _, data := range adv.ServiceData {
		if len(data) >= 2 {
			info.RecordID = fmt.Sprintf("%x", data)
		}
	}
	return info, nil
}

func isAppleManufacturer(adv ble.Advertisement) bool {
	for id := range adv.ManufacturerData {
		if id == AppleCompanyID {
			return true
		}
	}
	return false
}
