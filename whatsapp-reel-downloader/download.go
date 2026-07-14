package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	browserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	igAppID          = "936619743392459"
	graphQLDocID     = "8845758582119845"
	shortcodeAlpha   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

var (
	shortcodeRegex  = regexp.MustCompile(`instagram\.com/(?:[A-Za-z0-9_.]+/)?(?:reels?|p|tv)/([A-Za-z0-9_-]+)`)
	lsdTokenRegex   = regexp.MustCompile(`"LSD",\[\],\{"token":"([^"]+)"`)
	embedVideoRegex = regexp.MustCompile(`"video_url":"((?:[^"\\]|\\.)*)"`)
)

type graphQLMediaResponse struct {
	Data struct {
		ShortcodeMedia *struct {
			IsVideo  bool   `json:"is_video"`
			VideoURL string `json:"video_url"`
		} `json:"xdt_shortcode_media"`
	} `json:"data"`
}

type apiMediaInfoResponse struct {
	Items []struct {
		VideoVersions []struct {
			URL string `json:"url"`
		} `json:"video_versions"`
	} `json:"items"`
}

// DownloadInstagramReelBytes downloads the Instagram reel video as a byte slice.
func DownloadInstagramReelBytes(link string) ([]byte, string, error) {
	shortcode, err := extractShortcode(link)
	if err != nil {
		return nil, "", err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create cookie jar: %w", err)
	}
	client := &http.Client{Jar: jar, Timeout: 60 * time.Second}

	videoURL, err := resolveVideoURL(client, shortcode)
	if err != nil {
		return nil, "", err
	}

	videoBytes, err := downloadVideo(client, videoURL)
	if err != nil {
		return nil, "", err
	}

	return videoBytes, shortcode + ".mp4", nil
}

func extractShortcode(link string) (string, error) {
	matches := shortcodeRegex.FindStringSubmatch(link)
	if matches == nil {
		return "", fmt.Errorf("could not extract shortcode from link: %s", link)
	}
	return matches[1], nil
}

// resolveVideoURL tries multiple strategies to find the direct video URL for a
// reel. Anonymous access to Instagram is heavily rate-limited/gated, so an
// authenticated session cookie (IG_SESSION_ID) is used first when available.
func resolveVideoURL(client *http.Client, shortcode string) (string, error) {
	sessionID := os.Getenv("IG_SESSION_ID")
	var errs []string

	if sessionID != "" {
		videoURL, err := fetchVideoURLFromAPI(client, shortcode, sessionID)
		if err == nil {
			fmt.Println("Resolved video URL via authenticated media API")
			return videoURL, nil
		}
		errs = append(errs, "media API: "+err.Error())
	}

	videoURL, err := fetchVideoURLFromGraphQL(client, shortcode, sessionID)
	if err == nil {
		fmt.Println("Resolved video URL via GraphQL")
		return videoURL, nil
	}
	errs = append(errs, "GraphQL: "+err.Error())

	videoURL, err = fetchVideoURLFromEmbed(client, shortcode)
	if err == nil {
		fmt.Println("Resolved video URL via embed page")
		return videoURL, nil
	}
	errs = append(errs, "embed page: "+err.Error())

	hint := ""
	if sessionID == "" {
		hint = " (anonymous access is often blocked by Instagram; consider setting IG_SESSION_ID to the sessionid cookie of a logged-in account)"
	}
	return "", fmt.Errorf("all strategies failed%s: %s", hint, strings.Join(errs, "; "))
}

// fetchVideoURLFromAPI uses the private web API with an authenticated session
// cookie. This is the most reliable method but requires IG_SESSION_ID.
func fetchVideoURLFromAPI(client *http.Client, shortcode, sessionID string) (string, error) {
	mediaID, err := shortcodeToMediaID(shortcode)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.instagram.com/api/v1/media/"+mediaID+"/info/", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("X-IG-App-ID", igAppID)
	req.Header.Set("Referer", "https://www.instagram.com/reel/"+shortcode+"/")
	req.AddCookie(&http.Cookie{Name: "sessionid", Value: sessionID})

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("media API returned status %d", resp.StatusCode)
	}

	var result apiMediaInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode media API response: %w", err)
	}
	if len(result.Items) == 0 || len(result.Items[0].VideoVersions) == 0 {
		return "", fmt.Errorf("media API response contains no video")
	}
	return result.Items[0].VideoVersions[0].URL, nil
}

