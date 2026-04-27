"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import { api, type Bounty } from "@/lib/api";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

const statusColors: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  draft: "outline",
  funded: "secondary",
  active: "default",
  completed: "secondary",
};

export default function BountiesPage() {
  const { user } = useAuth();
  const [bounties, setBounties] = useState<Bounty[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (user) {
      api
        .listBounties()
        .then(setBounties)
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [user]);

  if (!user) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">
          Please sign in as an advertiser.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Bounties</h1>
          <p className="text-muted-foreground">
            Set up bounties with constraints; commenters hunt when their comments qualify.
          </p>
        </div>
        <Link href="/bounties/new" className={buttonVariants()}>
          New Bounty
        </Link>
      </div>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-4 w-3/4 rounded bg-muted" />
              </CardHeader>
              <CardContent>
                <div className="h-12 rounded bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : bounties.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-muted-foreground">
              No bounties yet. Create a bounty with rules (e.g. video category, min likes) and fund it.
            </p>
            <Link
              href="/bounties/new"
              className={buttonVariants({ className: "mt-4" })}
            >
              Create Bounty
            </Link>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {bounties.map((bounty) => (
            <Card key={bounty.id} className="transition-shadow hover:shadow-md">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">
                    &ldquo;{bounty.ad_text.slice(0, 40)}
                    {bounty.ad_text.length > 40 ? "…" : ""}&rdquo;
                  </CardTitle>
                  <Badge variant={statusColors[bounty.status] || "secondary"}>
                    {bounty.status}
                  </Badge>
                </div>
                <CardDescription>
                  Budget: ${(bounty.budget_cents / 100).toFixed(2)} &middot; Per claim: $
                  {(bounty.amount_per_claim_cents / 100).toFixed(2)}
                  {bounty.min_likes > 0 && ` · Min ${bounty.min_likes} likes`}
                  {bounty.video_category && ` · ${bounty.video_category}`}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex items-center justify-between">
                <span className="text-xs text-muted-foreground">
                  Created {new Date(bounty.created_at).toLocaleDateString()}
                </span>
                <Link
                  href={`/bounties/${bounty.id}`}
                  className={buttonVariants({ variant: "outline", size: "sm" })}
                >
                  View
                </Link>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
