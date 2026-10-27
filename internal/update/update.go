package update

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"toque/internal/install"
	"toque/internal/winx"
)

const (
	releases    = "https://github.com/dimelim/toque-releases/releases/"
	manifestURL = releases + "latest/download/latest.json"
	maxExeBytes = 100 << 20
)

type Manifest struct {
	Version   string    `json:"version"`
	URL       string    `json:"url"`
	SHA256    string    `json:"sha256"`
	Notes     string    `json:"notes"`
	Published time.Time `json:"published"`
}

type Release struct {
	Manifest
	raw []byte
	sig []byte
}

var client = &http.Client{Timeout: 5 * time.Minute}

func get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("el servidor de versiones respondió %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

func parse(raw, sig []byte) (Release, error) {
	key, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return Release{}, errors.New("la clave pública de actualizaciones no es válida")
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sig)))
	if err != nil || !ed25519.Verify(key, raw, signature) {
		return Release{}, errors.New("la firma de la actualización no es válida")
	}
	r := Release{raw: raw, sig: sig}
	if err := json.Unmarshal(raw, &r.Manifest); err != nil {
		return Release{}, err
	}
	if r.Version == "" || len(r.SHA256) != 64 || !strings.HasPrefix(r.URL, releases+"download/") {
		return Release{}, errors.New("el manifiesto de la actualización no es válido")
	}
	return r, nil
}

func Latest(ctx context.Context) (Release, error) {
	raw, err := get(ctx, manifestURL, 64<<10)
	if err != nil {
		return Release{}, err
	}
	sig, err := get(ctx, manifestURL+".sig", 4<<10)
	if err != nil {
		return Release{}, err
	}
	return parse(raw, sig)
}

func Newer(m Manifest, current string) bool {
	return install.Compare(m.Version, current) > 0
}

func dir() string {
	return winx.DataDir("update")
}

func exePath(version string) string {
	return filepath.Join(dir(), "Toque-"+version+".exe")
}

func fileSHA(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func Download(ctx context.Context, r Release) error {
	final := exePath(r.Version)
	if !strings.EqualFold(fileSHA(final), r.SHA256) {
		if err := fetchExe(ctx, r.Manifest, final); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(dir(), "pending.json"), r.raw, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir(), "pending.json.sig"), r.sig, 0o644)
}

func fetchExe(ctx context.Context, m Manifest, final string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("la descarga respondió %s", resp.Status)
	}
	tmp := final + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, maxExeBytes))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), m.SHA256) {
		err = errors.New("el archivo descargado no coincide con el que firmó Toque")
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, final)
}

func Pending(current string) (Manifest, string, bool) {
	raw, err := os.ReadFile(filepath.Join(dir(), "pending.json"))
	if err != nil {
		return Manifest{}, "", false
	}
	sig, err := os.ReadFile(filepath.Join(dir(), "pending.json.sig"))
	if err != nil {
		return Manifest{}, "", false
	}
	r, err := parse(raw, sig)
	if err != nil || !Newer(r.Manifest, current) {
		return Manifest{}, "", false
	}
	path := exePath(r.Version)
	if !strings.EqualFold(fileSHA(path), r.SHA256) {
		return Manifest{}, "", false
	}
	return r.Manifest, path, true
}

func Clear() {
	os.RemoveAll(dir())
}

func Apply(path string) error {
	cmd := exec.Command(path, "--apply-update")
	cmd.Dir = filepath.Dir(path)
	return cmd.Start()
}
