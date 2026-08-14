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
package ble

import (
	"context"
	"fmt"
	"runtime"
)

type realScanner struct{}

func NewRealScanner() Scanner {
	return &realScanner{}
}

func (r *realScanner) Scan(ctx context.Context) (<-chan Advertisement, error) {
	return nil, fmt.Errorf("real BLE scanner not yet implemented for %s", runtime.GOOS)
}

func (r *realScanner) Stop() error {
	return nil
}
