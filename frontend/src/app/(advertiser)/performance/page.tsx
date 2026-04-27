"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { useAuth } from "@/lib/auth-context";
import {
  api,
  type DealPerformanceRow,
  type DealPerformanceResponse,
} from "@/lib/api";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function PerformancePage() {
  const { user } = useAuth();
  const [deals, setDeals] = useState<DealPerformanceRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDealId, setSelectedDealId] = useState<string | null>(null);
  const [performance, setPerformance] = useState<DealPerformanceResponse | null>(
    null
  );
  const [loadingPerformance, setLoadingPerformance] = useState(false);

  useEffect(() => {
    if (user?.role === "advertiser") {
      api
        .listDealsPerformance()
        .then(setDeals)
        .catch(() => setDeals([]))
        .finally(() => setLoading(false));
    }
  }, [user]);

  useEffect(() => {
    if (!selectedDealId) {
      setPerformance(null);
      return;
    }
    let isStale = false;
    const currentDealId = selectedDealId;

    setLoadingPerformance(true);
    api
      .getDealPerformance(currentDealId)
      .then((data) => {
        if (!isStale && currentDealId === selectedDealId) {
          setPerformance(data);
        }
      })
      .catch(() => {
        if (!isStale && currentDealId === selectedDealId) {
          setPerformance(null);
        }
      })
      .finally(() => {
        if (!isStale && currentDealId === selectedDealId) {
          setLoadingPerformance(false);
        }
      });

    return () => {
      isStale = true;
    };
  }, [selectedDealId]);

  if (user?.role !== "advertiser") {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">
          Sign in as an advertiser to view deal performance.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/campaigns" className={buttonVariants({ variant: "ghost" })}>
          ← Campaigns
        </Link>
      </div>
      <div>
        <h1 className="text-3xl font-bold">Deal Performance</h1>
        <p className="text-muted-foreground">
          See each deal and how the comment is performing (likes, velocity over time). Open the comment on YouTube to view the edited ad.
        </p>
      </div>

      {loading ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            Loading deals…
          </CardContent>
        </Card>
      ) : deals.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            No deals yet. Create campaigns or bounties and complete deals to see performance here.
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {deals.map((deal) => (
            <Card key={deal.id} className="overflow-hidden">
              <CardHeader className="pb-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <CardTitle className="text-base">
                    Deal {deal.id.slice(0, 8)}…
                  </CardTitle>
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary">{deal.status}</Badge>
                    {deal.edited_at && (
                      <span className="text-xs text-muted-foreground">
                        Edited {new Date(deal.edited_at).toLocaleString()}
                      </span>
                    )}
                  </div>
                </div>
                <CardDescription className="line-clamp-2">
                  {deal.original_text}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex flex-wrap items-center gap-3">
                  <a
                    href={deal.comment_link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={buttonVariants({ variant: "outline", size: "sm" })}
                  >
                    View comment on YouTube
                  </a>
                  <span className="text-sm text-muted-foreground">
                    {deal.like_count.toLocaleString()} likes · {deal.velocity.toFixed(1)} likes/min
                  </span>
                  <button
                    type="button"
                    className="text-sm text-primary underline hover:no-underline"
                    onClick={() =>
                      setSelectedDealId(selectedDealId === deal.id ? null : deal.id)
                    }
                  >
                    {selectedDealId === deal.id ? "Hide graph" : "Show velocity graph"}
                  </button>
                </div>

                {selectedDealId === deal.id && (
                  <div className="rounded-lg border bg-muted/30 p-4">
                    {loadingPerformance ? (
                      <p className="py-8 text-center text-sm text-muted-foreground">
                        Loading metrics…
                      </p>
                    ) : performance && performance.metrics.length > 0 ? (
                      <div className="h-[280px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                          <LineChart
                            data={performance.metrics.map((m) => ({
                              ...m,
                              time: new Date(m.captured_at).toLocaleString(undefined, {
                                month: "short",
                                day: "numeric",
                                hour: "2-digit",
                                minute: "2-digit",
                              }),
                            }))}
                            margin={{ top: 5, right: 20, left: 0, bottom: 5 }}
                          >
                            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                            <XAxis
                              dataKey="time"
                              tick={{ fontSize: 11 }}
                              className="text-muted-foreground"
                            />
                            <YAxis
                              tick={{ fontSize: 11 }}
                              className="text-muted-foreground"
                              label={{
                                value: "Likes/min",
                                angle: -90,
                                position: "insideLeft",
                                style: { fontSize: 11 },
                              }}
                            />
                            <Tooltip
                              formatter={(value: number) => [value.toFixed(2), "Velocity"]}
                              labelFormatter={(label) => `Time: ${label}`}
                            />
                            <Line
                              type="monotone"
                              dataKey="velocity"
                              stroke="hsl(var(--chart-1))"
                              strokeWidth={2}
                              dot={{ r: 3 }}
                              name="Likes/min"
                            />
                          </LineChart>
                        </ResponsiveContainer>
                        <p className="mt-2 text-center text-xs text-muted-foreground">
                          Likes per minute since the comment was edited, based on an initial verifier capture and periodic performance worker updates.
                        </p>
                      </div>
                    ) : (
                      <p className="py-6 text-center text-sm text-muted-foreground">
                        No metrics yet. Data is recorded shortly after the comment is verified or edited, and then periodically by the performance worker.
                      </p>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
