package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return errors.New("toque-sign keygen <clave> | toque-sign sign <clave> <archivo>")
	}
	switch args[0] {
	case "keygen":
		return keygen(args[1])
	case "sign":
		if len(args) < 3 {
			return errors.New("falta el archivo a firmar")
		}
		return sign(args[1], args[2])
	}
	return fmt.Errorf("orden desconocida %q", args[0])
}

func keygen(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s ya existe, no lo sobrescribo", path)
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(priv)), 0o600); err != nil {
		return err
	}
	fmt.Println(base64.StdEncoding.EncodeToString(pub))
	return nil
}

func sign(keyPath, file string) error {
	raw, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(key) != ed25519.PrivateKeySize {
		return errors.New("la clave privada no es válida")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(ed25519.PrivateKey(key), data)
	return os.WriteFile(file+".sig", []byte(base64.StdEncoding.EncodeToString(sig)), 0o644)
}
