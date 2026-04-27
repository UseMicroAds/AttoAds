package youtube

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestFetchTrendingVideos_Integration(t *testing.T) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		t.Skip("set YOUTUBE_API_KEY to run integration test")
	}

	client := NewClient(apiKey)
	videos, err := client.FetchTrendingVideos(context.Background(), "US", 5)
	if err != nil {
		t.Fatalf("FetchTrendingVideos failed: %v", err)
	}

	raw, err := json.MarshalIndent(videos, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal trending videos response: %v", err)
	}
	t.Logf("raw response (prettified):\n%s", string(raw))

	t.Logf("received %d videos", len(videos))
	for i, v := range videos {
		t.Logf("%d) id=%s title=%q channel=%q views=%d",
			i+1, v.VideoID, v.Title, v.ChannelTitle, v.ViewCount)
	}
}
