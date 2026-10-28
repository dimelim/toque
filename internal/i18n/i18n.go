package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
	"sync/atomic"
)

type Lang string

const (
	ES Lang = "es"
	PT Lang = "pt"
	EN Lang = "en"
)

var (
	current atomic.Value
	dictsMu sync.RWMutex
	dicts   = map[Lang]map[string]string{}
)

func Parse(s string) (Lang, bool) {
	switch l := Lang(strings.ToLower(s)); l {
	case ES, PT, EN:
		return l, true
	}
	return "", false
}

func Load(files fs.FS) error {
	loaded := map[Lang]map[string]string{}
	err := fs.WalkDir(files, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".json" {
			return err
		}
		l, ok := Parse(strings.TrimSuffix(path.Base(p), ".json"))
		if !ok {
			return nil
		}
		raw, err := fs.ReadFile(files, p)
		if err != nil {
			return err
		}
		entries := map[string]string{}
		if err := json.Unmarshal(raw, &entries); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		if loaded[l] == nil {
			loaded[l] = map[string]string{}
		}
		for k, v := range entries {
			loaded[l][k] = v
		}
		return nil
	})
	if err != nil {
		return err
	}
	dictsMu.Lock()
	dicts = loaded
	dictsMu.Unlock()
	return nil
}

func Set(l Lang) { current.Store(l) }

func Current() Lang {
	if l, ok := current.Load().(Lang); ok {
		return l
	}
	return ES
}

func T(s string) string {
	l := Current()
	if l == ES {
		return s
	}
	dictsMu.RLock()
	defer dictsMu.RUnlock()
	if v := dicts[l][s]; v != "" {
		return v
	}
	return s
}

func Tf(format string, a ...any) string { return fmt.Sprintf(T(format), a...) }

func Mark(s string) string { return s }
