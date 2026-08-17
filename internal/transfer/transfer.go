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
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"howett.net/plist"

	"github.com/Falthera/ReverseDrop/internal/trust"
)

const (
	DefaultPort      = 8770
	certFileName     = "reversedrop.crt"
	keyFileName      = "reversedrop.key"
	defaultChunkSize = 4 * 1024 * 1024
)

// TransferRequest is the public request type kept for GUI compatibility.
type TransferRequest struct {
	ID          string   `json:"id"`
	FileName    string   `json:"file_name,omitempty"`
	FileSize    int64    `json:"file_size,omitempty"`
	FileNames   []string `json:"file_names,omitempty"`
	FileSizes   []int64  `json:"file_sizes,omitempty"`
	SenderName  string   `json:"sender_name,omitempty"`
	MimeType    string   `json:"mime_type,omitempty"`
}

// TransferResponse is the public response type kept for GUI compatibility.
type TransferResponse struct {
	ID        string `json:"id"`
	Accepted  bool   `json:"accepted"`
	SavePath  string `json:"save_path,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TransferStatus represents the state of a transfer.
type TransferStatus string

const (
	StatusPending   TransferStatus = "pending"
	StatusActive    TransferStatus = "active"
	StatusPaused    TransferStatus = "paused"
	StatusCompleted TransferStatus = "completed"
	StatusFailed    TransferStatus = "failed"
)

// TransferRecord tracks the progress and state of a single transfer.
type TransferRecord struct {
	ID            string        `json:"id"`
	FileName      string        `json:"file_name"`
	TotalBytes    int64         `json:"total_bytes"`
	BytesReceived int64         `json:"bytes_received"`
	Status        TransferStatus `json:"status"`
	PeerAddr      string        `json:"peer_addr"`
	SavePath      string        `json:"save_path,omitempty"`
	Error         string        `json:"error,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Chunks        []ChunkInfo   `json:"chunks,omitempty"`
}

// ChunkInfo describes a received chunk of a transfer.
type ChunkInfo struct {
	Offset int64 `json:"offset"`
	Length int64 `json:"length"`
}

// TransferStore persists transfer records to disk.
type TransferStore interface {
	Save(record *TransferRecord) error
	Load(id string) (*TransferRecord, error)
	List() ([]*TransferRecord, error)
	Delete(id string) error
	PartialPath(id string) string
}

// FileTransferStore implements TransferStore using JSON files.
type FileTransferStore struct {
	baseDir string
}

// NewFileTransferStore creates a new file-based transfer store.
func NewFileTransferStore(baseDir string) (*FileTransferStore, error) {
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create transfer store directory: %w", err)
	}
	return &FileTransferStore{baseDir: baseDir}, nil
}

func (s *FileTransferStore) recordPath(id string) string {
	return filepath.Join(s.baseDir, id+".json")
}

func (s *FileTransferStore) PartialPath(id string) string {
	return filepath.Join(s.baseDir, id+".partial")
}

func (s *FileTransferStore) Save(record *TransferRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.recordPath(record.ID), data, 0o600)
}

func (s *FileTransferStore) Load(id string) (*TransferRecord, error) {
	data, err := os.ReadFile(s.recordPath(id))
	if err != nil {
		return nil, err
	}
	var r TransferRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *FileTransferStore) List() ([]*TransferRecord, error) {
	files, err := filepath.Glob(filepath.Join(s.baseDir, "*.json"))
	if err != nil {
		return nil, err
	}
	var records []*TransferRecord
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var r TransferRecord
		if err := json.Unmarshal(data, &r); err != nil {
			continue
		}
		records = append(records, &r)
	}
	return records, nil
}

func (s *FileTransferStore) Delete(id string) error {
	_ = os.Remove(s.recordPath(id))
	_ = os.Remove(s.PartialPath(id))
	return nil
}

type Manager struct {
	port         int
	downloads    string
	certPath     string
	keyPath      string
	server       *http.Server
	askCallback  func(*AskRequest) *AskResponse
	autoAcceptFn func(string) bool
	trustStore   *trust.Store
	discoveries  map[net.Conn]bool
	asks         map[net.Conn]bool
	uploads      map[net.Conn]bool
	mu           sync.Mutex
	uploadSem    chan struct{}

	transfers    map[string]*TransferRecord
	transferMu   sync.RWMutex
	store        TransferStore
	maxParallel  int
	onUpdate     func(*TransferRecord)
}

