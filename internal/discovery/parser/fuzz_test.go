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

func FuzzClassifyAdv(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x01, 0x02})
	f.Add([]byte{0x4C, 0x00, 0x01, 0x02})
	f.Fuzz(func(t *testing.T, data []byte) {
		adv := ble.Advertisement{
			Address:          "AA:BB:CC:DD:EE:FF",
			RSSI:             -50,
			ManufacturerData: map[uint16][]byte{0x004C: data},
			ServiceData:      map[string][]byte{"1234": data},
		}
		_, _ = ClassifyAdv(adv)
	})
}
