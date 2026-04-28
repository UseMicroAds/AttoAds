import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms of Service | AttoAds",
  description:
    "The terms that govern advertiser and commenter use of AttoAds.",
};

const sections = [
  {
    title: "Acceptance of Terms",
    paragraphs: [
      "These Terms of Service govern your access to and use of AttoAds. By using AttoAds, creating an account, connecting a Google account, linking a wallet, creating a campaign, accepting a bounty, or receiving a payout, you agree to these terms.",
      "If you use AttoAds on behalf of a company or other organization, you represent that you have authority to bind that organization to these terms.",
    ],
  },
  {
    title: "The AttoAds Service",
    paragraphs: [
      "AttoAds is a marketplace that helps advertisers sponsor plain-text placements in eligible YouTube comments. Advertisers submit sponsorship text and fund campaigns or bounties. Commenters can authorize AttoAds to add approved sponsorship text to selected YouTube comments and may receive payment when the placement is verified.",
      "AttoAds does not own or operate YouTube, Google, blockchain networks, wallets, or third-party infrastructure used with the service.",
    ],
  },
  {
    title: "Accounts and Eligibility",
    bullets: [
      "You must provide accurate account information and keep your account secure.",
      "You are responsible for activity that occurs through your AttoAds account.",
      "You may only connect Google, YouTube, and wallet accounts that you own or are authorized to use.",
      "You must comply with all laws, platform rules, and third-party terms that apply to your use of AttoAds.",
    ],
  },
  {
    title: "Google and YouTube Authorization",
    paragraphs: [
      "AttoAds uses Google OAuth to authenticate users and access YouTube features that support the marketplace. By connecting your Google account, you authorize AttoAds to access the Google and YouTube data described in the Privacy Policy and to perform actions you authorize through the service.",
      "For commenters, this may include identifying your YouTube channel, reading relevant YouTube comment information, appending approved advertiser text to selected comments, and verifying that the placement appears in the comment.",
    ],
  },
  {
    title: "Commenter Responsibilities",
    bullets: [
      "You may only offer comments and YouTube channels that you control.",
      "You authorize AttoAds to append approved sponsorship text to selected comments when you accept or become matched with a marketplace placement.",
      "You are responsible for ensuring your use of AttoAds complies with YouTube rules, Google policies, disclosure obligations, advertising laws, and any other applicable requirements.",
      "You must not use AttoAds to manipulate engagement, impersonate others, mislead viewers, or violate another person's rights.",
      "You acknowledge that YouTube may remove, demote, restrict, or otherwise affect comments or accounts independent of AttoAds.",
    ],
  },
  {
    title: "Advertiser Responsibilities",
    bullets: [
      "You are responsible for the accuracy, legality, and appropriateness of sponsorship text and campaign materials you submit.",
      "You must not submit deceptive, illegal, harmful, infringing, discriminatory, or abusive content.",
      "You must not promote prohibited products or services or content that violates YouTube, Google, or applicable advertising rules.",
      "You are responsible for any disclosures, approvals, substantiation, and compliance obligations that apply to your advertising.",
      "AttoAds may reject, remove, pause, or cancel campaigns or bounties at its discretion.",
    ],
  },
  {
    title: "Marketplace Payments",
    paragraphs: [
      "Advertisers may fund campaigns or bounties, and commenters may receive payouts when AttoAds verifies that the agreed sponsorship text appears in the selected comment. Exact pricing, payout amounts, and release conditions are shown in the product flow or campaign terms at the time of use.",
      "Some payments may use USDC, smart contracts, wallets, or blockchain networks. Blockchain transactions can be irreversible, delayed, failed, or affected by network fees, wallet errors, contract errors, or third-party outages. AttoAds is not responsible for losses caused by incorrect wallet addresses, user wallet compromise, blockchain network behavior, or third-party wallet providers.",
    ],
  },
  {
    title: "Content and Permissions",
    paragraphs: [
      "You retain ownership of content you submit or control. You grant AttoAds a limited, worldwide, non-exclusive license to host, process, display, analyze, and transmit content as needed to operate the service.",
      "If you are a commenter, you also grant AttoAds permission to modify selected YouTube comments by appending approved sponsorship text when authorized through the service. If you are an advertiser, you grant AttoAds permission to use your submitted sponsorship text to fulfill campaign placements.",
    ],
  },
  {
    title: "Platform Dependencies",
    paragraphs: [
      "AttoAds depends on Google, YouTube, Google APIs, OAuth access, blockchain networks, wallets, hosting providers, and other third-party services. These services may change, suspend, restrict, or terminate functionality at any time.",
      "AttoAds does not guarantee that YouTube comments will remain visible, ranked, monetizable, editable, or eligible for sponsorship. Google may reject OAuth scopes, require additional verification, impose security assessments, or revoke API access.",
    ],
  },
  {
    title: "Prohibited Conduct",
    bullets: [
      "Do not violate laws, platform terms, or the rights of others.",
      "Do not submit spam, malware, deceptive claims, illegal offers, or harmful content.",
      "Do not attempt to bypass security, access accounts or data without authorization, scrape the service, or interfere with AttoAds systems.",
      "Do not use AttoAds to manipulate platform integrity, misrepresent sponsorships, or evade required disclosures.",
      "Do not use AttoAds in a way that could harm users, advertisers, commenters, Google, YouTube, or AttoAds.",
    ],
  },
  {
    title: "Suspension and Termination",
    paragraphs: [
      "AttoAds may suspend or terminate access, reject campaigns, stop placements, withhold or reverse unpaid marketplace activity, or remove content if we believe a user has violated these terms, created risk, caused harm, or used the service improperly.",
      "You may stop using AttoAds at any time. Some records may be retained as described in the Privacy Policy.",
    ],
  },
  {
    title: "Disclaimers and Limits of Liability",
    paragraphs: [
      "AttoAds is provided on an as-is and as-available basis. To the maximum extent permitted by law, AttoAds disclaims warranties of merchantability, fitness for a particular purpose, non-infringement, availability, accuracy, and uninterrupted operation.",
      "To the maximum extent permitted by law, AttoAds will not be liable for indirect, incidental, special, consequential, exemplary, or punitive damages, lost profits, lost revenue, lost data, platform actions, blockchain failures, or third-party service issues.",
    ],
  },
  {
    title: "Indemnity",
    paragraphs: [
      "You agree to defend, indemnify, and hold harmless AttoAds from claims, damages, losses, liabilities, costs, and expenses arising from your use of the service, your content, your campaigns, your comments, your wallet activity, your violation of these terms, or your violation of law or third-party rights.",
    ],
  },
  {
    title: "Changes to These Terms",
    paragraphs: [
      "AttoAds may update these terms from time to time. Updated terms will be posted on this page with a new effective date. Your continued use of AttoAds after changes become effective means you accept the updated terms.",
    ],
  },
  {
    title: "Contact",
    paragraphs: [
      "For questions about these terms, contact AttoAds at support@attoads.com.",
    ],
  },
];

export default function TermsPage() {
  return (
    <div className="mx-auto max-w-3xl space-y-10">
      <header className="space-y-4">
        <p className="text-sm font-medium text-muted-foreground">
          Updated: April 27, 2026
        </p>
        <h1 className="text-4xl font-bold tracking-tight">Terms of Service</h1>
        <p className="text-lg text-muted-foreground">
          These terms explain the rules for using AttoAds as an advertiser,
          commenter, or marketplace participant.
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
