package dataportability

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const baseURL = "https://dataportability.googleapis.com/v1"

// Client calls the Google Data Portability API (initiate + get state).
type Client struct {
	httpClient  *http.Client
	tokenSource oauth2.TokenSource
}

// NewClient returns a client that uses the given token source for auth (token is refreshed automatically).
func NewClient(tokenSource oauth2.TokenSource) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		tokenSource: tokenSource,
	}
}

func (c *Client) getToken(ctx context.Context) (*oauth2.Token, error) {
	return c.tokenSource.Token()
}

// InitiateRequest is the request body for portabilityArchive:initiate.
type InitiateRequest struct {
	Resources []string `json:"resources"`
	StartTime string   `json:"startTime,omitempty"`
	EndTime   string   `json:"endTime,omitempty"`
}

// InitiateResponse is the response from portabilityArchive:initiate.
type InitiateResponse struct {
	ArchiveJobID string `json:"archiveJobId"`
	AccessType   string `json:"accessType"`
}

// InitiateExport starts a YouTube comments export for the given time range.
// Resource name per schema: youtube.comments. startTime/endTime in RFC3339.
func (c *Client) InitiateExport(ctx context.Context, startTime, endTime time.Time) (jobID string, err error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	reqBody := InitiateRequest{
		Resources: []string{"youtube.comments"},
		StartTime: startTime.UTC().Format(time.RFC3339),
		EndTime:   endTime.UTC().Format(time.RFC3339),
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/portabilityArchive:initiate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("initiate request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("initiate %d: %s", resp.StatusCode, string(slurp))
	}

	var out InitiateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ArchiveJobID, nil
}

// State values from getPortabilityArchiveState.
const (
	StateUnspecified = "STATE_UNSPECIFIED"
	StateInProgress  = "IN_PROGRESS"
	StateComplete   = "COMPLETE"
	StateFailed     = "FAILED"
	StateCancelled  = "CANCELLED"
)

// ArchiveState is the response from getPortabilityArchiveState.
type ArchiveState struct {
	State     string   `json:"state"`
	URLs      []string `json:"urls"`
	Name      string   `json:"name"`
	StartTime string   `json:"startTime"`
	ExportTime string  `json:"exportTime"`
}

// GetArchiveState returns the job state and, when COMPLETE, signed download URLs.
func (c *Client) GetArchiveState(ctx context.Context, jobID string) (*ArchiveState, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}
	path := fmt.Sprintf("%s/archiveJobs/%s/portabilityArchiveState", baseURL, jobID)
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get state %d: %s", resp.StatusCode, string(slurp))
	}

	var out ArchiveState
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UserComment matches the shape we return to the frontend.
type UserComment struct {
	CommentID   string    `json:"comment_id"`
	VideoID     string    `json:"video_id"`
	TextDisplay string    `json:"text_display"`
	LikeCount   int64     `json:"like_count"`
	PublishedAt time.Time `json:"published_at"`
}

// DownloadAndParseComments fetches the first signed URL and parses YouTube comment CSV.
// Schema: Channel ID, Comment ID, Video ID, Comment Text (and others). We map to UserComment.
func (c *Client) DownloadAndParseComments(ctx context.Context, signedURL string) ([]UserComment, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", signedURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download archive: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download %d: %s", resp.StatusCode, string(slurp))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Archive may be raw CSV or ZIP. Try CSV first (schema says comment object exported as CSV).
	if bytes.HasPrefix(data, []byte("Channel ID")) || bytes.Contains(data, []byte("Comment ID")) {
		return parseCommentCSV(data)
	}
	// If ZIP, we'd need archive/zip; for now assume single CSV or JSON
	return parseCommentCSV(data)
}

// parseCommentCSV parses CSV with columns per YouTube comment schema.
// Schema fields: Channel ID, Comment ID, Video ID, Comment Text (repeated), Parent Comment ID, Post ID, ...
func parseCommentCSV(data []byte) ([]UserComment, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}
	if len(rows) < 2 {
		return nil, nil
	}

	header := rows[0]
	idxCommentID := -1
	idxVideoID := -1
	idxCommentText := -1
	for i, h := range header {
		h = strings.TrimSpace(h)
		switch h {
		case "Comment ID":
			idxCommentID = i
		case "Video ID":
			idxVideoID = i
		case "Comment Text", "Comment Text Segments":
			idxCommentText = i
		}
	}
	if idxCommentID == -1 || idxVideoID == -1 {
		return nil, fmt.Errorf("csv missing required columns (Comment ID, Video ID)")
	}
	if idxCommentText == -1 {
		idxCommentText = idxCommentID // fallback
	}

	var out []UserComment
	for _, row := range rows[1:] {
		if idxCommentID >= len(row) || idxVideoID >= len(row) {
			continue
		}
		commentID := strings.TrimSpace(row[idxCommentID])
		videoID := strings.TrimSpace(row[idxVideoID])
		text := ""
		if idxCommentText < len(row) {
			text = strings.TrimSpace(row[idxCommentText])
		}
		if commentID == "" && videoID == "" {
			continue
		}
		out = append(out, UserComment{
			CommentID:   commentID,
			VideoID:     videoID,
			TextDisplay: text,
			LikeCount:   0,
			PublishedAt: time.Now(),
		})
	}
	return out, nil
}
