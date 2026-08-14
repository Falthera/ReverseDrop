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
package transfer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultPort = 9999
)

type TransferRequest struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	SenderName  string `json:"sender_name,omitempty"`
}

type TransferResponse struct {
	ID        string `json:"id"`
	Accepted  bool   `json:"accepted"`
	SavePath  string `json:"save_path,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Manager struct {
	port      int
	downloads string
}

func NewManager(port int, downloadsDir string) *Manager {
	if downloadsDir == "" {
		home, _ := os.UserHomeDir()
		downloadsDir = filepath.Join(home, "Downloads", "ReverseDrop")
	}
	return &Manager{
		port:      port,
		downloads: downloadsDir,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", m.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go m.handleConn(ctx, conn)
		}
	}()
	return nil
}

func (m *Manager) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	var req TransferRequest
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &req); err != nil {
		return
	}
	resp := TransferResponse{ID: req.ID, Accepted: true, SavePath: filepath.Join(m.downloads, req.FileName)}
	data, _ := json.Marshal(resp)
	if _, err := fmt.Fprintf(conn, "%s\n", data); err != nil {
		return
	}
	if err := m.receiveFile(ctx, conn, reader, resp.SavePath, req.FileSize); err != nil {
		return
	}
}

func (m *Manager) receiveFile(ctx context.Context, conn net.Conn, reader *bufio.Reader, dest string, size int64) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.CopyN(f, reader, size)
	return err
}

func SendFile(ctx context.Context, addr string, req TransferRequest) (*TransferResponse, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	data, _ := json.Marshal(req)
	if _, err := fmt.Fprintf(conn, "%s\n", data); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	var resp TransferResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &resp); err != nil {
		return nil, err
	}
	if !resp.Accepted {
		return &resp, fmt.Errorf("transfer rejected: %s", resp.Error)
	}
	f, err := os.Open(req.FileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, _ := f.Stat()
	_, err = io.CopyN(conn, f, stat.Size())
	return &resp, err
}