// fetchVideoURLFromGraphQL replays the web client's GraphQL post query. It
// first loads the reel page to obtain the csrf cookie and LSD token.
func fetchVideoURLFromGraphQL(client *http.Client, shortcode, sessionID string) (string, error) {
	pageReq, err := http.NewRequest(http.MethodGet, "https://www.instagram.com/reel/"+shortcode+"/", nil)
	if err != nil {
		return "", err
	}
	pageReq.Header.Set("User-Agent", browserUserAgent)
	if sessionID != "" {
		pageReq.AddCookie(&http.Cookie{Name: "sessionid", Value: sessionID})
	}

	pageResp, err := client.Do(pageReq)
	if err != nil {
		return "", fmt.Errorf("failed to load reel page: %w", err)
	}
	pageBody, err := io.ReadAll(io.LimitReader(pageResp.Body, 4<<20))
	pageResp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read reel page: %w", err)
	}

	lsd := ""
	if m := lsdTokenRegex.FindSubmatch(pageBody); m != nil {
		lsd = string(m[1])
	}
	csrf := ""
	igURL, _ := url.Parse("https://www.instagram.com/")
	for _, c := range client.Jar.Cookies(igURL) {
		if c.Name == "csrftoken" {
			csrf = c.Value
		}
	}

	variables, _ := json.Marshal(map[string]interface{}{
		"shortcode":               shortcode,
		"fetch_tagged_user_count": nil,
		"hoisted_comment_id":      nil,
		"hoisted_reply_id":        nil,
	})
	form := url.Values{
		"variables":                {string(variables)},
		"doc_id":                   {graphQLDocID},
		"lsd":                      {lsd},
		"server_timestamps":        {"true"},
		"fb_api_caller_class":      {"RelayModern"},
		"fb_api_req_friendly_name": {"PolarisPostActionLoadPostQueryQuery"},
	}

	req, err := http.NewRequest(http.MethodPost, "https://www.instagram.com/graphql/query", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-IG-App-ID", igAppID)
	req.Header.Set("X-FB-LSD", lsd)
	req.Header.Set("X-CSRFToken", csrf)
	req.Header.Set("X-ASBD-ID", "129477")
	req.Header.Set("X-FB-Friendly-Name", "PolarisPostActionLoadPostQueryQuery")
	req.Header.Set("Origin", "https://www.instagram.com")
	req.Header.Set("Referer", "https://www.instagram.com/reel/"+shortcode+"/")
	if sessionID != "" {
		req.AddCookie(&http.Cookie{Name: "sessionid", Value: sessionID})
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GraphQL returned status %d", resp.StatusCode)
	}

	var result graphQLMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode GraphQL response: %w", err)
	}
	media := result.Data.ShortcodeMedia
	if media == nil {
		return "", fmt.Errorf("GraphQL response contains no media (access likely blocked)")
	}
	if !media.IsVideo || media.VideoURL == "" {
		return "", fmt.Errorf("post is not a video")
	}
	return media.VideoURL, nil
}

// fetchVideoURLFromEmbed scrapes the captioned embed page as a last resort.
func fetchVideoURLFromEmbed(client *http.Client, shortcode string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.instagram.com/reel/"+shortcode+"/embed/captioned/", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Referer", "https://www.instagram.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("embed page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}

	m := embedVideoRegex.FindSubmatch(body)
	if m == nil {
		return "", fmt.Errorf("no video URL found in embed page")
	}
	// The URL is JSON-escaped inside the HTML (\/ and \uXXXX sequences).
	var videoURL string
	if err := json.Unmarshal([]byte(`"`+string(m[1])+`"`), &videoURL); err != nil {
		return "", fmt.Errorf("failed to unescape embed video URL: %w", err)
	}
	return videoURL, nil
}

func downloadVideo(client *http.Client, videoURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, videoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Referer", "https://www.instagram.com/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("video URL returned status %d", resp.StatusCode)
	}

	videoBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read video data: %w", err)
	}
	return videoBytes, nil
}

// shortcodeToMediaID converts a post shortcode to the numeric media ID used by
// the private API (base64url digits, most significant first).
func shortcodeToMediaID(shortcode string) (string, error) {
	id := new(big.Int)
	sixtyFour := big.NewInt(64)
	for _, c := range shortcode {
		idx := strings.IndexRune(shortcodeAlpha, c)
		if idx < 0 {
			return "", fmt.Errorf("invalid character %q in shortcode %s", c, shortcode)
		}
		id.Mul(id, sixtyFour)
		id.Add(id, big.NewInt(int64(idx)))
	}
	return id.String(), nil
}