type ProgressFunc func(state string, bytesSent int64, total int64)

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
	transfersDir := filepath.Join(cfgDir, "reversedrop", "transfers")
	store, _ := NewFileTransferStore(transfersDir)
	return &Manager{
		port:        port,
		downloads:   downloadsDir,
		certPath:    filepath.Join(certDir, certFileName),
		keyPath:     filepath.Join(certDir, keyFileName),
		discoveries: make(map[net.Conn]bool),
		asks:        make(map[net.Conn]bool),
		uploads:     make(map[net.Conn]bool),
		uploadSem:   make(chan struct{}, 4),
		transfers:   make(map[string]*TransferRecord),
		store:       store,
		maxParallel: 3,
	}
}

func (m *Manager) SetAutoAccept(enabled bool, store *trust.Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !enabled {
		m.autoAcceptFn = nil
		m.trustStore = nil
		return
	}
	m.trustStore = store
	m.autoAcceptFn = func(senderID string) bool {
		if m.trustStore == nil {
			return false
		}
		return m.trustStore.Get(senderID) == trust.TrustLevelTrusted
	}
}

func (m *Manager) SetMaxParallel(n int) {
	if n <= 0 {
		n = 3
	}
	m.mu.Lock()
	m.maxParallel = n
	m.uploadSem = make(chan struct{}, n)
	m.mu.Unlock()
}

func (m *Manager) SetOnTransferUpdate(cb func(*TransferRecord)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onUpdate = cb
}

func (m *Manager) GetTransfers() []*TransferRecord {
	m.transferMu.RLock()
	defer m.transferMu.RUnlock()
	var result []*TransferRecord
	for _, r := range m.transfers {
		result = append(result, r)
	}
	return result
}

func (m *Manager) GetTransfer(id string) (*TransferRecord, bool) {
	m.transferMu.RLock()
	defer m.transferMu.RUnlock()
	r, ok := m.transfers[id]
	return r, ok
}

func (m *Manager) PauseTransfer(id string) error {
	m.transferMu.Lock()
	defer m.transferMu.Unlock()
	r, ok := m.transfers[id]
	if !ok {
		return fmt.Errorf("transfer not found: %s", id)
	}
	r.Status = StatusPaused
	r.UpdatedAt = time.Now()
	if m.store != nil {
		m.store.Save(r)
	}
	if m.onUpdate != nil {
		m.onUpdate(r)
	}
	return nil
}

func (m *Manager) ResumeTransfer(id string) error {
	m.transferMu.Lock()
	defer m.transferMu.Unlock()
	r, ok := m.transfers[id]
	if !ok {
		return fmt.Errorf("transfer not found: %s", id)
	}
	if r.Status != StatusPaused && r.Status != StatusFailed {
		return fmt.Errorf("transfer cannot be resumed: %s", r.Status)
	}
	r.Status = StatusActive
	r.Error = ""
	r.UpdatedAt = time.Now()
	if m.store != nil {
		m.store.Save(r)
	}
	if m.onUpdate != nil {
		m.onUpdate(r)
	}
	return nil
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
		Certificates:             []tls.Certificate{cert},
		NextProtos:               []string{"airdrop"},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}
	addr := fmt.Sprintf(":%d", m.port)
	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to start TLS listener: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/Discover", m.handleDiscover)
	mux.HandleFunc("/Ask", m.handleAsk)
	mux.HandleFunc("/Upload", m.handleUpload)
	mux.HandleFunc("/UploadStatus", m.handleUploadStatus)
	mux.HandleFunc("/Error", m.handleError)
	m.server = &http.Server{Handler: mux}
	go m.server.Serve(ln)
	go func() {
		<-ctx.Done()
		ln.Close()
		m.server.Close()
	}()
	return nil
}

func (m *Manager) Stop() {
	if m.server != nil {
		m.server.Close()
	}
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
	uuid := fmt.Sprintf("%x", md5.New().Sum(nil))
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"ReverseDrop"},
			CommonName:   fmt.Sprintf("com.apple.idms.appleid.prd.%s", uuid),
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

