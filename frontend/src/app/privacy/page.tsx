import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy | AttoAds",
  description:
    "How AttoAds collects, uses, stores, and shares user data, including Google and YouTube data.",
};

const sections = [
  {
    title: "Overview",
    paragraphs: [
      "AttoAds is a micro-sponsorship marketplace that connects advertisers with YouTube commenters. Advertisers create plain-text sponsorship placements, and commenters can authorize AttoAds to add approved sponsorship text to eligible YouTube comments in exchange for payment.",
      "This Privacy Policy explains how AttoAds collects, uses, stores, and shares information when you use the service. It also describes how AttoAds uses Google API Services and YouTube data.",
    ],
  },
  {
    title: "Information We Collect",
    bullets: [
      "Account information, including your email address, selected user role, and account timestamps.",
      "Google account information, including the email and profile information returned during Google OAuth sign-in.",
      "YouTube channel information, including your YouTube channel ID and channel title.",
      "YouTube comment information, including comment IDs, author channel IDs, display names, original comment text, like counts, velocity metrics, publication timestamps, and update timestamps.",
      "OAuth tokens, including access tokens, refresh tokens, and token expiration times needed to provide authorized YouTube features.",
      "Marketplace information, including campaigns, bounties, advertiser text, deals, verification status, transaction records, and performance metrics.",
      "Wallet information, including the blockchain wallet address you choose to link for payments.",
      "Technical information needed to operate and secure the service, such as logs, request metadata, and error information.",
    ],
  },
  {
    title: "Google API Services and YouTube Access",
    paragraphs: [
      "AttoAds uses Google OAuth to authenticate users and request access to YouTube features. The app currently requests the openid, email, profile, and YouTube force-ssl scopes.",
      "The YouTube force-ssl scope allows AttoAds to manage YouTube account data that you authorize, including reading YouTube channel and comment information and updating selected comments. AttoAds uses this access only for user-facing marketplace features you choose to use.",
    ],
  },
  {
    title: "How We Use Information",
    bullets: [
      "Authenticate users and maintain AttoAds accounts.",
      "Identify a user's YouTube channel and match eligible comments to that user.",
      "Discover and rank YouTube comments for marketplace opportunities.",
      "Append approved advertiser text to selected YouTube comments after a marketplace deal is accepted.",
      "Verify that approved sponsorship text appears in the relevant comment.",
      "Track comment performance metrics such as likes and velocity for campaign reporting.",
      "Record campaign, bounty, deal, wallet, and payout information.",
      "Prevent fraud, enforce our terms, debug issues, maintain security, and comply with legal obligations.",
    ],
  },
  {
    title: "How We Store and Protect Information",
    paragraphs: [
      "AttoAds stores account, YouTube channel, OAuth token, comment, campaign, deal, wallet, metric, and transaction records in backend systems used to operate the service.",
      "We use technical and organizational safeguards designed to protect user data, including HTTPS in production, backend access controls, limited internal access, and token handling practices intended to reduce unauthorized access. No method of transmission or storage is perfectly secure.",
    ],
  },
  {
    title: "How We Share Information",
    paragraphs: [
      "AttoAds does not sell Google user data. We do not share Google user data for unrelated advertising, profiling, or resale.",
      "We may share information with service providers that help us operate, host, secure, analyze, or support AttoAds; with blockchain networks when a transaction is submitted; with other users when information is intentionally displayed as part of marketplace functionality; or when required to comply with law, prevent harm, or enforce our terms.",
    ],
  },
  {
    title: "Google API Limited Use Disclosure",
    paragraphs: [
      "AttoAds use and transfer of information received from Google APIs will adhere to the Google API Services User Data Policy, including the Limited Use requirements.",
      "AttoAds uses Google user data only to provide or improve user-facing features of the service. AttoAds does not use Google user data to serve unrelated advertising, does not sell Google user data, and does not transfer Google user data to train generalized artificial intelligence or machine learning models.",
    ],
  },
  {
    title: "User Controls and Data Deletion",
    bullets: [
      "You can revoke AttoAds access to your Google account from your Google Account permissions page.",
      "You can stop using AttoAds or request deletion of your AttoAds account data by contacting us.",
      "You can unlink a wallet from your AttoAds account where the app provides that option.",
      "Some records may be retained when needed for security, fraud prevention, legal compliance, accounting, dispute resolution, or blockchain transaction history.",
    ],
  },
  {
    title: "Retention",
    paragraphs: [
      "We retain user data for as long as needed to provide AttoAds, maintain marketplace records, prevent abuse, comply with legal obligations, resolve disputes, and enforce agreements. When data is no longer needed, we delete, de-identify, or aggregate it where reasonably practical.",
    ],
  },
  {
    title: "Changes to This Policy",
    paragraphs: [
      "We may update this Privacy Policy from time to time. If we materially change how AttoAds uses Google user data, we will update this policy and request any consent required before using Google user data in a new way.",
    ],
  },
  {
    title: "Contact",
    paragraphs: [
      "For privacy questions or data deletion requests, contact AttoAds at support@attoads.app.",
    ],
  },
];

export default function PrivacyPage() {
  return (
    <div className="mx-auto max-w-3xl space-y-10">
      <header className="space-y-4">
        <p className="text-sm font-medium text-muted-foreground">
          Updated: April 27, 2026
        </p>
        <h1 className="text-4xl font-bold tracking-tight">Privacy Policy</h1>
        <p className="text-lg text-muted-foreground">
          This policy describes how AttoAds handles user information,
          including data received from Google APIs and YouTube.
        </p>
      </header>

      <div className="space-y-8">
        {sections.map((section) => (
          <section key={section.title} className="space-y-3">
            <h2 className="text-2xl font-semibold tracking-tight">
              {section.title}
            </h2>
            {section.paragraphs?.map((paragraph) => (
              <p
                key={paragraph}
                className="leading-7 text-muted-foreground"
              >
                {paragraph}
              </p>
            ))}
            {section.bullets && (
              <ul className="list-disc space-y-2 pl-6 text-muted-foreground">
                {section.bullets.map((bullet) => (
                  <li key={bullet} className="leading-7">
                    {bullet}
                  </li>
                ))}
              </ul>
            )}
          </section>
        ))}
      </div>
    </div>
  );
}
