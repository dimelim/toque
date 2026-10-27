package roblox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

func PublicServers(ctx context.Context, placeID int64, pages int) ([]PublicServer, error) {
	var all []PublicServer
	cursor := ""
	for i := 0; i < pages; i++ {
		url := fmt.Sprintf("https://games.roblox.com/v1/games/%d/servers/Public?sortOrder=Desc&excludeFullGames=true&limit=100", placeID)
		if cursor != "" {
			url += "&cursor=" + cursor
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return all, err
		}
		req.Header.Set("Accept", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			return all, err
		}
		var page serverPage
		err = json.NewDecoder(resp.Body).Decode(&page)
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			return all, fmt.Errorf("roblox rate limited the server list")
		}
		if err != nil {
			return all, err
		}
		all = append(all, page.Data...)
		if page.NextPageCursor == "" {
			break
		}
		cursor = page.NextPageCursor
		time.Sleep(400 * time.Millisecond)
	}
	return all, nil
}

func RankCandidates(servers []PublicServer, exclude map[string]bool, n int) []PublicServer {
	var c []PublicServer
	for _, s := range servers {
		if exclude[s.ID] || s.Playing >= s.MaxPlayers || s.Playing == 0 || s.Ping <= 0 {
			continue
		}
		c = append(c, s)
	}
	sort.Slice(c, func(i, j int) bool { return c[i].Ping < c[j].Ping })
	if len(c) > n {
		c = c[:n]
	}
	return c
}
