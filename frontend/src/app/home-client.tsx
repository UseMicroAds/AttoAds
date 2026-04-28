"use client";

import Image from "next/image";
import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";

export function HomeClient() {
  return (
    <div className="flex flex-col items-center justify-center gap-8 pt-24 text-center">
      <Image
        src="/attoads-logo-with-text.svg"
        alt="AttoAds logo"
        width={160}
        height={160}
        className="h-32 w-32 shadow-sm sm:h-40 sm:w-40"
        priority
      />
      <h1 className="max-w-3xl text-5xl font-bold tracking-tight sm:text-6xl">
        Turn viral comments into
        <span className="text-primary"> revenue</span>
      </h1>
      <p className="max-w-xl text-lg text-muted-foreground">
        AttoAds connects brands with YouTube&apos;s top commenters. Advertisers
        get high-visibility placements. Commenters get paid instantly in USDC.
      </p>
      <div className="flex gap-4">
        <Link href="/login" className={buttonVariants({ size: "lg" })}>
          Get Started
        </Link>
        <Link
          href="/marketplace"
          className={buttonVariants({ size: "lg", variant: "outline" })}
        >
          Browse Marketplace
        </Link>
      </div>

      <div className="mt-16 grid w-full max-w-4xl gap-8 sm:grid-cols-3">
        <FeatureCard
          title="Discover"
          description="Our engine monitors trending YouTube videos and identifies comments going viral in real-time."
        />
        <FeatureCard
          title="Place"
          description="Advertisers create campaigns with plain-text shoutouts. We match them to high-velocity comments."
        />
        <FeatureCard
          title="Earn"
          description="Commenters receive instant USDC payments the moment their comment edit is verified."
        />
      </div>
    </div>
  );
}

function FeatureCard({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="rounded-xl border bg-card p-6 text-left">
      <h3 className="mb-2 text-lg font-semibold">{title}</h3>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}

