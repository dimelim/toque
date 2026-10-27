package roblox

import (
	"errors"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

var errReparse = errors.New("ruta insegura: un tramo es un enlace o junction")

func isReparse(path string) bool {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return true
	}
	attrs, err := windows.GetFileAttributes(p)
	if err != nil {
		return false
	}
	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}

func SafeTarget(root, target string) error {
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	if !strings.HasPrefix(strings.ToLower(target), strings.ToLower(root)+`\`) {
		return errors.New("ruta fuera de la carpeta de Roblox")
	}
	for p := target; len(p) >= len(root); p = filepath.Dir(p) {
		if isReparse(p) {
			return errReparse
		}
		if p == filepath.Dir(p) {
			break
		}
	}
	return nil
}
