package youtube

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

type TrendingVideo struct {
	VideoID      string
	Title        string
	ChannelTitle string
	ThumbnailURL string
	ViewCount    uint64
}

type Comment struct {
	CommentID         string
	AuthorChannelID   string
	AuthorDisplayName string
	TextDisplay       string
	LikeCount         int64
}

func (c *Client) FetchTrendingVideos(ctx context.Context, regionCode string, maxResults int64) ([]TrendingVideo, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}

	resp, err := svc.Videos.List([]string{"snippet", "statistics"}).
		Chart("mostPopular").
		RegionCode(regionCode).
		MaxResults(maxResults).
		Do()
	if err != nil {
		return nil, fmt.Errorf("fetch trending videos: %w", err)
	}

	videos := make([]TrendingVideo, 0, len(resp.Items))
	for _, item := range resp.Items {
		thumb := ""
		if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Medium != nil {
			thumb = item.Snippet.Thumbnails.Medium.Url
		}
		videos = append(videos, TrendingVideo{
			VideoID:      item.Id,
			Title:        item.Snippet.Title,
			ChannelTitle: item.Snippet.ChannelTitle,
			ThumbnailURL: thumb,
			ViewCount:    item.Statistics.ViewCount,
		})
	}
	return videos, nil
}

func (c *Client) FetchTopComments(ctx context.Context, videoID string, maxResults int64) ([]Comment, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}

	resp, err := svc.CommentThreads.List([]string{"snippet"}).
		VideoId(videoID).
		Order("relevance").
		MaxResults(maxResults).
		TextFormat("plainText").
		Do()
	if err != nil {
		return nil, fmt.Errorf("fetch comments for %s: %w", videoID, err)
	}

	comments := make([]Comment, 0, len(resp.Items))
	for _, item := range resp.Items {
		snip := item.Snippet.TopLevelComment.Snippet
		authorChID := ""
		if snip.AuthorChannelId != nil {
			authorChID = snip.AuthorChannelId.Value
		}
		comments = append(comments, Comment{
			CommentID:         item.Snippet.TopLevelComment.Id,
			AuthorChannelID:   authorChID,
			AuthorDisplayName: snip.AuthorDisplayName,
			TextDisplay:       snip.TextDisplay,
			LikeCount:         int64(snip.LikeCount),
		})
	}
	return comments, nil
}

func (c *Client) FetchCommentText(ctx context.Context, commentID string) (string, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return "", fmt.Errorf("create youtube service: %w", err)
	}

	resp, err := svc.Comments.List([]string{"snippet"}).
		Id(commentID).
		TextFormat("plainText").
		Do()
	if err != nil {
		return "", fmt.Errorf("fetch comment %s: %w", commentID, err)
	}
	if len(resp.Items) == 0 {
		return "", fmt.Errorf("comment %s not found", commentID)
	}

	return resp.Items[0].Snippet.TextDisplay, nil
}

func (c *Client) UpdateComment(ctx context.Context, token *oauth2.Token, oauthCfg *oauth2.Config, commentID, newText string) error {
	svc, err := yt.NewService(ctx, option.WithTokenSource(oauthCfg.TokenSource(ctx, token)))
	if err != nil {
		return fmt.Errorf("create youtube service: %w", err)
	}

	_, err = svc.Comments.Update([]string{"snippet"}, &yt.Comment{
		Id: commentID,
		Snippet: &yt.CommentSnippet{
			TextOriginal: newText,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("update comment %s: %w", commentID, err)
	}
	return nil
}
