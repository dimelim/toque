package roblox

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type BasicSettings struct {
	path string
	text string
}

func LoadBasicSettings(path string) (*BasicSettings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &BasicSettings{path: path, text: string(b)}, nil
}

func (s *BasicSettings) pattern(kind, name string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`<%s name="%s">([^<]*)</%s>`, kind, regexp.QuoteMeta(name), kind))
}

func (s *BasicSettings) get(kind, name string) (string, bool) {
	m := s.pattern(kind, name).FindStringSubmatch(s.text)
	if m == nil {
		return "", false
	}
	return m[1], true
}

func (s *BasicSettings) set(kind, name, value string) bool {
	re := s.pattern(kind, name)
	if !re.MatchString(s.text) {
		return false
	}
	s.text = re.ReplaceAllString(s.text, fmt.Sprintf(`<%s name="%s">%s</%s>`, kind, name, value, kind))
	return true
}

func (s *BasicSettings) Int(name string) (int, bool) {
	v, ok := s.get("int", name)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	return n, err == nil
}

func (s *BasicSettings) SetInt(name string, v int) bool {
	return s.set("int", name, strconv.Itoa(v))
}

func (s *BasicSettings) Token(name string) (int, bool) {
	v, ok := s.get("token", name)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	return n, err == nil
}

func (s *BasicSettings) SetToken(name string, v int) bool {
	return s.set("token", name, strconv.Itoa(v))
}

func (s *BasicSettings) Bool(name string) (bool, bool) {
	v, ok := s.get("bool", name)
	return v == "true", ok
}

func (s *BasicSettings) SetBool(name string, v bool) bool {
	return s.set("bool", name, strconv.FormatBool(v))
}

func (s *BasicSettings) Save() error {
	if err := SafeTarget(filepath.Dir(filepath.Dir(s.path)), s.path); err != nil {
		return err
	}
	tmp := s.path + ".toque-tmp"
	if err := os.WriteFile(tmp, []byte(s.text), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
