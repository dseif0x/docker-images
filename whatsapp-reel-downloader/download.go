package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type IGDownloadResponse struct {
	Status string `json:"status"`
	Data   struct {
		Filename string `json:"filename"`
		VideoURL string `json:"videoUrl"`
	} `json:"data"`
}

// DownloadInstagramReelBytes downloads the Instagram reel video as a byte slice.
func DownloadInstagramReelBytes(link string) ([]byte, string, error) {
	apiURL := "https://instagram-reels-downloader-tau.vercel.app/api/video?postUrl=" + link

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	var result IGDownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode API response: %w", err)
	}

	if result.Status != "success" || result.Data.VideoURL == "" {
		return nil, "", fmt.Errorf("invalid API response or missing video URL")
	}

	videoResp, err := http.Get(result.Data.VideoURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download video: %w", err)
	}
	defer videoResp.Body.Close()

	if videoResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("video URL returned non-200 status: %d", videoResp.StatusCode)
	}

	videoBytes, err := io.ReadAll(videoResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read video data: %w", err)
	}

	return videoBytes, result.Data.Filename, nil
}
