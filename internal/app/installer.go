package app

import (
	"os"
	"os/exec"
	"time"

	"toque/internal/install"
)

func (s *Service) InstallState() install.State {
	return install.Current(Version)
}

func (s *Service) Install(desktop bool) error {
	target, err := install.Install(Version, desktop)
	if err != nil {
		return err
	}
	cmd := exec.Command(target, "--launched")
	cmd.Dir = install.Dir()
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
}

func (s *Service) Uninstall() error {
	if err := s.RevertAll(); err != nil {
		return err
	}
	if err := install.Uninstall(); err != nil {
		return err
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
}
