"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { api, type Campaign, type ChannelAuthoredComment } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function MyCommentsTestPage() {
  const filterChannelId = "UCf7PZSTHrz2c5b8TAxYryZw";
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, channel } = useAuth();

  const [videoId, setVideoId] = useState("");
  const [comments, setComments] = useState<ChannelAuthoredComment[]>([]);
  const [loadingComments, setLoadingComments] = useState(false);
  const [commentsError, setCommentsError] = useState<string | null>(null);

  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [selectedCampaignId, setSelectedCampaignId] = useState("");
  const [creatingForCommentId, setCreatingForCommentId] = useState<string | null>(
    null
  );
  const [dealMessage, setDealMessage] = useState<string | null>(null);

  const canCreateDeals = user?.role === "advertiser";
  const signedInChannelId = channel?.channel_id || "";

  const fetchComments = useCallback(async (targetVideoId: string) => {
    if (!targetVideoId) return;

    setLoadingComments(true);
    setCommentsError(null);
    setDealMessage(null);

    try {
      const data = await api.listAllCommentsByVideo(targetVideoId);
      setComments(data);
    } catch (err) {
      setComments([]);
      setCommentsError(
        err instanceof Error ? err.message : "Failed to load comments"
      );
    } finally {
      setLoadingComments(false);
    }
  }, []);

  useEffect(() => {
    if (!canCreateDeals) return;

    api
      .listCampaigns()
      .then((data) => {
        setCampaigns(data);
        if (data.length > 0) {
          setSelectedCampaignId(data[0].id);
        }
      })
      .catch(() => {});
  }, [canCreateDeals]);

  useEffect(() => {
    const fromQuery = searchParams.get("videoId");
    if (fromQuery) {
      setVideoId(fromQuery);
      fetchComments(fromQuery);
    }
  }, [searchParams, fetchComments]);

  const shareUrl = useMemo(() => {
    if (typeof window === "undefined" || !videoId) return "";
    const url = new URL(window.location.href);
    url.searchParams.set("videoId", videoId);
    return url.toString();
  }, [videoId]);

  const filteredComments = useMemo(() => {
    // Keep advertiser test flow broad; only commenters are constrained to their own channel.
    if (user?.role !== "commenter" || !signedInChannelId) return comments;
    return comments.filter(
      (comment) => comment.author_channel_id === signedInChannelId
    );
  }, [comments, signedInChannelId, user?.role]);

  const handleSearch = () => {
    const next = videoId.trim();
    if (!next) return;
    router.replace(`/test/my-comments?videoId=${encodeURIComponent(next)}`);
    fetchComments(next);
  };

  const handleCreateDeal = async (comment: ChannelAuthoredComment) => {
    if (!selectedCampaignId) {
      setDealMessage("Select a campaign first.");
      return;
    }

    setCreatingForCommentId(comment.comment_id);
    setDealMessage(null);

    try {
      let marketplaceCommentID = comment.marketplace_comment_id;
      if (!marketplaceCommentID) {
        const registered = await api.registerCommentForTesting({
          comment_id: comment.comment_id,
          video_id: comment.video_id,
          author_channel_id: comment.author_channel_id,
          author_display_name: comment.author_display_name,
          text: comment.text,
          like_count: comment.like_count,
        });
        marketplaceCommentID = registered.id;
      }

      await api.createDeal(selectedCampaignId, marketplaceCommentID);
      setComments((prev) =>
        prev.map((row) =>
          row.comment_id === comment.comment_id
            ? {
                ...row,
                marketplace_comment_id: marketplaceCommentID,
              }
            : row
        )
      );
      setDealMessage(
        "Deal created. The verifier worker will edit the comment and release payment after verification."
      );
    } catch (err) {
      setDealMessage(
        err instanceof Error ? err.message : "Failed to create deal"
      );
    } finally {
      setCreatingForCommentId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">My Comments Test Page</h1>
        <p className="text-muted-foreground">
          Load all comments for a YouTube video, sorted by relevance.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Load comments by video ID</CardTitle>
          <CardDescription>
            Paste a YouTube video ID and browse all relevant comments.
            As an advertiser, create a deal from any listed row.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-col gap-3 sm:flex-row">
            <Input
              placeholder="YouTube video ID (e.g. dQw4w9WgXcQ)"
              value={videoId}
              onChange={(e) => setVideoId(e.target.value)}
            />
            <Button onClick={handleSearch} disabled={!videoId.trim()}>
              Load Comments
            </Button>
          </div>

          {shareUrl && (
            <p className="text-xs text-muted-foreground">
              Share link for advertisers: {shareUrl}
            </p>
          )}
        </CardContent>
      </Card>

      {canCreateDeals && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Create paid edit deal</CardTitle>
            <CardDescription>
              Pick a campaign, then click &ldquo;Pay &amp; Request Edit&rdquo; on any
              comment. If it is not in the marketplace yet, it will be registered
              automatically for testing.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {campaigns.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No campaigns found. Create and fund a campaign first.
              </p>
            ) : (
              <select
                className="h-10 w-full rounded-md border bg-background px-3 text-sm"
                value={selectedCampaignId}
                onChange={(e) => setSelectedCampaignId(e.target.value)}
              >
                {campaigns.map((campaign) => (
                  <option key={campaign.id} value={campaign.id}>
                    {campaign.ad_text} - ${(campaign.price_per_placement / 100).toFixed(2)}
                    /placement ({campaign.status})
                  </option>
                ))}
              </select>
            )}
          </CardContent>
        </Card>
      )}

      {dealMessage && (
        <Card>
          <CardContent className="py-4 text-sm">{dealMessage}</CardContent>
        </Card>
      )}

      {loadingComments ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            Loading comments...
          </CardContent>
        </Card>
      ) : commentsError ? (
        <Card>
          <CardContent className="py-8 text-center text-destructive">
            {commentsError}
          </CardContent>
        </Card>
      ) : filteredComments.length === 0 ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            {signedInChannelId
              ? "No comments from your signed-in channel were found for this video."
              : "No comments found from YouTube for this video."}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filteredComments.map((comment) => (
            <Card key={comment.comment_id}>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between gap-2">
                  <CardTitle className="truncate text-sm">
                    {comment.author_display_name}
                  </CardTitle>
                  <Badge
                    variant={
                      comment.marketplace_status === "available"
                        ? "default"
                        : "secondary"
                    }
                  >
                    {comment.marketplace_comment_id
                      ? comment.marketplace_status || "listed"
                      : "not listed"}
                  </Badge>
                </div>
                <CardDescription className="text-xs">
                  {comment.like_count.toLocaleString()} likes
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="line-clamp-4 text-sm">{comment.text}</p>
                {canCreateDeals && (
                  <Button
                    size="sm"
                    className="w-full"
                    disabled={
                      !selectedCampaignId ||
                      creatingForCommentId === comment.comment_id
                    }
                    onClick={() => handleCreateDeal(comment)}
                  >
                    {creatingForCommentId === comment.comment_id
                      ? "Creating deal..."
                      : "Pay & Request Edit"}
                  </Button>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
