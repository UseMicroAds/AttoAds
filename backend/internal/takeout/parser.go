package takeout

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/attoads/attoads-backend/internal/models"
)

// ParseTakeoutZip opens the ZIP at filePath and extracts YouTube comments from known Takeout layouts.
// Supports: Youtube/comments/comments.csv, YouTube and YouTube Music/my-comments/*.html, and activity JSON.
func ParseTakeoutZip(zipPath string) ([]models.CommentExportRow, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	var all []models.CommentExportRow
	seen := make(map[string]bool)

	for _, f := range r.File {
		base := strings.ToLower(path.Base(f.Name))
		dir := path.Dir(f.Name)

		// CSV: "Youtube/comments/comments.csv" or "Takeout/Youtube/comments/comments.csv"
		if strings.HasSuffix(base, ".csv") && (strings.Contains(dir, "comment") || strings.Contains(base, "comment")) {
			rows, err := parseCSVFromZip(f)
			if err != nil {
				continue
			}
			for _, row := range rows {
				key := row.CommentID + "|" + row.VideoID
				if !seen[key] {
					seen[key] = true
					all = append(all, row)
				}
			}
		}

		// HTML: "YouTube and YouTube Music/my-comments/*.html"
		if strings.HasSuffix(base, ".html") && strings.Contains(dir, "comment") {
			rows, err := parseHTMLFromZip(f)
			if err != nil {
				continue
			}
			for _, row := range rows {
				key := row.CommentID + "|" + row.VideoID
				if !seen[key] {
					seen[key] = true
					all = append(all, row)
				}
			}
		}

		// JSON activity: "YouTube and YouTube Music/My activity/My activity.json" or similar
		if strings.HasSuffix(base, ".json") && (strings.Contains(dir, "youtube") || strings.Contains(dir, "YouTube")) {
			rows, err := parseJSONActivityFromZip(f)
			if err != nil {
				continue
			}
			for _, row := range rows {
				if row.VideoID == "" {
					continue
				}
				key := row.CommentID + "|" + row.VideoID
				if !seen[key] {
					seen[key] = true
					all = append(all, row)
				}
			}
		}
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("no YouTube comments found in Takeout zip (looked for comments.csv, my-comments/*.html, or activity JSON)")
	}
	return all, nil
}

func parseCSVFromZip(f *zip.File) ([]models.CommentExportRow, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 {
		return nil, fmt.Errorf("invalid csv")
	}

	header := rows[0]
	idxCommentID, idxVideoID, idxText, idxTime := -1, -1, -1, -1
	for i, h := range header {
		h = strings.TrimSpace(strings.Trim(strings.ToLower(h), "\""))
		switch h {
		case "comment id", "comment_id", "id":
			idxCommentID = i
		case "video id", "video_id", "videoid":
			idxVideoID = i
		case "comment text", "text", "content", "comment":
			idxText = i
		case "time", "timestamp", "published", "date":
			idxTime = i
		}
	}
	if idxVideoID == -1 {
		return nil, fmt.Errorf("csv missing video id column")
	}
	if idxCommentID == -1 {
		idxCommentID = idxVideoID
	}
	if idxText == -1 {
		idxText = idxCommentID
	}

	var out []models.CommentExportRow
	for _, row := range rows[1:] {
		if idxVideoID >= len(row) {
			continue
		}
		videoID := strings.TrimSpace(strings.Trim(row[idxVideoID], "\""))
		if videoID == "" {
			continue
		}
		commentID := ""
		if idxCommentID < len(row) {
			commentID = strings.TrimSpace(strings.Trim(row[idxCommentID], "\""))
		}
		if commentID == "" {
			commentID = videoID + "_" + fmt.Sprintf("%d", len(out))
		}
		text := ""
		if idxText < len(row) {
			text = strings.TrimSpace(strings.Trim(row[idxText], "\""))
		}
		pubAt := time.Now()
		if idxTime < len(row) {
			if t, err := time.Parse(time.RFC3339, strings.Trim(row[idxTime], "\"")); err == nil {
				pubAt = t
			} else if t, err := time.Parse("2006-01-02 15:04:05 UTC", strings.Trim(row[idxTime], "\"")); err == nil {
				pubAt = t
			}
		}
		out = append(out, models.CommentExportRow{
			CommentID:   commentID,
			VideoID:     videoID,
			TextDisplay: text,
			LikeCount:   0,
			PublishedAt: pubAt,
		})
	}
	return out, nil
}

var (
	videoIDRe   = regexp.MustCompile(`youtube\.com/watch\?v=([a-zA-Z0-9_-]{11})`)
	timeRe      = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2}):(\d{2})`)
	commentIDRe = regexp.MustCompile(`comment_id=([a-zA-Z0-9_-]+)`)
)

func parseHTMLFromZip(f *zip.File) ([]models.CommentExportRow, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	s := string(data)

	var out []models.CommentExportRow
	videoIDs := videoIDRe.FindAllStringSubmatch(s, -1)
	times := timeRe.FindAllStringSubmatch(s, -1)
	commentIDs := commentIDRe.FindAllStringSubmatch(s, -1)

	for i, m := range videoIDs {
		if len(m) < 2 {
			continue
		}
		videoID := m[1]
		commentID := ""
		if i < len(commentIDs) && len(commentIDs[i]) >= 2 {
			commentID = commentIDs[i][1]
		}
		if commentID == "" {
			commentID = videoID + "_" + fmt.Sprintf("%d", i)
		}
		pubAt := time.Now()
		if i < len(times) && len(times[i]) >= 7 {
			// timeRe has submatches for year, month, day, hour, min, sec
			tstr := times[i][0]
			if t, err := time.Parse("2006-01-02 15:04:05", tstr); err == nil {
				pubAt = t
			}
		}
		out = append(out, models.CommentExportRow{
			CommentID:   commentID,
			VideoID:     videoID,
			TextDisplay: "",
			LikeCount:   0,
			PublishedAt: pubAt,
		})
	}
	return out, nil
}

type takeoutActivity struct {
	Title    string `json:"title"`
	TitleURL string `json:"titleUrl"`
	Time     string `json:"time"`
	Details  []struct {
		Comment string `json:"comment"`
	} `json:"details"`
}

func parseJSONActivityFromZip(f *zip.File) ([]models.CommentExportRow, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	var items []takeoutActivity
	if bytes.HasPrefix(data, []byte("[")) {
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, err
		}
	} else {
		var single takeoutActivity
		if err := json.Unmarshal(data, &single); err != nil {
			return nil, err
		}
		items = []takeoutActivity{single}
	}

	var out []models.CommentExportRow
	for i, item := range items {
		if item.TitleURL == "" || !strings.Contains(item.TitleURL, "youtube.com/watch") {
			continue
		}
		videoID := ""
		for _, m := range videoIDRe.FindAllStringSubmatch(item.TitleURL, -1) {
			if len(m) >= 2 {
				videoID = m[1]
				break
			}
		}
		if videoID == "" {
			continue
		}
		commentID := videoID + "_" + fmt.Sprintf("%d", i)
		text := ""
		if len(item.Details) > 0 {
			text = item.Details[0].Comment
		}
		pubAt := time.Now()
		if item.Time != "" {
			if t, err := time.Parse(time.RFC3339, item.Time); err == nil {
				pubAt = t
			}
		}
		out = append(out, models.CommentExportRow{
			CommentID:   commentID,
			VideoID:     videoID,
			TextDisplay: text,
			LikeCount:   0,
			PublishedAt: pubAt,
		})
	}
	return out, nil
}
