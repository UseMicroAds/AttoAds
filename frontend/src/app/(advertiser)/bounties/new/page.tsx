"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function NewBountyPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [adText, setAdText] = useState("");
  const [budgetCents, setBudgetCents] = useState("");
  const [amountPerClaimCents, setAmountPerClaimCents] = useState("");
  const [videoCategory, setVideoCategory] = useState("");
  const [minLikes, setMinLikes] = useState("0");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    const budget = Math.round(parseFloat(budgetCents) * 100);
    const perClaim = Math.round(parseFloat(amountPerClaimCents) * 100);
    if (!adText.trim() || budget <= 0 || perClaim <= 0) {
      setError("Ad text, budget, and amount per claim are required.");
      return;
    }
    if (perClaim > budget) {
      setError("Amount per claim cannot exceed total budget.");
      return;
    }
    setLoading(true);
    try {
      const bounty = await api.createBounty({
        ad_text: adText.trim(),
        budget_cents: budget,
        amount_per_claim_cents: perClaim,
        video_category: videoCategory.trim() || undefined,
        min_likes: parseInt(minLikes, 10) || 0,
      });
      router.push(`/bounties/${bounty.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create bounty");
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">Please sign in as an advertiser.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-xl space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/bounties" className={buttonVariants({ variant: "ghost" })}>
          ← Bounties
        </Link>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>New Bounty</CardTitle>
          <CardDescription>
            Set rules (e.g. video category, min likes). Commenters whose comments match can claim.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="ad_text" className="mb-1 block text-sm font-medium">
                Ad text (appended to comment)
              </label>
              <Input
                id="ad_text"
                value={adText}
                onChange={(e) => setAdText(e.target.value)}
                placeholder="Check out our product..."
                className="mt-1"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="budget" className="mb-1 block text-sm font-medium">
                  Total budget (USD)
                </label>
                <Input
                  id="budget"
                  type="number"
                  step="0.01"
                  min="0"
                  value={budgetCents}
                  onChange={(e) => setBudgetCents(e.target.value)}
                  placeholder="10.00"
                  className="mt-1"
                />
              </div>
              <div>
                <label htmlFor="per_claim" className="mb-1 block text-sm font-medium">
                  Amount per claim (USD)
                </label>
                <Input
                  id="per_claim"
                  type="number"
                  step="0.01"
                  min="0"
                  value={amountPerClaimCents}
                  onChange={(e) => setAmountPerClaimCents(e.target.value)}
                  placeholder="1.00"
                  className="mt-1"
                />
              </div>
            </div>
            <div>
              <label htmlFor="video_category" className="mb-1 block text-sm font-medium">
                Video category (optional)
              </label>
              <Input
                id="video_category"
                value={videoCategory}
                onChange={(e) => setVideoCategory(e.target.value)}
                placeholder="e.g. AI, Tech"
                className="mt-1"
              />
            </div>
            <div>
              <label htmlFor="min_likes" className="mb-1 block text-sm font-medium">
                Min likes on comment
              </label>
              <Input
                id="min_likes"
                type="number"
                min="0"
                value={minLikes}
                onChange={(e) => setMinLikes(e.target.value)}
                className="mt-1"
              />
            </div>
            {error && (
              <p className="text-sm text-destructive">{error}</p>
            )}
            <Button type="submit" disabled={loading}>
              {loading ? "Creating…" : "Create Bounty"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
