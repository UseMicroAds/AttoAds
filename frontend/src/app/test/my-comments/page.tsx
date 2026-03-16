"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import {
  api,
  type Bounty,
  type Campaign,
  type ChannelAuthoredComment,
  type ViralComment,
} from "@/lib/api";
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

  // Bounty workflow (advertiser: create + list; commenter: hunt + claim)
  const [advertiserBounties, setAdvertiserBounties] = useState<Bounty[]>([]);
  const [bountyAdText, setBountyAdText] = useState("Test bounty ad text");
  const [bountyBudget, setBountyBudget] = useState("10");
  const [bountyAmount, setBountyAmount] = useState("5");
  const [bountyMinLikes, setBountyMinLikes] = useState("0");
  const [creatingBounty, setCreatingBounty] = useState(false);
  const [bountyCreateError, setBountyCreateError] = useState<string | null>(null);
  const [activeBounties, setActiveBounties] = useState<Bounty[]>([]);
  /** For each bounty id, the list of eligible viral comments (used to show which bounties my comment can claim) */
  const [eligibleByBountyId, setEligibleByBountyId] = useState<
    Record<string, ViralComment[]>
  >({});
  const [loadingEligibleByBounty, setLoadingEligibleByBounty] = useState(false);
  const [claimingForBounty, setClaimingForBounty] = useState<string | null>(
    null
  );
  const [bountyMessage, setBountyMessage] = useState<string | null>(null);
  /** Viral comment IDs that already have a deal (so commenter shouldn't see them as claimable again) */
  const [myDealCommentIds, setMyDealCommentIds] = useState<Set<string>>(
    () => new Set()
  );

  const canCreateDeals = user?.role === "advertiser";
  const isCommenter = user?.role === "commenter";
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
    if (user?.role !== "advertiser") return;
    api.listBounties().then(setAdvertiserBounties).catch(() => {});
  }, [user?.role]);

  useEffect(() => {
    if (!isCommenter) return;
    api
      .listActiveBounties()
      .then((list) => setActiveBounties(list.filter((b) => b.status !== "completed")))
      .catch(() => {});
  }, [isCommenter]);

  useEffect(() => {
    if (!isCommenter) return;
    api
      .listMyDeals()
      .then((deals) =>
        setMyDealCommentIds(new Set(deals.map((d) => d.comment_id)))
      )
      .catch(() => setMyDealCommentIds(new Set()));
  }, [isCommenter]);

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

  /** For commenters: exclude comments that already have a deal (already claimed / edited) so they can't hunt again. */
  const commentsAvailableForBounty = useMemo(() => {
    if (user?.role !== "commenter") return filteredComments;
    return filteredComments.filter(
      (c) =>
        !c.marketplace_comment_id || !myDealCommentIds.has(c.marketplace_comment_id)
    );
  }, [filteredComments, user?.role, myDealCommentIds]);

  /** For commenter: which active bounties can be claimed with this comment (comment must be in marketplace) */
  const bountiesForComment = useCallback(
    (comment: ChannelAuthoredComment): Bounty[] => {
      const marketplaceId = comment.marketplace_comment_id;
      if (!marketplaceId || !isCommenter) return [];
      return activeBounties.filter((b) =>
        (eligibleByBountyId[b.id] ?? []).some((vc) => vc.id === marketplaceId)
      );
    },
    [isCommenter, activeBounties, eligibleByBountyId]
  );

  /** Bounties this comment could qualify for by min_likes (for comments not yet in marketplace) */
  const bountiesQualifiableByLikes = useCallback(
    (comment: ChannelAuthoredComment): Bounty[] => {
      if (!isCommenter) return [];
      return activeBounties.filter((b) => comment.like_count >= b.min_likes);
    },
    [isCommenter, activeBounties]
  );

  // When commenter has loaded a video and has comments, fetch eligible comments per bounty
  useEffect(() => {
    if (
      !isCommenter ||
      !videoId ||
      activeBounties.length === 0 ||
      filteredComments.length === 0
    ) {
      setEligibleByBountyId({});
      return;
    }
    setLoadingEligibleByBounty(true);
    Promise.all(
      activeBounties.map((b) =>
        api
          .listEligibleCommentsForBounty(b.id)
          .then((list) => ({ bountyId: b.id, list }))
      )
    )
      .then((results) => {
        const next: Record<string, ViralComment[]> = {};
        for (const { bountyId, list } of results) {
          next[bountyId] = list;
        }
        setEligibleByBountyId(next);
      })
      .catch(() => setEligibleByBountyId({}))
      .finally(() => setLoadingEligibleByBounty(false));
  }, [isCommenter, videoId, activeBounties, filteredComments.length]);

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

  const handleCreateBounty = async (e: React.FormEvent) => {
    e.preventDefault();
    setBountyCreateError(null);
    const budgetCents = Math.round(parseFloat(bountyBudget) * 100);
    const amountCents = Math.round(parseFloat(bountyAmount) * 100);
    if (!bountyAdText.trim() || budgetCents <= 0 || amountCents <= 0) {
      setBountyCreateError("Ad text, budget, and amount per claim are required.");
      return;
    }
    if (amountCents > budgetCents) {
      setBountyCreateError("Amount per claim cannot exceed budget.");
      return;
    }
    setCreatingBounty(true);
    try {
      const bounty = await api.createBounty({
        ad_text: bountyAdText.trim(),
        budget_cents: budgetCents,
        amount_per_claim_cents: amountCents,
        min_likes: parseInt(bountyMinLikes, 10) || 0,
      });
      setAdvertiserBounties((prev) => [bounty, ...prev]);
      setBountyCreateError(null);
      router.push(`/bounties/${bounty.id}`);
    } catch (err) {
      setBountyCreateError(
        err instanceof Error ? err.message : "Failed to create bounty"
      );
    } finally {
      setCreatingBounty(false);
    }
  };

  const handleClaimBounty = async (
    bountyId: string,
    marketplaceCommentId: string | null,
    comment: ChannelAuthoredComment
  ) => {
    setBountyMessage(null);
    const claimingKey = marketplaceCommentId ?? `comment-${comment.comment_id}`;
    setClaimingForBounty(claimingKey);
    try {
      let viralCommentId = marketplaceCommentId;
      if (!viralCommentId) {
        const registered = await api.registerCommentForTesting({
          comment_id: comment.comment_id,
          video_id: comment.video_id,
          author_channel_id: comment.author_channel_id,
          author_display_name: comment.author_display_name,
          text: comment.text,
          like_count: comment.like_count,
        });
        viralCommentId = registered.id;
        setComments((prev) =>
          prev.map((row) =>
            row.comment_id === comment.comment_id
              ? { ...row, marketplace_comment_id: registered.id }
              : row
          )
        );
      }
      await api.claimBounty(bountyId, viralCommentId);
      setBountyMessage(
        "Claim submitted. The system will append the ad to your comment; you'll be paid after verification."
      );
      setEligibleByBountyId((prev) => ({
        ...prev,
        [bountyId]: (prev[bountyId] ?? []).filter(
          (c) => c.id !== viralCommentId
        ),
      }));
      setMyDealCommentIds((prev) => new Set(prev).add(viralCommentId));
    } catch (err) {
      setBountyMessage(
        err instanceof Error ? err.message : "Failed to claim bounty"
      );
    } finally {
      setClaimingForBounty(null);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">My Comments Test Page</h1>
        <p className="text-muted-foreground">
          {isCommenter
            ? "Enter a video ID to load comments. Comments from your channel are shown; if one meets a bounty's requirements, you can hunt the bounty there."
            : "Load all comments for a YouTube video. As an advertiser, create deals or register comments for testing."}
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Load comments by video ID</CardTitle>
          <CardDescription>
            {isCommenter
              ? "Paste a YouTube video ID. Your comments on that video will be listed; bounties you can claim appear under each comment."
              : "Paste a YouTube video ID and browse all relevant comments. As an advertiser, create a deal from any listed row."}
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

      {(canCreateDeals || isCommenter) && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Test bounty challenge workflow</CardTitle>
            <CardDescription>
              {isCommenter
                ? "Enter a video ID and load comments. Your comments from your channel appear below; if a comment is in the marketplace and meets a bounty's requirements, you can claim it there."
                : "Create and fund a bounty. Commenters can then claim it from their comments on the test page."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {canCreateDeals && (
              <div className="space-y-4">
                <h4 className="text-sm font-medium">Create test bounty</h4>
                <form onSubmit={handleCreateBounty} className="flex flex-wrap items-end gap-3">
                  <Input
                    placeholder="Ad text"
                    value={bountyAdText}
                    onChange={(e) => setBountyAdText(e.target.value)}
                    className="min-w-[180px]"
                  />
                  <Input
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="Budget $"
                    value={bountyBudget}
                    onChange={(e) => setBountyBudget(e.target.value)}
                    className="w-24"
                  />
                  <Input
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="Per claim $"
                    value={bountyAmount}
                    onChange={(e) => setBountyAmount(e.target.value)}
                    className="w-24"
                  />
                  <Input
                    type="number"
                    min="0"
                    placeholder="Min likes"
                    value={bountyMinLikes}
                    onChange={(e) => setBountyMinLikes(e.target.value)}
                    className="w-24"
                  />
                  <Button type="submit" disabled={creatingBounty}>
                    {creatingBounty ? "Creating…" : "Create bounty"}
                  </Button>
                </form>
                {bountyCreateError && (
                  <p className="text-sm text-destructive">{bountyCreateError}</p>
                )}
                {advertiserBounties.length > 0 && (
                  <div>
                    <p className="mb-2 text-sm font-medium">Your bounties</p>
                    <ul className="space-y-1 text-sm">
                      {advertiserBounties.map((b) => (
                        <li key={b.id} className="flex items-center gap-2">
                          <Link
                            href={`/bounties/${b.id}`}
                            className="text-primary underline hover:no-underline"
                          >
                            {b.ad_text.slice(0, 40)}
                            {b.ad_text.length > 40 ? "…" : ""}
                          </Link>
                          <Badge variant="secondary">{b.status}</Badge>
                          <span className="text-muted-foreground">
                            ${(b.amount_per_claim_cents / 100).toFixed(2)}/claim
                          </span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}

            {isCommenter && videoId && (
              <p className="text-sm text-muted-foreground">
                Your comments on this video are listed below. Bounties you can claim appear under each comment.
              </p>
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
      ) : commentsAvailableForBounty.length === 0 ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            {isCommenter && filteredComments.length > 0
              ? "All your comments on this video have already been claimed (or are in a deal)."
              : signedInChannelId
                ? "No comments from your signed-in channel were found for this video."
                : "No comments found from YouTube for this video."}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {commentsAvailableForBounty.map((comment) => (
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
                {isCommenter && (
                  <div className="border-t pt-3">
                    <p className="mb-2 text-xs font-medium text-muted-foreground">
                      Bounty hunt
                    </p>
                    {!comment.marketplace_comment_id ? (
                      (() => {
                        const bounties = bountiesQualifiableByLikes(comment);
                        return bounties.length === 0 ? (
                          <p className="text-xs text-muted-foreground">
                            No active bounties match (min likes). Your comment
                            has {comment.like_count} likes.
                          </p>
                        ) : (
                          <div className="flex flex-wrap gap-2">
                            {bounties.map((b) => (
                              <Button
                                key={b.id}
                                size="sm"
                                variant="secondary"
                                disabled={
                                  claimingForBounty ===
                                  `comment-${comment.comment_id}`
                                }
                                onClick={() =>
                                  handleClaimBounty(b.id, null, comment)
                                }
                              >
                                {claimingForBounty ===
                                `comment-${comment.comment_id}`
                                  ? "Claiming…"
                                  : `Claim $${(b.amount_per_claim_cents / 100).toFixed(2)}`}
                              </Button>
                            ))}
                          </div>
                        );
                      })()
                    ) : loadingEligibleByBounty ? (
                      <p className="text-xs text-muted-foreground">
                        Checking bounties…
                      </p>
                    ) : (() => {
                      const bounties = bountiesForComment(comment);
                      return bounties.length === 0 ? (
                        <p className="text-xs text-muted-foreground">
                          No active bounties match this comment.
                        </p>
                      ) : (
                        <div className="flex flex-wrap gap-2">
                          {bounties.map((b) => (
                            <Button
                              key={b.id}
                              size="sm"
                              variant="secondary"
                              disabled={
                                claimingForBounty ===
                                comment.marketplace_comment_id
                              }
                              onClick={() =>
                                handleClaimBounty(
                                  b.id,
                                  comment.marketplace_comment_id!,
                                  comment
                                )
                              }
                            >
                              {claimingForBounty ===
                              comment.marketplace_comment_id
                                ? "Claiming…"
                                : `Claim $${(b.amount_per_claim_cents / 100).toFixed(2)}`}
                            </Button>
                          ))}
                        </div>
                      );
                    })()}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {isCommenter && bountyMessage && (
        <Card>
          <CardContent className="py-4 text-sm text-muted-foreground">
            {bountyMessage}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
