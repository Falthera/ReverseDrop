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
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultPort    = 9999
	certFileName   = "reversedrop.crt"
	keyFileName    = "reversedrop.key"
)

type TransferRequest struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	SenderName  string `json:"sender_name,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
}

type TransferResponse struct {
	ID        string `json:"id"`
	Accepted  bool   `json:"accepted"`
	SavePath  string `json:"save_path,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Manager struct {
	port        int
	downloads   string
	certPath    string
	keyPath     string
}

type ProgressFunc func(bytesSent int64, total int64)

func NewManager(port int, downloadsDir string) *Manager {
	if downloadsDir == "" {
		home, _ := os.UserHomeDir()
		downloadsDir = filepath.Join(home, "Downloads", "ReverseDrop")
	}
	cfgDir, _ := os.UserConfigDir()
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	certDir := filepath.Join(cfgDir, "reversedrop")
	return &Manager{
		port:      port,
		downloads: downloadsDir,
		certPath:  filepath.Join(certDir, certFileName),
		keyPath:   filepath.Join(certDir, keyFileName),
	}
}

func (m *Manager) Start(ctx context.Context) error {
	if err := m.ensureCert(); err != nil {
		return err
	}
	cert, err := tls.LoadX509KeyPair(m.certPath, m.keyPath)
	if err != nil {
		return err
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"reversedrop"},
	}
	addr := fmt.Sprintf(":%d", m.port)
	ln, err := tls.Listen("tcp", addr, config)
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

func (m *Manager) ensureCert() error {
	if _, err := os.Stat(m.certPath); err == nil {
		if _, err := os.Stat(m.keyPath); err == nil {
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(m.certPath), 0o700); err != nil {
		return err
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ReverseDrop"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err := os.WriteFile(m.certPath, certPEM, 0o600); err != nil {
		return err
	}
	return os.WriteFile(m.keyPath, keyPEM, 0o600)
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
	if err := m.receiveFile(ctx, conn, reader, resp.SavePath, req.FileSize, nil); err != nil {
		return
	}
}

func (m *Manager) receiveFile(ctx context.Context, conn net.Conn, reader *bufio.Reader, dest string, size int64, onProgress ProgressFunc) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	var written int64
	buf := make([]byte, 64*1024)
	for written < size {
		n, err := reader.Read(buf)
		if n > 0 {
			if _, wErr := f.Write(buf[:n]); wErr != nil {
				return wErr
			}
			written += int64(n)
			if onProgress != nil {
				onProgress(written, size)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

func SendFile(ctx context.Context, addr string, req TransferRequest, onProgress ProgressFunc) (*TransferResponse, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"reversedrop"}})
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
	buf := make([]byte, 64*1024)
	var sent int64
	for sent < stat.Size() {
		n, rErr := f.Read(buf)
		if n > 0 {
			if _, wErr := conn.Write(buf[:n]); wErr != nil {
				return &resp, wErr
			}
			sent += int64(n)
			if onProgress != nil {
				onProgress(sent, stat.Size())
			}
		}
		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return &resp, rErr
		}
	}
	return &resp, nil
}
