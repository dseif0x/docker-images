package main

import (
	"os"
	"testing"
)

func TestExtractShortcode(t *testing.T) {
	cases := map[string]string{
		"https://www.instagram.com/reel/DabGHCOBcJB/?igsh=OGw4dHQ5dDFuYWpv": "DabGHCOBcJB",
		"https://www.instagram.com/reels/DatrKZhO4jO/":                      "DatrKZhO4jO",
		"https://www.instagram.com/p/DabGHCOBcJB/":                          "DabGHCOBcJB",
		"https://www.instagram.com/someuser/reel/DabGHCOBcJB/":              "DabGHCOBcJB",
	}
	for link, want := range cases {
		got, err := extractShortcode(link)
		if err != nil || got != want {
			t.Errorf("extractShortcode(%s) = %q, %v; want %q", link, got, err, want)
		}
	}
}

func TestFindReelLink(t *testing.T) {
	cases := map[string]string{
		"https://www.instagram.com/reel/DfiXWNJgiub/?exln=NHh1em5wYXNxc3Uz": "https://www.instagram.com/reel/DfiXWNJgiub/?exln=NHh1em5wYXNxc3Uz",
		"look at this https://www.instagram.com/reel/DfiXWNJgiub/ lol":      "https://www.instagram.com/reel/DfiXWNJgiub/",
		"https://instagram.com/reels/DatrKZhO4jO/":                          "https://instagram.com/reels/DatrKZhO4jO/",
		"instagram.com/p/DabGHCOBcJB/":                                      "https://instagram.com/p/DabGHCOBcJB/",
		"https://m.instagram.com/reel/DabGHCOBcJB/?igsh=x":                  "https://m.instagram.com/reel/DabGHCOBcJB/?igsh=x",
		"https://www.instagram.com/share/reel/BAqWzLx1E2/":                  "https://www.instagram.com/share/reel/BAqWzLx1E2/",
		"https://www.instagram.com/someuser/":                               "",
		"hello":                                                             "",
	}
	for text, want := range cases {
		if got := findReelLink(text); got != want {
			t.Errorf("findReelLink(%q) = %q; want %q", text, got, want)
		}
	}
}

func TestShortcodeToMediaID(t *testing.T) {
	got, err := shortcodeToMediaID("DabGHCOBcJB")
	if err != nil || got != "3934765571136406081" {
		t.Errorf("shortcodeToMediaID = %q, %v; want 3934765571136406081", got, err)
	}
}

func TestResolveLive(t *testing.T) {
	if os.Getenv("IG_LIVE_TEST") == "" {
		t.Skip("set IG_LIVE_TEST=1 to hit Instagram for real")
	}
	u, _, err := DownloadInstagramReelBytes("https://www.instagram.com/reel/DabGHCOBcJB/?igsh=OGw4dHQ5dDFuYWpv")
	_ = u
	t.Logf("live result: err=%v", err)
}
