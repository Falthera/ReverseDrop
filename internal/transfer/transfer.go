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
	"log/slog"
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
		return fmt.Errorf("failed to ensure certificate: %w", err)
	}
	cert, err := tls.LoadX509KeyPair(m.certPath, m.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"reversedrop"},
		MinVersion:   tls.VersionTLS12,
	}
	config.PreferServerCipherSuites = true
	addr := fmt.Sprintf(":%d", m.port)
	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to start TLS listener: %w", err)
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
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate certificate serial: %w", err)
	}
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"ReverseDrop"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err := os.WriteFile(m.certPath, certPEM, 0o600); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}
	if err := os.WriteFile(m.keyPath, keyPEM, 0o600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}
	return nil
}

func generateCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"ReverseDrop"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return tls.X509KeyPair(certPEM, keyPEM)
}

func (m *Manager) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		slog.Warn("failed to read request line", "error", err)
		return
	}
	var req TransferRequest
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &req); err != nil {
		slog.Warn("failed to parse transfer request", "error", err)
		return
	}
	safeName := filepath.Base(req.FileName)
	savePath := filepath.Join(m.downloads, safeName)
	resp := TransferResponse{ID: req.ID, Accepted: true, SavePath: savePath}
	data, _ := json.Marshal(resp)
	if _, err := fmt.Fprintf(conn, "%s\n", data); err != nil {
		slog.Warn("failed to send response", "error", err)
		return
	}
	if err := m.receiveFile(ctx, conn, reader, resp.SavePath, req.FileSize, nil); err != nil {
		slog.Warn("failed to receive file", "error", err)
		return
	}
}

func (m *Manager) receiveFile(ctx context.Context, conn net.Conn, reader *bufio.Reader, dest string, size int64, onProgress ProgressFunc) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer f.Close()

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				conn.SetDeadline(time.Now().Add(30 * time.Second))
			}
		}
	}()

	var written int64
	buf := make([]byte, 64*1024)
	for written < size {
		n, err := reader.Read(buf)
		if n > 0 {
			totalWritten := 0
			for totalWritten < n {
				w, wErr := f.Write(buf[totalWritten:n])
				if w > 0 {
					totalWritten += w
				}
				if wErr != nil {
					close(done)
					return fmt.Errorf("failed to write to file at offset %d: %w", written+int64(totalWritten), wErr)
				}
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
			close(done)
			return fmt.Errorf("failed to read from connection at offset %d: %w", written, err)
		}
	}
	close(done)
	return nil
}

func SendFile(ctx context.Context, addr string, req TransferRequest, onProgress ProgressFunc) (*TransferResponse, error) {
	cert, err := generateCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"reversedrop"},
	}
	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TLS connection: %w", err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				conn.SetDeadline(time.Now().Add(30 * time.Second))
			}
		}
	}()

	data, _ := json.Marshal(req)
	if _, err := fmt.Fprintf(conn, "%s\n", data); err != nil {
		close(done)
		return nil, fmt.Errorf("failed to send transfer request: %w", err)
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		close(done)
		return nil, fmt.Errorf("failed to read transfer response: %w", err)
	}
	var resp TransferResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &resp); err != nil {
		close(done)
		return nil, fmt.Errorf("failed to parse transfer response: %w", err)
	}
	if !resp.Accepted {
		close(done)
		return &resp, fmt.Errorf("transfer rejected: %s", resp.Error)
	}
	f, err := os.Open(req.FileName)
	if err != nil {
		close(done)
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		close(done)
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	buf := make([]byte, 64*1024)
	var sent int64
	for sent < stat.Size() {
		n, rErr := f.Read(buf)
		if n > 0 {
			totalWritten := 0
			for totalWritten < n {
				w, wErr := conn.Write(buf[totalWritten:n])
				if w > 0 {
					totalWritten += w
				}
				if wErr != nil {
					close(done)
					return &resp, fmt.Errorf("failed to write data at offset %d: %w", sent+int64(totalWritten), wErr)
				}
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
			close(done)
			return &resp, fmt.Errorf("failed to read file at offset %d: %w", sent, rErr)
		}
	}
	close(done)
	return &resp, nil
}
