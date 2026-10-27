package roblox

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type PublicServer struct {
	ID         string  `json:"id"`
	MaxPlayers int     `json:"maxPlayers"`
	Playing    int     `json:"playing"`
	FPS        float64 `json:"fps"`
	Ping       float64 `json:"ping"`
}

type serverPage struct {
	Data           []PublicServer `json:"data"`
	NextPageCursor string         `json:"nextPageCursor"`
}

var httpClient = &http.Client{Timeout: 12 * time.Second}

type RateLimitError struct {
	Wait time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("roblox limits the server list for %s", e.Wait)
}

func rateLimitWait(h http.Header) time.Duration {
	if secs, err := strconv.Atoi(h.Get("x-ratelimit-reset")); err == nil && secs > 0 && secs <= 60 {
		return time.Duration(secs+1) * time.Second
	}
	return 60 * time.Second
}

type SortOrder string

const (
	FullestFirst  SortOrder = "Desc"
	EmptiestFirst SortOrder = "Asc"
)

func PublicPage(ctx context.Context, placeID int64, order SortOrder, cursor string) ([]PublicServer, string, error) {
	url := fmt.Sprintf("https://games.roblox.com/v1/games/%d/servers/Public?sortOrder=%s&excludeFullGames=true&limit=100", placeID, order)
	if cursor != "" {
		url += "&cursor=" + cursor
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, "", &RateLimitError{Wait: rateLimitWait(resp.Header)}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("roblox answered %s for the server list", resp.Status)
	}
	var page serverPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, "", err
	}
	return page.Data, page.NextPageCursor, nil
}

func PublicServers(ctx context.Context, placeID int64, pages int) ([]PublicServer, error) {
	var all []PublicServer
	cursor := ""
	for i := 0; i < pages; i++ {
		page, next, err := PublicPage(ctx, placeID, FullestFirst, cursor)
		all = append(all, page...)
		if err != nil {
			return all, err
		}
		if next == "" {
			break
		}
		cursor = next
		time.Sleep(400 * time.Millisecond)
	}
	return all, nil
}

func SampleCandidates(servers []PublicServer, exclude map[string]bool, n int, seed int64) []PublicServer {
	var c []PublicServer
	for _, s := range servers {
		if exclude[s.ID] || s.Playing >= s.MaxPlayers || s.Playing == 0 {
			continue
		}
		c = append(c, s)
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })
	if len(c) > n {
		c = c[:n]
	}
	return c
}
