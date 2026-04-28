import type { Metadata } from "next";
import { siteConfig } from "@/lib/seo";
import { MarketplaceClient } from "./marketplace-client";

export const metadata: Metadata = {
  title: "Marketplace",
  description:
    "Browse viral YouTube comments available for advertiser sponsorship through MicroAds.",
  alternates: {
    canonical: "/marketplace",
  },
  openGraph: {
    title: `Marketplace | ${siteConfig.name}`,
    description:
      "Browse viral YouTube comments available for advertiser sponsorship through MicroAds.",
    url: "/marketplace",
  },
};

export default function MarketplacePage() {
  return <MarketplaceClient />;
}