// AirDrop message types -----------------------------------------------------

type DiscoverRequest struct {
	SenderComputerName string
	SenderModelName    string
	SenderRecordData   []byte
	SenderID           string
	TransferVersion    int
}

type DiscoverResponse struct {
	ReceiverComputerName string
	ReceiverModelName    string
	ReceiverRecordData   []byte
	StatusCode           int
}

type AskRequest struct {
	SenderComputerName string
	SenderModelName    string
	SenderID           string
	TransferID         string
	Files              []FileMeta
}

type AskResponse struct {
	ReceiverComputerName string
	ReceiverModelName    string
	Accepted             bool
	SavePath             string
}

type FileMeta struct {
	FileType           string
	FileName           string
	FileBomPath        string
	FileIcon           []byte
	FileIsDirectory    bool
	ConvertMediaFormats bool
}

// CPIO newc archive --------------------------------------------------------

func writeCpioArchive(files []FileMeta, baseDir string) ([]byte, error) {
	var buf bytes.Buffer
	w := newCpioNewcWriter(&buf)
	for _, f := range files {
		path := filepath.Join(baseDir, f.FileName)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", path, err)
		}
		rec := cpioRecord{
			name:    f.FileBomPath,
			mode:    0644,
			size:    uint64(len(data)),
			mtime:   uint32(time.Now().Unix()),
			data:    data,
		}
		if err := w.WriteRecord(rec); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeCpioArchiveStream(w io.Writer, files []FileMeta, baseDir string) error {
	writer := newCpioNewcWriter(w)
	for _, f := range files {
		path := filepath.Join(baseDir, f.FileName)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		rec := cpioRecord{
			name:    f.FileBomPath,
			mode:    0644,
			size:    uint64(len(data)),
			mtime:   uint32(time.Now().Unix()),
			data:    data,
		}
		if err := writer.WriteRecord(rec); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

func extractCpioArchive(data []byte, destDir string) error {
	r := newCpioNewcReader(bytes.NewReader(data))
	for {
		rec, err := r.ReadRecord()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if rec.name == "TRAILER!!!" {
			return nil
		}
		fullPath := filepath.Join(destDir, rec.name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return err
		}
		if rec.mode&0o170000 == 0o040000 {
			if err := os.MkdirAll(fullPath, os.FileMode(rec.mode&0o777)); err != nil {
				return err
			}
			continue
		}
		if err := os.WriteFile(fullPath, rec.data, os.FileMode(rec.mode&0o777)); err != nil {
			return err
		}
	}
}

type cpioRecord struct {
	name  string
	mode  uint16
	uid   uint16
	gid   uint16
	mtime uint32
	size  uint64
	data  []byte
}

type cpioNewcWriter struct {
	w io.Writer
}

func newCpioNewcWriter(w io.Writer) *cpioNewcWriter {
	return &cpioNewcWriter{w: w}
}

func (w *cpioNewcWriter) WriteRecord(rec cpioRecord) error {
	header := make([]byte, 110)
	copy(header[0:6], []byte("070701"))
	writeOctalField(header[6:14], uint64(rec.mtime))
	writeOctalField(header[14:22], uint64(rec.mode)|0o100000)
	writeOctalField(header[22:30], uint64(rec.uid))
	writeOctalField(header[30:38], uint64(rec.gid))
	writeOctalField(header[38:46], 1)
	writeOctalField(header[46:54], uint64(len(rec.data)))
	writeOctalField(header[54:62], 0)
	writeOctalField(header[62:70], 0)
	writeOctalField(header[70:78], 0)
	writeOctalField(header[78:86], 0)
	nameSize := uint64(len(rec.name) + 1)
	writeOctalField(header[86:94], nameSize)
	writeOctalField(header[94:102], 0)
	if _, err := w.w.Write(header); err != nil {
		return err
	}
	if _, err := w.w.Write([]byte(rec.name + "\x00")); err != nil {
		return err
	}
	pad := (110 + int(nameSize)) % 4
	if pad != 0 {
		if _, err := w.w.Write(make([]byte, 4-pad)); err != nil {
			return err
		}
	}
	if len(rec.data) > 0 {
		pad = len(rec.data) % 4
		if _, err := w.w.Write(rec.data); err != nil {
			return err
		}
		if pad != 0 {
			if _, err := w.w.Write(make([]byte, 4-pad)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *cpioNewcWriter) Close() error {
	trailer := cpioRecord{name: "TRAILER!!!", mode: 0, size: 0, mtime: 0}
	return w.WriteRecord(trailer)
}

type cpioNewcReader struct {
	r io.Reader
}

func newCpioNewcReader(r io.Reader) *cpioNewcReader {
	return &cpioNewcReader{r: r}
}

func (r *cpioNewcReader) ReadRecord() (cpioRecord, error) {
	header := make([]byte, 110)
	if _, err := io.ReadFull(r.r, header); err != nil {
		return cpioRecord{}, err
	}
	magic := string(header[0:6])
	if magic != "070701" {
		return cpioRecord{}, fmt.Errorf("invalid cpio magic: %s", magic)
	}
	nameSize := parseOctalField(header[86:94])
	fileSize := parseOctalField(header[46:54])
	mode := uint16(parseOctalField(header[14:22]) & 0xFFFF)
	name := make([]byte, nameSize)
	if _, err := io.ReadFull(r.r, name); err != nil {
		return cpioRecord{}, err
	}
	pad := (110 + int(nameSize)) % 4
	if pad != 0 {
		if _, err := io.CopyN(io.Discard, r.r, int64(4-pad)); err != nil {
			return cpioRecord{}, err
		}
	}
	rec := cpioRecord{
		name: string(name[:nameSize-1]),
		mode: mode,
	}
	if fileSize > 0 {
		rec.data = make([]byte, fileSize)
		if _, err := io.ReadFull(r.r, rec.data); err != nil {
			return cpioRecord{}, err
		}
		pad := int(fileSize) % 4
		if pad != 0 {
			if _, err := io.CopyN(io.Discard, r.r, int64(4-pad)); err != nil {
				return cpioRecord{}, err
			}
		}
	}
	return rec, nil
}

func writeOctalField(b []byte, v uint64) {
	for i := range b {
		b[i] = '0'
	}
	s := fmt.Sprintf("%07o", v)
	copy(b[0:7], []byte(s))
}

func parseOctalField(b []byte) uint64 {
	var v uint64
	for _, c := range b {
		if c >= '0' && c <= '7' {
			v = v*8 + uint64(c-'0')
		}
	}
	return v
}

// DVZip compression/decompression ------------------------------------------

func compressDVZip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	chunkSize := 256 * 1024
	for offset := 0; offset < len(data); offset += chunkSize {
		end := offset + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[offset:end]
		var compressed bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
		zw.Write(chunk)
		zw.Close()
		compressedBytes := compressed.Bytes()
		if len(compressedBytes) < len(chunk) {
			h := make([]byte, 4)
			binary.BigEndian.PutUint32(h, uint32(len(compressedBytes)))
			buf.Write(h)
			buf.Write(compressedBytes)
		} else {
			h := make([]byte, 4)
			binary.BigEndian.PutUint32(h, uint32(len(chunk))|0x80000000)
			buf.Write(h)
			buf.Write(chunk)
		}
	}
	term := make([]byte, 4)
	buf.Write(term)
	return buf.Bytes(), nil
}

func decompressDVZip(data []byte) ([]byte, error) {
	var out bytes.Buffer
	r := bytes.NewReader(data)
	for {
		header := make([]byte, 4)
		if _, err := io.ReadFull(r, header); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, err
		}
		raw := binary.BigEndian.Uint32(header)
		stored := raw&(1<<31) != 0
		length := uint32(raw & 0x7FFFFFFF)
		if length == 0 {
			break
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, fmt.Errorf("dvzip truncated block: wanted %d bytes: %w", length, err)
		}
		if stored {
			out.Write(payload)
		} else {
			decompressed, err := zlibDecompress(payload)
			if err != nil {
				return nil, err
			}
			out.Write(decompressed)
		}
	}
	return out.Bytes(), nil
}

func zlibDecompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// HTTP handlers ------------------------------------------------------------

func (m *Manager) handleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	var req DiscoverRequest
	if _, err := plist.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("invalid bplist: %v", err), http.StatusBadRequest)
		return
	}
	resp := DiscoverResponse{
		ReceiverComputerName: getHostname(),
		ReceiverModelName:    getModelName(),
		ReceiverRecordData:   []byte{},
		StatusCode:           100,
	}
	respBytes, err := plist.Marshal(resp, plist.BinaryFormat)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-apple-binary-plist")
	w.Write(respBytes)
}

func (m *Manager) handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	var req AskRequest
	if _, err := plist.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("invalid bplist: %v", err), http.StatusBadRequest)
		return
	}
	var accepted bool
	savePath := filepath.Join(m.downloads, sanitizeFilename(req.Files[0].FileName))
	m.mu.Lock()
	autoAccept := m.autoAcceptFn != nil && m.autoAcceptFn(req.SenderID)
	m.mu.Unlock()
	if autoAccept {
		accepted = true
	} else if m.askCallback != nil {
		resp := m.askCallback(&req)
		accepted = resp != nil && resp.Accepted
		if !accepted {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		savePath = resp.SavePath
	}
	if !accepted {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	_ = savePath

	transferID := req.TransferID
	if transferID == "" {
		transferID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%10000)
	}
	record := &TransferRecord{
		ID:        transferID,
		FileName:  sanitizeFilename(req.Files[0].FileName),
		Status:    StatusPending,
		PeerAddr:  r.RemoteAddr,
		SavePath:  savePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.transferMu.Lock()
	m.transfers[transferID] = record
	m.transferMu.Unlock()
	if m.store != nil {
		m.store.Save(record)
	}
	if m.onUpdate != nil {
		m.onUpdate(record)
	}

	resp := AskResponse{
		ReceiverComputerName: getHostname(),
		ReceiverModelName:    getModelName(),
	}
	respBytes, err := plist.Marshal(resp, plist.BinaryFormat)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-apple-binary-plist")
	w.Write(respBytes)
}

func (m *Manager) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	select {
	case m.uploadSem <- struct{}{}:
	default:
		http.Error(w, "too many concurrent uploads", http.StatusServiceUnavailable)
		return
	}
	defer func() { <-m.uploadSem }()

	transferID := r.Header.Get("Transfer-ID")
	if transferID == "" {
		http.Error(w, "missing Transfer-ID", http.StatusBadRequest)
		return
	}

	offsetStr := r.Header.Get("Transfer-Offset")
	totalStr := r.Header.Get("Transfer-Total")
	var offset, total int64
	if offsetStr != "" {
		offset, _ = strconv.ParseInt(offsetStr, 10, 64)
	}
	if totalStr != "" {
		total, _ = strconv.ParseInt(totalStr, 10, 64)
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	m.transferMu.Lock()
	record, exists := m.transfers[transferID]
	if !exists {
		record = &TransferRecord{
			ID:        transferID,
			Status:    StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.transfers[transferID] = record
	}
	m.transferMu.Unlock()

	if total > 0 {
		record.TotalBytes = total
	}
	if record.FileName == "" {
		record.FileName = sanitizeFilename(r.URL.Query().Get("file_name"))
	}

	if offset != record.BytesReceived {
		w.Header().Set("Transfer-Offset", strconv.FormatInt(record.BytesReceived, 10))
		w.WriteHeader(http.StatusConflict)
		if m.onUpdate != nil {
			m.onUpdate(record)
		}
		return
	}

	partialPath := m.partialPath(transferID)
	f, err := os.OpenFile(partialPath, os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := f.WriteAt(data, offset); err != nil {
		f.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f.Close()

	record.BytesReceived += int64(len(data))
	record.UpdatedAt = time.Now()
	record.Chunks = append(record.Chunks, ChunkInfo{Offset: offset, Length: int64(len(data))})
	if m.store != nil {
		m.store.Save(record)
	}

	complete := false
	if record.TotalBytes > 0 && record.BytesReceived >= record.TotalBytes {
		complete = true
		record.Status = StatusCompleted
		partialData, readErr := os.ReadFile(partialPath)
		if readErr == nil {
			if err := extractCpioArchive(partialData, m.downloads); err != nil {
				record.Status = StatusFailed
				record.Error = err.Error()
			} else {
				os.Remove(partialPath)
				if m.store != nil {
					m.store.Delete(transferID)
				}
			}
		} else {
			record.Status = StatusFailed
			record.Error = readErr.Error()
		}
	}

	w.Header().Set("Transfer-Offset", strconv.FormatInt(record.BytesReceived, 10))
	if complete {
		w.Header().Set("Transfer-Complete", "true")
	}
	w.WriteHeader(http.StatusOK)

	if m.onUpdate != nil {
		m.onUpdate(record)
	}
}

func (m *Manager) partialPath(id string) string {
	if m.store != nil {
		return m.store.PartialPath(id)
	}
	cfgDir, _ := os.UserConfigDir()
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgDir, "reversedrop", "transfers", id+".partial")
}

func (m *Manager) handleUploadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	transferID := r.URL.Query().Get("transferID")
	if transferID == "" {
		http.Error(w, "missing transferID", http.StatusBadRequest)
		return
	}
	m.transferMu.RLock()
	record, exists := m.transfers[transferID]
	m.transferMu.RUnlock()
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Transfer-Offset", strconv.FormatInt(record.BytesReceived, 10))
	if record.TotalBytes > 0 {
		w.Header().Set("Transfer-Total", strconv.FormatInt(record.TotalBytes, 10))
	}
	w.Header().Set("Transfer-Status", string(record.Status))
	w.WriteHeader(http.StatusOK)
}

func (m *Manager) handleError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "error", http.StatusBadRequest)
}

