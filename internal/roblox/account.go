package roblox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"toque/internal/i18n"
)

type Account struct {
	UserID      int64  `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

func ReadAccount(inst Install) (Account, error) {
	b, err := os.ReadFile(filepath.Join(inst.Root, "LocalStorage", "appStorage.json"))
	if err != nil {
		return Account{}, err
	}
	var raw struct {
		UserID      json.RawMessage `json:"UserId"`
		Username    string          `json:"Username"`
		DisplayName string          `json:"DisplayName"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return Account{}, err
	}
	id, err := parseUserID(raw.UserID)
	if err != nil || raw.Username == "" {
		return Account{}, errors.New(i18n.T("Roblox no tiene una sesión iniciada"))
	}
	return Account{UserID: id, Username: raw.Username, DisplayName: raw.DisplayName}, nil
}

func parseUserID(raw json.RawMessage) (int64, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strconv.ParseInt(text, 10, 64)
	}
	var n int64
	err := json.Unmarshal(raw, &n)
	return n, err
}

func AvatarHeadshot(ctx context.Context, userID int64) (string, error) {
	url := fmt.Sprintf("https://thumbnails.roblox.com/v1/users/avatar-headshot?userIds=%d&size=150x150&format=Png&isCircular=false", userID)
	var meta struct {
		Data []struct {
			State    string `json:"state"`
			ImageURL string `json:"imageUrl"`
		} `json:"data"`
	}
	if err := getJSON(ctx, url, &meta); err != nil {
		return "", err
	}
	if len(meta.Data) == 0 || meta.Data[0].State != "Completed" || meta.Data[0].ImageURL == "" {
		return "", errors.New(i18n.T("Roblox no tiene lista la foto del avatar"))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, meta.Data[0].ImageURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(i18n.T("la foto del avatar respondió %s"), resp.Status)
	}
	img, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(img), nil
}

func getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("roblox answered %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
