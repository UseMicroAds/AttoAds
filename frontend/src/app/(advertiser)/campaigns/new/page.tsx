"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function NewCampaignPage() {
  const router = useRouter();
  const [adText, setAdText] = useState("");
  const [budget, setBudget] = useState("");
  const [pricePerPlacement, setPricePerPlacement] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const budgetCents = Math.round(parseFloat(budget) * 100);
      const priceCents = Math.round(parseFloat(pricePerPlacement) * 100);

      if (!adText || budgetCents <= 0 || priceCents <= 0) {
        setError("All fields are required and must be positive");
        return;
      }

      await api.createCampaign({
        ad_text: adText,
        budget_cents: budgetCents,
        price_per_placement: priceCents,
      });

      router.push("/campaigns");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create campaign");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-lg">
      <Card>
        <CardHeader>
          <CardTitle>Create Campaign</CardTitle>
          <CardDescription>
            Set up a new advertising campaign. Your ad text will be appended to
            viral YouTube comments.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Ad Text</label>
              <Textarea
                placeholder='e.g., "Find X at Whole Foods" or "Search for Y App"'
                value={adText}
                onChange={(e) => setAdText(e.target.value)}
                maxLength={200}
              />
              <p className="text-xs text-muted-foreground">
                {adText.length}/200 characters. Plain text only — no links.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Total Budget (USD)</label>
                <Input
                  type="number"
                  step="0.01"
                  min="1"
                  placeholder="50.00"
                  value={budget}
                  onChange={(e) => setBudget(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  Price per Placement (USD)
                </label>
                <Input
                  type="number"
                  step="0.01"
                  placeholder="5.00"
                  value={pricePerPlacement}
                  onChange={(e) => setPricePerPlacement(e.target.value)}
                />
              </div>
            </div>

            {budget && pricePerPlacement && (
              <p className="text-sm text-muted-foreground">
                ≈ {Math.floor(parseFloat(budget) / parseFloat(pricePerPlacement))}{" "}
                placements
              </p>
            )}

            {error && (
              <p className="text-sm text-destructive">{error}</p>
            )}

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Creating..." : "Create Campaign"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
