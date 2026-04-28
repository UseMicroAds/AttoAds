const html = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Privacy Policy | AttoAds</title>
    <meta
      name="description"
      content="How AttoAds collects, uses, stores, and shares user data, including Google and YouTube data."
    />
    <style>
      :root {
        color-scheme: light;
        font-family:
          Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont,
          "Segoe UI", sans-serif;
        color: #171717;
        background: #ffffff;
      }

      body {
        margin: 0;
        padding: 48px 24px;
      }

      main {
        max-width: 760px;
        margin: 0 auto;
      }

      nav {
        margin-bottom: 32px;
      }

      a {
        color: #374151;
      }

      h1 {
        margin: 0 0 8px;
        font-size: 40px;
        line-height: 1.1;
      }

      h2 {
        margin: 32px 0 12px;
        font-size: 24px;
        line-height: 1.25;
      }

      p,
      li {
        color: #4b5563;
        font-size: 16px;
        line-height: 1.7;
      }

      ul {
        padding-left: 24px;
      }

      .updated {
        margin: 0 0 24px;
        color: #6b7280;
        font-size: 14px;
        font-weight: 600;
      }
    </style>
  </head>
  <body>
    <main>
      <nav>
        <a href="/">AttoAds home</a>
      </nav>

      <article>
        <header>
          <p class="updated">Updated: April 27, 2026</p>
          <h1>Privacy Policy</h1>
          <p>
            This policy describes how AttoAds handles user information,
            including data received from Google APIs and YouTube.
          </p>
        </header>

        <section>
          <h2>Overview</h2>
          <p>
            AttoAds is a micro-sponsorship marketplace that connects
            advertisers with YouTube commenters. Advertisers create plain-text
            sponsorship placements, and commenters can authorize AttoAds to add
            approved sponsorship text to eligible YouTube comments in exchange
            for payment.
          </p>
          <p>
            This Privacy Policy explains how AttoAds collects, uses, stores,
            and shares information when you use the service. It also describes
            how AttoAds uses Google API Services and YouTube data.
          </p>
        </section>

        <section>
          <h2>Information We Collect</h2>
          <ul>
            <li>
              Account information, including your email address, selected user
              role, and account timestamps.
            </li>
            <li>
              Google account information, including the email and profile
              information returned during Google OAuth sign-in.
            </li>
            <li>
              YouTube channel information, including your YouTube channel ID
              and channel title.
            </li>
            <li>
              YouTube comment information, including comment IDs, author
              channel IDs, display names, original comment text, like counts,
              velocity metrics, publication timestamps, and update timestamps.
            </li>
            <li>
              OAuth tokens, including access tokens, refresh tokens, and token
              expiration times needed to provide authorized YouTube features.
            </li>
            <li>
              Marketplace information, including campaigns, bounties,
              advertiser text, deals, verification status, transaction records,
              and performance metrics.
            </li>
            <li>
              Wallet information, including the blockchain wallet address you
              choose to link for payments.
            </li>
            <li>
              Technical information needed to operate and secure the service,
              such as logs, request metadata, and error information.
            </li>
          </ul>
        </section>

        <section>
          <h2>Google API Services and YouTube Access</h2>
          <p>
            AttoAds uses Google OAuth to authenticate users and request access
            to YouTube features. The app currently requests the openid, email,
            profile, and YouTube force-ssl scopes.
          </p>
          <p>
            The YouTube force-ssl scope allows AttoAds to manage YouTube
            account data that you authorize, including reading YouTube channel
            and comment information and updating selected comments. AttoAds uses
            this access only for user-facing marketplace features you choose to
            use.
          </p>
        </section>

        <section>
          <h2>How We Use Information</h2>
          <ul>
            <li>Authenticate users and maintain AttoAds accounts.</li>
            <li>
              Identify a user's YouTube channel and match eligible comments to
              that user.
            </li>
            <li>
              Discover and rank YouTube comments for marketplace opportunities.
            </li>
            <li>
              Append approved advertiser text to selected YouTube comments after
              a marketplace deal is accepted.
            </li>
            <li>
              Verify that approved sponsorship text appears in the relevant
              comment.
            </li>
            <li>
              Track comment performance metrics such as likes and velocity for
              campaign reporting.
            </li>
            <li>
              Record campaign, bounty, deal, wallet, and payout information.
            </li>
            <li>
              Prevent fraud, enforce our terms, debug issues, maintain
              security, and comply with legal obligations.
            </li>
          </ul>
        </section>

        <section>
          <h2>How We Store and Protect Information</h2>
          <p>
            AttoAds stores account, YouTube channel, OAuth token, comment,
            campaign, deal, wallet, metric, and transaction records in backend
            systems used to operate the service.
          </p>
          <p>
            We use technical and organizational safeguards designed to protect
            user data, including HTTPS in production, backend access controls,
            limited internal access, and token handling practices intended to
            reduce unauthorized access. No method of transmission or storage is
            perfectly secure.
          </p>
        </section>

        <section>
          <h2>How We Share Information</h2>
          <p>
            AttoAds does not sell Google user data. We do not share Google user
            data for unrelated advertising, profiling, or resale.
          </p>
          <p>
            We may share information with service providers that help us
            operate, host, secure, analyze, or support AttoAds; with blockchain
            networks when a transaction is submitted; with other users when
            information is intentionally displayed as part of marketplace
            functionality; or when required to comply with law, prevent harm, or
            enforce our terms.
          </p>
        </section>

        <section>
          <h2>Google API Limited Use Disclosure</h2>
          <p>
            AttoAds use and transfer of information received from Google APIs
            will adhere to the Google API Services User Data Policy, including
            the Limited Use requirements.
          </p>
          <p>
            AttoAds uses Google user data only to provide or improve
            user-facing features of the service. AttoAds does not use Google
            user data to serve unrelated advertising, does not sell Google user
            data, and does not transfer Google user data to train generalized
            artificial intelligence or machine learning models.
          </p>
        </section>

        <section>
          <h2>User Controls and Data Deletion</h2>
          <ul>
            <li>
              You can revoke AttoAds access to your Google account from your
              Google Account permissions page.
            </li>
            <li>
              You can stop using AttoAds or request deletion of your AttoAds
              account data by contacting us.
            </li>
            <li>
              You can unlink a wallet from your AttoAds account where the app
              provides that option.
            </li>
            <li>
              Some records may be retained when needed for security, fraud
              prevention, legal compliance, accounting, dispute resolution, or
              blockchain transaction history.
            </li>
          </ul>
        </section>

        <section>
          <h2>Retention</h2>
          <p>
            We retain user data for as long as needed to provide AttoAds,
            maintain marketplace records, prevent abuse, comply with legal
            obligations, resolve disputes, and enforce agreements. When data is
            no longer needed, we delete, de-identify, or aggregate it where
            reasonably practical.
          </p>
        </section>

        <section>
          <h2>Changes to This Policy</h2>
          <p>
            We may update this Privacy Policy from time to time. If we
            materially change how AttoAds uses Google user data, we will update
            this policy and request any consent required before using Google
            user data in a new way.
          </p>
        </section>

        <section>
          <h2>Contact</h2>
          <p>
            For privacy questions or data deletion requests, contact AttoAds at
            <a href="mailto:support@attoads.app">support@attoads.app</a>.
          </p>
        </section>
      </article>
    </main>
  </body>
</html>`;

export function GET() {
  return new Response(html, {
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": "public, max-age=3600",
    },
  });
}
