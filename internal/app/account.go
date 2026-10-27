package app

import (
	"context"
	"sync"
	"time"

	"toque/internal/roblox"
)

type Account struct {
	roblox.Account
	Avatar string `json:"avatar"`
}

var (
	avatarMu   sync.Mutex
	avatarUser int64
	avatarData string
	avatarAt   time.Time
)

func (s *Service) Account() (Account, error) {
	inst, err := roblox.Find()
	if err != nil {
		return Account{}, err
	}
	acc, err := roblox.ReadAccount(inst)
	if err != nil {
		return Account{}, err
	}
	return Account{Account: acc, Avatar: avatarFor(acc.UserID)}, nil
}

func avatarFor(userID int64) string {
	avatarMu.Lock()
	defer avatarMu.Unlock()
	if avatarUser == userID && avatarData != "" && time.Since(avatarAt) < 30*time.Minute {
		return avatarData
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if img, err := roblox.AvatarHeadshot(ctx, userID); err == nil {
		avatarUser, avatarData, avatarAt = userID, img, time.Now()
	}
	if avatarUser != userID {
		return ""
	}
	return avatarData
}
