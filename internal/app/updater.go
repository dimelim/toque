package app

import (
	"context"
	"errors"
	"os"
	"time"

	"toque/internal/install"
	"toque/internal/update"
	"toque/internal/winx"
)

type UpdateInfo struct {
	Current string    `json:"current"`
	Latest  string    `json:"latest"`
	Notes   string    `json:"notes"`
	State   string    `json:"state"`
	Error   string    `json:"error"`
	Checked time.Time `json:"checked"`
}

func othersRunning() bool {
	self := uint32(os.Getpid())
	for _, p := range winx.FindByName("Toque.exe") {
		if p.PID != self {
			return true
		}
	}
	return false
}

func ApplyPendingUpdate() bool {
	if !install.Current(Version).RunningInstalled || othersRunning() {
		return false
	}
	_, path, ok := update.Pending(Version)
	if !ok {
		update.Clear()
		return false
	}
	return update.Apply(path) == nil
}

func (s *Service) setUpdate(fn func(*UpdateInfo)) {
	s.updMu.Lock()
	fn(&s.upd)
	info := s.upd
	s.updMu.Unlock()
	if s.emit != nil {
		s.emit("update", info)
	}
}

func (s *Service) UpdateStatus() UpdateInfo {
	s.updMu.Lock()
	defer s.updMu.Unlock()
	return s.upd
}

func (s *Service) StartUpdates() {
	if !install.Current(Version).RunningInstalled {
		return
	}
	go func() {
		for {
			s.checkUpdate()
			time.Sleep(30 * time.Minute)
		}
	}()
}

func (s *Service) checkUpdate() {
	if !s.checking.CompareAndSwap(false, true) {
		return
	}
	defer s.checking.Store(false)
	s.setUpdate(func(u *UpdateInfo) { u.State, u.Error = "checking", "" })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	r, err := update.Latest(ctx)
	if err != nil {
		s.setUpdate(func(u *UpdateInfo) { u.State, u.Error, u.Checked = "error", err.Error(), time.Now() })
		return
	}
	if !update.Newer(r.Manifest, Version) {
		s.setUpdate(func(u *UpdateInfo) { u.State, u.Latest, u.Notes, u.Checked = "current", r.Version, r.Notes, time.Now() })
		return
	}
	s.setUpdate(func(u *UpdateInfo) { u.State, u.Latest, u.Notes = "downloading", r.Version, r.Notes })
	if err := update.Download(ctx, r); err != nil {
		s.setUpdate(func(u *UpdateInfo) { u.State, u.Error, u.Checked = "error", err.Error(), time.Now() })
		return
	}
	s.setUpdate(func(u *UpdateInfo) { u.State, u.Checked = "ready", time.Now() })
}

func (s *Service) CheckUpdate() (UpdateInfo, error) {
	if !install.Current(Version).RunningInstalled {
		return s.UpdateStatus(), errors.New("instala Toque para recibir actualizaciones")
	}
	s.checkUpdate()
	return s.UpdateStatus(), nil
}

func (s *Service) ApplyUpdate() error {
	if err := s.idle(); err != nil {
		return err
	}
	_, path, ok := update.Pending(Version)
	if !ok {
		return errors.New("no hay ninguna actualización lista")
	}
	if err := update.Apply(path); err != nil {
		return err
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
}
