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
	VideoID           string
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
			VideoID:           item.Snippet.VideoId,
			AuthorChannelID:   authorChID,
			AuthorDisplayName: snip.AuthorDisplayName,
			TextDisplay:       snip.TextDisplay,
			LikeCount:         int64(snip.LikeCount),
		})
	}
	return comments, nil
}

func (c *Client) FetchCommentsByAuthorOnChannel(ctx context.Context, channelID string) ([]Comment, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}

	var (
		comments      []Comment
		nextPageToken string
	)

	// Paginate deeply so this endpoint can serve "all" available comments for testing.
	for page := 0; page < 50; page++ {
		call := svc.CommentThreads.List([]string{"snippet"}).
			AllThreadsRelatedToChannelId(channelID).
			MaxResults(100).
			TextFormat("plainText")

		if nextPageToken != "" {
			call = call.PageToken(nextPageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("fetch comments for channel %s: %w", channelID, err)
		}

		for _, item := range resp.Items {
			if item.Snippet == nil || item.Snippet.TopLevelComment == nil || item.Snippet.TopLevelComment.Snippet == nil {
				continue
			}

			top := item.Snippet.TopLevelComment
			snip := top.Snippet
			authorChID := ""
			if snip.AuthorChannelId != nil {
				authorChID = snip.AuthorChannelId.Value
			}

			if authorChID != channelID {
				continue
			}

			comments = append(comments, Comment{
				CommentID:         top.Id,
				VideoID:           item.Snippet.VideoId,
				AuthorChannelID:   authorChID,
				AuthorDisplayName: snip.AuthorDisplayName,
				TextDisplay:       snip.TextDisplay,
				LikeCount:         int64(snip.LikeCount),
			})
		}

		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}

	return comments, nil
}

func (c *Client) FetchCommentsAuthoredOnOwnChannel(ctx context.Context, channelID string) ([]Comment, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}

	chResp, err := svc.Channels.List([]string{"contentDetails"}).Id(channelID).Do()
	if err != nil {
		return nil, fmt.Errorf("fetch channel %s: %w", channelID, err)
	}
	if len(chResp.Items) == 0 || chResp.Items[0].ContentDetails == nil || chResp.Items[0].ContentDetails.RelatedPlaylists == nil {
		return nil, nil
	}

	uploadsPlaylist := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	if uploadsPlaylist == "" {
		return nil, nil
	}

	videoIDs, err := c.fetchUploadVideoIDs(ctx, svc, uploadsPlaylist, 200)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	comments := make([]Comment, 0, 128)

	for _, videoID := range videoIDs {
		videoComments, err := c.fetchVideoThreadCommentsByAuthor(ctx, svc, channelID, videoID)
		if err != nil {
			continue
		}

		for _, cm := range videoComments {
			if _, exists := seen[cm.CommentID]; exists {
				continue
			}
			seen[cm.CommentID] = struct{}{}
			comments = append(comments, cm)
		}
	}

	return comments, nil
}

func (c *Client) FetchAllCommentsByVideo(ctx context.Context, videoID string) ([]Comment, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}

	return c.fetchVideoThreadCommentsByRelevance(ctx, svc, videoID, 0)
}

func (c *Client) fetchUploadVideoIDs(ctx context.Context, svc *yt.Service, playlistID string, maxVideos int) ([]string, error) {
	videoIDs := make([]string, 0, maxVideos)
	pageToken := ""

	for len(videoIDs) < maxVideos {
		remaining := maxVideos - len(videoIDs)
		pageSize := int64(50)
		if remaining < 50 {
			pageSize = int64(remaining)
		}

		call := svc.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(playlistID).
			MaxResults(pageSize)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("fetch upload playlist items: %w", err)
		}

		for _, item := range resp.Items {
			if item.Snippet == nil || item.Snippet.ResourceId == nil {
				continue
			}
			if item.Snippet.ResourceId.VideoId == "" {
				continue
			}
			videoIDs = append(videoIDs, item.Snippet.ResourceId.VideoId)
			if len(videoIDs) >= maxVideos {
				break
			}
		}

		if resp.NextPageToken == "" || len(videoIDs) >= maxVideos {
			break
		}
		pageToken = resp.NextPageToken
	}

	return videoIDs, nil
}

