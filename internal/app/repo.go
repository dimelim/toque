package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const repoName = "dimelim/toque"

type Repo struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Stars int    `json:"stars"`
}

var (
	repoMu    sync.Mutex
	repoStars = -1
	repoAt    time.Time
)

func (s *Service) Repo() Repo {
	repoMu.Lock()
	defer repoMu.Unlock()
	if time.Since(repoAt) > 30*time.Minute {
		if n, err := fetchStars(); err == nil {
			repoStars, repoAt = n, time.Now()
		}
	}
	return Repo{Name: repoName, URL: "https://github.com/" + repoName, Stars: repoStars}
}

func fetchStars() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repoName, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Toque/"+Version)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github respondió %s", resp.Status)
	}
	var body struct {
		Stars int `json:"stargazers_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	return body.Stars, nil
}