func getHostname() string {
	host, _ := os.Hostname()
	if host == "" {
		return "ReverseDrop-Device"
	}
	return host
}

func getModelName() string {
	return "ReverseDrop"
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	if name == "" || name == "." {
		return "untitled"
	}
	return name
}

// SendFile is the public client API -----------------------------------------

func SendFile(ctx context.Context, addr string, req TransferRequest, onProgress ProgressFunc) (*TransferResponse, error) {
	cert, err := generateCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"airdrop"},
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

	if onProgress != nil {
		onProgress("Connecting...", 0, 0)
	}
	if err := sendDiscover(conn, req); err != nil {
		close(done)
		return nil, fmt.Errorf("discover failed: %w", err)
	}
	if onProgress != nil {
		onProgress("Discovering...", 0, 0)
	}
	if err := sendAsk(conn, req); err != nil {
		close(done)
		return nil, fmt.Errorf("ask failed: %w", err)
	}
	if onProgress != nil {
		onProgress("Sending metadata...", 0, 0)
	}

	path := req.FileName
	var files []FileMeta
	var baseDir string

	if len(req.FileNames) > 0 {
		files = make([]FileMeta, 0, len(req.FileNames))
		for _, fn := range req.FileNames {
			files = append(files, FileMeta{
				FileType:    "public.data",
				FileName:    filepath.Base(fn),
				FileBomPath: "./" + filepath.Base(fn),
			})
		}
		baseDir = filepath.Dir(req.FileNames[0])
	} else {
		files = []FileMeta{{
			FileType:    "public.data",
			FileName:    req.FileName,
			FileBomPath: "./" + req.FileName,
		}}
		baseDir = filepath.Dir(path)
	}
	cpioData, err := writeCpioArchive(files, baseDir)
	if err != nil {
		close(done)
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}
	archive, err := compressDVZip(cpioData)
	if err != nil {
		close(done)
		return nil, fmt.Errorf("failed to compress archive: %w", err)
	}

	sent, err := sendChunkedUpload(conn, req, archive, defaultChunkSize, onProgress)
	close(done)
	if err != nil {
		saveTransferState(req.ID, addr, path, req.FileSize, archive, sent)
		return nil, fmt.Errorf("upload failed at offset %d: %w", sent, err)
	}
	deleteTransferState(req.ID)
	return &TransferResponse{ID: req.ID, Accepted: true, SavePath: req.FileName}, nil
}