func (c *Client) fetchVideoThreadCommentsByAuthor(ctx context.Context, svc *yt.Service, authorChannelID, videoID string) ([]Comment, error) {
	all, err := c.fetchVideoThreadComments(ctx, svc, videoID)
	if err != nil {
		return nil, err
	}

	filtered := make([]Comment, 0, len(all))
	for _, cm := range all {
		if cm.AuthorChannelID == authorChannelID {
			filtered = append(filtered, cm)
		}
	}

	return filtered, nil
}

func (c *Client) fetchVideoThreadCommentsByRelevance(ctx context.Context, svc *yt.Service, videoID string, maxResults int64) ([]Comment, error) {
	var (
		out       []Comment
		pageToken string
		remaining = maxResults
	)

	for {
		pageSize := int64(100)
		if remaining > 0 && remaining < pageSize {
			pageSize = remaining
		}

		call := svc.CommentThreads.List([]string{"snippet"}).
			VideoId(videoID).
			Order("relevance").
			MaxResults(pageSize).
			TextFormat("plainText")
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Context(ctx).Do()
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			if item.Snippet == nil || item.Snippet.TopLevelComment == nil || item.Snippet.TopLevelComment.Snippet == nil {
				continue
			}

			top := item.Snippet.TopLevelComment
			snip := top.Snippet
			topAuthorChID := ""
			if snip.AuthorChannelId != nil {
				topAuthorChID = snip.AuthorChannelId.Value
			}

			out = append(out, Comment{
				CommentID:         top.Id,
				VideoID:           videoID,
				AuthorChannelID:   topAuthorChID,
				AuthorDisplayName: snip.AuthorDisplayName,
				TextDisplay:       snip.TextDisplay,
				LikeCount:         int64(snip.LikeCount),
			})
		}

		if remaining > 0 {
			remaining -= int64(len(resp.Items))
			if remaining <= 0 {
				break
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return out, nil
}

func (c *Client) fetchVideoThreadComments(ctx context.Context, svc *yt.Service, videoID string) ([]Comment, error) {
	var (
		out       []Comment
		pageToken string
	)

	for page := 0; page < 20; page++ {
		call := svc.CommentThreads.List([]string{"snippet", "replies"}).
			VideoId(videoID).
			Order("time").
			MaxResults(100).
			TextFormat("plainText")
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Context(ctx).Do()
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			if item.Snippet == nil || item.Snippet.TopLevelComment == nil || item.Snippet.TopLevelComment.Snippet == nil {
				continue
			}

			top := item.Snippet.TopLevelComment
			snip := top.Snippet
			topAuthorChID := ""
			if snip.AuthorChannelId != nil {
				topAuthorChID = snip.AuthorChannelId.Value
			}
			out = append(out, Comment{
				CommentID:         top.Id,
				VideoID:           videoID,
				AuthorChannelID:   topAuthorChID,
				AuthorDisplayName: snip.AuthorDisplayName,
				TextDisplay:       snip.TextDisplay,
				LikeCount:         int64(snip.LikeCount),
			})

			if item.Replies != nil {
				for _, reply := range item.Replies.Comments {
					if reply == nil || reply.Snippet == nil {
						continue
					}
					rs := reply.Snippet
					replyAuthorChID := ""
					if rs.AuthorChannelId != nil {
						replyAuthorChID = rs.AuthorChannelId.Value
					}
					out = append(out, Comment{
						CommentID:         reply.Id,
						VideoID:           videoID,
						AuthorChannelID:   replyAuthorChID,
						AuthorDisplayName: rs.AuthorDisplayName,
						TextDisplay:       rs.TextDisplay,
						LikeCount:         int64(rs.LikeCount),
					})
				}
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return out, nil
}

func (c *Client) FetchVideo(ctx context.Context, videoID string) (*TrendingVideo, error) {
	svc, err := yt.NewService(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}

	resp, err := svc.Videos.List([]string{"snippet", "statistics"}).Id(videoID).Do()
	if err != nil {
		return nil, fmt.Errorf("fetch video %s: %w", videoID, err)
	}
	if len(resp.Items) == 0 || resp.Items[0].Snippet == nil {
		return nil, fmt.Errorf("video %s not found", videoID)
	}

	item := resp.Items[0]
	thumb := ""
	if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Medium != nil {
		thumb = item.Snippet.Thumbnails.Medium.Url
	}

	return &TrendingVideo{
		VideoID:      item.Id,
		Title:        item.Snippet.Title,
		ChannelTitle: item.Snippet.ChannelTitle,
		ThumbnailURL: thumb,
		ViewCount:    item.Statistics.ViewCount,
	}, nil
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
