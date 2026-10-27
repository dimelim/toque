package roblox

import "testing"

func TestParsePrivateLink(t *testing.T) {
	cases := []struct {
		link string
		want Target
		ok   bool
	}{
		{"https://www.roblox.com/share?code=4f5d8e1c0b6a2a4c9a8f3b2e1d0c9b8a&type=Server", Target{ShareCode: "4f5d8e1c0b6a2a4c9a8f3b2e1d0c9b8a"}, true},
		{"https://www.roblox.com/games/14315258385/Realistic-Street-Soccer?privateServerLinkCode=1020499390872690887970171", Target{PlaceID: 14315258385, LinkCode: "1020499390872690887970171"}, true},
		{"  https://roblox.com/share?code=abc123XYZ&type=Server  ", Target{ShareCode: "abc123XYZ"}, true},
		{"https://www.roblox.com/share?code=abc123XYZ&type=ExperienceDetails", Target{}, false},
		{"https://evilroblox.com/share?code=abc123XYZ&type=Server", Target{}, false},
		{"http://www.roblox.com/share?code=abc123XYZ&type=Server", Target{}, false},
		{"https://www.roblox.com/games/14315258385/x?privateServerLinkCode=12ab", Target{}, false},
		{"https://www.roblox.com/share?code=abc%22%20--x&type=Server", Target{}, false},
		{"https://www.roblox.com/games/14315258385/Realistic-Street-Soccer", Target{}, false},
	}
	for _, c := range cases {
		got, err := ParsePrivateLink(c.link)
		if (err == nil) != c.ok || got != c.want {
			t.Errorf("ParsePrivateLink(%q) = %+v, %v", c.link, got, err)
		}
	}
}

func TestTargetURL(t *testing.T) {
	cases := []struct {
		target Target
		want   string
	}{
		{Target{ShareCode: "abc123"}, "roblox://navigation/share_links?code=abc123&type=Server"},
		{Target{PlaceID: 14315258385, LinkCode: "102049"}, "roblox://experiences/start?placeId=14315258385&linkCode=102049"},
		{Target{PlaceID: 14315258385, JobID: "81806afc-e000-4858-b546-220dfe00c90e"}, "roblox://experiences/start?placeId=14315258385&gameInstanceId=81806afc-e000-4858-b546-220dfe00c90e"},
		{Target{PlaceID: 14315258385}, "roblox://experiences/start?placeId=14315258385"},
	}
	for _, c := range cases {
		if got := c.target.URL(); got != c.want {
			t.Errorf("%+v.URL() = %q, want %q", c.target, got, c.want)
		}
	}
}
