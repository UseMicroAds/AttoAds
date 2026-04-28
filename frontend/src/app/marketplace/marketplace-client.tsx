"use client";

import { useEffect, useState } from "react";
import { api, type ViralComment } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function MarketplaceClient() {
  const [comments, setComments] = useState<ViralComment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .listComments(20, 0)
      .then(setComments)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Marketplace</h1>
        <p className="text-muted-foreground">
          Browse viral comments available for sponsorship
        </p>
      </div>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i} className="animate-pulse">
              <CardHeader>
                <div className="h-4 w-3/4 rounded bg-muted" />
                <div className="h-3 w-1/2 rounded bg-muted" />
              </CardHeader>
              <CardContent>
                <div className="h-16 rounded bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : comments.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-muted-foreground">
              No viral comments available yet. The discovery engine is scanning
              for trending content.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {comments.map((comment) => (
            <CommentCard key={comment.id} comment={comment} />
          ))}
        </div>
      )}
    </div>
  );
}

function CommentCard({ comment }: { comment: ViralComment }) {
  const videoUrl = `https://www.youtube.com/watch?v=${comment.video_id}`;

  return (
    <Card className="transition-shadow hover:shadow-md">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium">
            {comment.author_display_name}
          </CardTitle>
          <Badge variant="secondary">
            {Math.round(comment.velocity)} likes/min
          </Badge>
        </div>
        <CardDescription className="text-xs">
          {comment.like_count.toLocaleString()} likes
          {comment.video_category && (
            <>
              {" "}
              ·{" "}
              <span className="text-muted-foreground">
                {comment.video_category}
              </span>
            </>
          )}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {comment.video_title && (
          <p
            className="mb-1 line-clamp-1 text-xs text-muted-foreground"
            title={comment.video_title}
          >
            {comment.video_title}
          </p>
        )}
        <p className="line-clamp-3 text-sm">{comment.original_text}</p>
        <div className="mt-3 flex items-center justify-between gap-2">
          <p className="text-xs text-muted-foreground">
            Spotted {new Date(comment.first_seen).toLocaleDateString()}
          </p>
          <Button
            size="sm"
            variant="outline"
            onClick={() => window.open(videoUrl, "_blank", "noopener,noreferrer")}
          >
            Original video
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

