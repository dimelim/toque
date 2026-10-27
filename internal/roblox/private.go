package roblox

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	shareCode = regexp.MustCompile(`^[A-Za-z0-9_-]{6,64}$`)
	linkCode  = regexp.MustCompile(`^[0-9]{6,40}$`)
	gamePath  = regexp.MustCompile(`^/games/([0-9]{1,15})(/|$)`)
)

var errNotPrivate = errors.New("ese enlace no es de un servidor privado de Roblox")

func ParsePrivateLink(raw string) (Target, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" {
		return Target{}, errNotPrivate
	}
	host := strings.ToLower(u.Hostname())
	if host != "roblox.com" && !strings.HasSuffix(host, ".roblox.com") {
		return Target{}, errNotPrivate
	}
	q := u.Query()
	if strings.TrimSuffix(u.Path, "/") == "/share" {
		if q.Get("type") != "Server" || !shareCode.MatchString(q.Get("code")) {
			return Target{}, errNotPrivate
		}
		return Target{ShareCode: q.Get("code")}, nil
	}
	m := gamePath.FindStringSubmatch(u.Path)
	code := q.Get("privateServerLinkCode")
	if m == nil || !linkCode.MatchString(code) {
		return Target{}, errNotPrivate
	}
	place, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return Target{}, errNotPrivate
	}
	return Target{PlaceID: place, LinkCode: code}, nil
}
