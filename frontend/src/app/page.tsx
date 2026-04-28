import type { Metadata } from "next";
import { siteConfig } from "@/lib/seo";
import { HomeClient } from "./home-client";

export const metadata: Metadata = {
  title: siteConfig.title,
  description: siteConfig.description,
  alternates: {
    canonical: "/",
  },
};

export default function HomePage() {
  return <HomeClient />;
}