func sendDiscover(conn net.Conn, req TransferRequest) error {
	discReq := DiscoverRequest{
		SenderComputerName: req.SenderName,
		SenderModelName:    "ReverseDrop",
		SenderRecordData:   []byte{},
		SenderID:           "com.falthera.reversedrop",
		TransferVersion:    1,
	}
	body, err := plist.Marshal(discReq, plist.BinaryFormat)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", "https://"+conn.RemoteAddr().String()+"/Discover", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/x-apple-binary-plist")
	if err := httpReq.Write(conn); err != nil {
		return err
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func sendAsk(conn net.Conn, req TransferRequest) error {
	files := []FileMeta{{
		FileType:           "public.data",
		FileName:           req.FileName,
		FileBomPath:        "./" + req.FileName,
		FileIsDirectory:    false,
		ConvertMediaFormats: false,
	}}
	if len(req.FileNames) > 0 {
		files = make([]FileMeta, 0, len(req.FileNames))
		for _, fn := range req.FileNames {
			files = append(files, FileMeta{
				FileType:           "public.data",
				FileName:           filepath.Base(fn),
				FileBomPath:        "./" + filepath.Base(fn),
				FileIsDirectory:    false,
				ConvertMediaFormats: false,
			})
		}
	}
	askReq := AskRequest{
		SenderComputerName: req.SenderName,
		SenderModelName:    "ReverseDrop",
		SenderID:           "com.falthera.reversedrop",
		TransferID:         req.ID,
		Files:              files,
	}
	body, err := plist.Marshal(askReq, plist.BinaryFormat)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", "https://"+conn.RemoteAddr().String()+"/Ask", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/x-apple-binary-plist")
	if err := httpReq.Write(conn); err != nil {
		return err
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func sendChunkedUpload(conn net.Conn, req TransferRequest, archive []byte, chunkSize int, onProgress ProgressFunc) (int64, error) {
	return sendChunkedUploadFrom(conn, req, archive, chunkSize, 0, onProgress)
}

func sendChunkedUploadFrom(conn net.Conn, req TransferRequest, archive []byte, chunkSize int, startOffset int64, onProgress ProgressFunc) (int64, error) {
	total := int64(len(archive))
	sent := startOffset

	for sent < total {
		end := sent + int64(chunkSize)
		if end > total {
			end = total
		}
		chunk := archive[sent:end]

		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			if _, werr := io.Copy(pw, bytes.NewReader(chunk)); werr != nil {
				slog.Warn("chunk pipe write error", "error", werr)
			}
		}()

		httpReq, err := http.NewRequest("POST", "https://"+conn.RemoteAddr().String()+"/Upload", pr)
		if err != nil {
			return sent, err
		}
		httpReq.Header.Set("Content-Type", "application/x-dvzip")
		httpReq.Header.Set("Transfer-ID", req.ID)
		httpReq.Header.Set("Transfer-Offset", strconv.FormatInt(sent, 10))
		httpReq.Header.Set("Transfer-Total", strconv.FormatInt(total, 10))
		httpReq.TransferEncoding = []string{"chunked"}

		if err := httpReq.Write(conn); err != nil {
			return sent, err
		}

		resp, err := http.ReadResponse(bufio.NewReader(conn), httpReq)
		if err != nil {
			return sent, err
		}
		resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			serverOffsetStr := resp.Header.Get("Transfer-Offset")
			if serverOffsetStr != "" {
				serverOffset, _ := strconv.ParseInt(serverOffsetStr, 10, 64)
				if serverOffset > sent {
					sent = serverOffset
				} else {
					sent += int64(len(chunk))
				}
			} else {
				sent += int64(len(chunk))
			}
		case http.StatusConflict:
			serverOffsetStr := resp.Header.Get("Transfer-Offset")
			if serverOffsetStr != "" {
				serverOffset, _ := strconv.ParseInt(serverOffsetStr, 10, 64)
				sent = serverOffset
			}
		default:
			return sent, fmt.Errorf("upload failed with status %d", resp.StatusCode)
		}

		if onProgress != nil {
			state := "Transferring..."
			if startOffset > 0 {
				state = "Resuming..."
			}
			onProgress(state, sent, total)
		}
	}

	return sent, nil
}

// ResumeFileTransfer resumes a previously failed or paused transfer.
func ResumeFileTransfer(ctx context.Context, transferID string, onProgress ProgressFunc) (*TransferResponse, error) {
	state, err := loadTransferState(transferID)
	if err != nil {
		return nil, fmt.Errorf("failed to load transfer state: %w", err)
	}

	addr := state["addr"].(string)
	path := state["path"].(string)
	fileSize := int64(state["file_size"].(float64))
	fileName := state["file_name"].(string)
	startOffset := int64(state["sent"].(float64))

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	files := []FileMeta{{
		FileType:    "public.data",
		FileName:    fileName,
		FileBomPath: "./" + fileName,
	}}
	cpioData, err := writeCpioArchive(files, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	archive, err := compressDVZip(cpioData)
	if err != nil {
		return nil, err
	}

	cert, err := generateCert()
	if err != nil {
		return nil, err
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"airdrop"},
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

	req := TransferRequest{
		ID:         transferID,
		FileName:   fileName,
		FileSize:   fileSize,
		SenderName: "ReverseDrop User",
	}

	if err := sendDiscover(conn, req); err != nil {
		close(done)
		return nil, fmt.Errorf("discover failed: %w", err)
	}
	if err := sendAsk(conn, req); err != nil {
		close(done)
		return nil, fmt.Errorf("ask failed: %w", err)
	}

	sent, err := sendChunkedUploadFrom(conn, req, archive, defaultChunkSize, startOffset, onProgress)
	close(done)
	if err != nil {
		return nil, fmt.Errorf("resume failed at offset %d: %w", sent, err)
	}
	deleteTransferState(transferID)
	return &TransferResponse{ID: transferID, Accepted: true, SavePath: fileName}, nil
}

func saveTransferState(id, addr, path string, fileSize int64, archive []byte, sent int64) error {
	cfgDir, _ := os.UserConfigDir()
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	dir := filepath.Join(cfgDir, "reversedrop", "transfers")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	state := map[string]interface{}{
		"id":        id,
		"addr":      addr,
		"path":      path,
		"sent":      sent,
		"total":     int64(len(archive)),
		"file_name": filepath.Base(path),
		"file_size": fileSize,
	}
	data, _ := json.Marshal(state)
	return os.WriteFile(filepath.Join(dir, id+".json"), data, 0o600)
}

func loadTransferState(id string) (map[string]interface{}, error) {
	cfgDir, _ := os.UserConfigDir()
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	path := filepath.Join(cfgDir, "reversedrop", "transfers", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return state, nil
}

func deleteTransferState(id string) {
	cfgDir, _ := os.UserConfigDir()
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	path := filepath.Join(cfgDir, "reversedrop", "transfers", id+".json")
	os.Remove(path)
}

// Exported wrappers for GUI and tests ---------------------------------------

// WriteCpioArchive creates a CPIO archive from files.
func WriteCpioArchive(files []FileMeta, baseDir string) ([]byte, error) {
	return writeCpioArchive(files, baseDir)
}

// CompressDVZip compresses data using the DVZip format.
func CompressDVZip(data []byte) ([]byte, error) {
	return compressDVZip(data)
}

// GenerateCert generates a temporary TLS certificate.
func GenerateCert() (tls.Certificate, error) {
	return generateCert()
}

// SendDiscover sends a discover request over the connection.
func SendDiscover(conn net.Conn, req TransferRequest) error {
	return sendDiscover(conn, req)
}

// SendAsk sends an ask request over the connection.
func SendAsk(conn net.Conn, req TransferRequest) error {
	return sendAsk(conn, req)
}

// SendChunkedUploadFrom sends chunks starting from a given offset.
func SendChunkedUploadFrom(conn net.Conn, req TransferRequest, archive []byte, chunkSize int, startOffset int64, onProgress ProgressFunc) (int64, error) {
	return sendChunkedUploadFrom(conn, req, archive, chunkSize, startOffset, onProgress)
}

// LoadTransferState loads a transfer state from disk.
func LoadTransferState(id string) (map[string]interface{}, error) {
	return loadTransferState(id)
}

// DeleteTransferState deletes a transfer state from disk.
func DeleteTransferState(id string) {
	deleteTransferState(id)
}
