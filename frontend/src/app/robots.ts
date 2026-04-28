import type { MetadataRoute } from "next";
import { siteConfig } from "@/lib/seo";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: ["/", "/marketplace", "/privacy", "/terms"],
      disallow: [
        "/bounties",
        "/callback",
        "/campaigns",
        "/dashboard",
        "/login",
        "/performance",
        "/test",
      ],
    },
    sitemap: `${siteConfig.url}/sitemap.xml`,
  };
}

