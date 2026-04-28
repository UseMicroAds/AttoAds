"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const GOOGLE_CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID || "";

function buildGoogleOAuthURL(role: string) {
  const params = new URLSearchParams({
    client_id: GOOGLE_CLIENT_ID,
    redirect_uri: `${window.location.origin}/callback`,
    response_type: "code",
    scope:
      "openid email profile https://www.googleapis.com/auth/youtube.force-ssl",
    access_type: "offline",
    prompt: "consent",
    state: role,
  });
  return `https://accounts.google.com/o/oauth2/v2/auth?${params}`;
}

export default function LoginPage() {
  const [role, setRole] = useState<"commenter" | "advertiser">("commenter");

  const handleLogin = () => {
    window.location.href = buildGoogleOAuthURL(role);
  };

  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Sign in to AttoAds</CardTitle>
          <CardDescription>
            Connect your YouTube account to get started
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <Tabs
            value={role}
            onValueChange={(v) => setRole(v as "commenter" | "advertiser")}
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="commenter">Commenter</TabsTrigger>
              <TabsTrigger value="advertiser">Advertiser</TabsTrigger>
            </TabsList>
            <TabsContent value="commenter" className="mt-4">
              <p className="text-sm text-muted-foreground">
                Monetize your top YouTube comments. Get paid in USDC when brands
                sponsor your viral posts.
              </p>
            </TabsContent>
            <TabsContent value="advertiser" className="mt-4">
              <p className="text-sm text-muted-foreground">
                Place your brand alongside the most-viewed YouTube comments.
                Pure billboard awareness at micro-transaction prices.
              </p>
            </TabsContent>
          </Tabs>

          <Button onClick={handleLogin} className="w-full" size="lg">
            Continue with Google
          </Button>

          <p className="text-center text-xs text-muted-foreground">
            By signing in, you grant AttoAds permission to manage YouTube
            comments on your behalf.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
