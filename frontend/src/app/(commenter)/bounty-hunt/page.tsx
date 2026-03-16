"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, type Bounty, type ViralComment } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function BountyHuntPage() {
  const { user } = useAuth();
  const [bounties, setBounties] = useState<Bounty[]>([]);
  const [selectedBountyId, setSelectedBountyId] = useState<string | null>(null);
  const [eligibleComments, setEligibleComments] = useState<ViralComment[]>([]);
  const [loadingBounties, setLoadingBounties] = useState(true);
  const [loadingComments, setLoadingComments] = useState(false);
  const [claimingFor, setClaimingFor] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (user) {
      api
        .listActiveBounties()
        .then(setBounties)
        .catch(() => {})
        .finally(() => setLoadingBounties(false));
    }
  }, [user]);

  useEffect(() => {
    if (!selectedBountyId) {
      setEligibleComments([]);
      return;
    }
    setLoadingComments(true);
    api
      .listEligibleCommentsForBounty(selectedBountyId)
      .then(setEligibleComments)
      .catch(() => setEligibleComments([]))
      .finally(() => setLoadingComments(false));
  }, [selectedBountyId]);

  const handleClaim = async (commentId: string) => {
    if (!selectedBountyId) return;
    setMessage(null);
    setClaimingFor(commentId);
    try {
      await api.claimBounty(selectedBountyId, commentId);
      setMessage("Claim submitted. Add the ad text to your comment; you'll be paid after verification.");
      setEligibleComments((prev) => prev.filter((c) => c.id !== commentId));
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Failed to claim");
    } finally {
      setClaimingFor(null);
    }
  };

  if (!user) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">Please sign in to hunt bounties.</p>
      </div>
    );
  }

  const selectedBounty = bounties.find((b) => b.id === selectedBountyId);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Bounty Hunt</h1>
        <p className="text-muted-foreground">
          Find bounties that match your comments. If your comment meets the rules, claim and add the ad text.
        </p>
      </div>

      {loadingBounties ? (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            Loading bounties...
          </CardContent>
        </Card>
      ) : bounties.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            No active bounties. Check back later.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Active Bounties</CardTitle>
              <CardDescription>Select one to see eligible comments</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              {bounties.map((bounty) => (
                <button
                  key={bounty.id}
                  type="button"
                  onClick={() => setSelectedBountyId(bounty.id)}
                  className={`w-full rounded-lg border p-3 text-left transition-colors ${
                    selectedBountyId === bounty.id
                      ? "border-primary bg-muted/50"
                      : "hover:bg-muted/30"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="truncate font-medium">
                      &ldquo;{bounty.ad_text.slice(0, 30)}…&rdquo;
                    </span>
                    <Badge variant="secondary">
                      ${(bounty.amount_per_claim_cents / 100).toFixed(2)}/claim
                    </Badge>
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Min {bounty.min_likes} likes
                    {bounty.video_category && ` · ${bounty.video_category}`}
                  </p>
                </button>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Eligible comments</CardTitle>
              <CardDescription>
                {selectedBounty
                  ? `Your comments that match: min ${selectedBounty.min_likes} likes${selectedBounty.video_category ? `, category ${selectedBounty.video_category}` : ""}`
                  : "Select a bounty"}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {!selectedBountyId ? (
                <p className="text-sm text-muted-foreground">
                  Select a bounty from the list.
                </p>
              ) : loadingComments ? (
                <p className="text-sm text-muted-foreground">Loading...</p>
              ) : eligibleComments.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No eligible comments from your channel for this bounty. Only your comments that meet the rules are shown.
                </p>
              ) : (
                <ul className="space-y-3">
                  {eligibleComments.map((c) => (
                    <li
                      key={c.id}
                      className="flex items-center justify-between rounded border p-3"
                    >
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm">{c.original_text}</p>
                        <p className="text-xs text-muted-foreground">
                          {c.like_count} likes
                        </p>
                      </div>
                      <Button
                        size="sm"
                        onClick={() => handleClaim(c.id)}
                        disabled={claimingFor === c.id}
                      >
                        {claimingFor === c.id ? "Claiming…" : "Claim"}
                      </Button>
                    </li>
                  ))}
                </ul>
              )}
              {message && (
                <p className="mt-4 text-sm text-muted-foreground">{message}</p>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
