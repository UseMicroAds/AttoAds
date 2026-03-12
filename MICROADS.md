# Project Overview: Viral Comment Monetization Platform

## 1. The Core Concept & Value Proposition
* **The Problem:** Small to medium businesses (SMBs) need cheap, high-visibility brand awareness but are priced out of traditional ad networks. Meanwhile, YouTube users who land top viral comments get social clout but zero financial reward.
* **The Solution:** A micro-sponsorship marketplace connecting brands with users who hold the top comment slots on trending videos. SMBs pay to have the user edit their comment to include a one-liner text shoutout (e.g., "Find X at Whole Foods" or "Search for Y App").
* **The Trade-off:** This is pure "billboard" advertising. Without clickable links, advertisers cannot track direct ROI, but it protects the comment from YouTube's spam filters and prevents algorithm ghosting.

---

## 2. Go-to-Market & Operational Feasibility
* **Target Audience:** Consumer packaged goods (CPG), consumer mobile apps, and indie hackers who rely on ambient brand awareness rather than direct-response clicks.
* **Unit Economics:** The platform takes a margin on high-volume, low-cost transactions (e.g., advertiser pays $50, commenter gets $35, platform keeps $15). Requires massive automation to be venture-scalable.
* **Platform Risk:** Modifying comments for commercial purposes sits in a gray area of YouTube's Terms of Service. Sticking to plain-text, unlinked shoutouts minimizes the risk of automated shadowbans.
* **The Cold Start Problem:** The primary business challenge is acquiring the initial pool of advertisers willing to fund campaigns before a large base of commenters is established.

---

## 3. Product Strategy: Automated MVP via OAuth
For the Minimum Viable Product (MVP), the platform relies on **Full Backend Automation** using Google OAuth.
* **Why Automated:** Virality has a brutally short half-life. Relying on users to manually update comments introduces human latency that destroys the value for advertisers. By implementing the restricted `force-ssl` OAuth scope, the platform guarantees instant execution the moment a deal closes.
* **The MVP Flow:** Commenters authenticate via OAuth, granting the platform permission to manage their YouTube account. When an advertiser funds a campaign, the backend instantly fires a request via the YouTube Data API (`comments.update`) to modify the comment text, and immediately triggers the smart contract or backend escrow to release funds to the commenter.
* **The Trade-off (The Audit Trap):** To move fast, the MVP will initially operate as an "Unverified App" within Google Cloud. This means accepting a hard cap of 100 authorized test users and a friction-heavy warning screen during onboarding. The goal of the MVP is to prove advertisers will actually pay for these automated insertions before investing time and capital into passing the official Google security assessment.

---

## 4. System Architecture & Tech Stack
The system is optimized for high-frequency polling, background processing, and frictionless cross-border micro-payments.

### Frontend (Advertiser & Commenter Portals)
* **Stack:** next.js, Tailwind CSS, Shadcn UI.
* **Authentication:** WalletConnect for mobile-first Web3 wallet linking, and basic Google OAuth for YouTube Channel verification.

### Core Backend (API & State Management)
* **Stack:** Go (Golang) for high concurrency and low server costs.
* **Database:** PostgreSQL for persistent state (users, campaigns, transactions) and Redis for managing active polling queues.

### Discovery Engine (The Deterministic Layer)
* **Stack:** A dedicated Go-based background worker (cron job) interacting directly with the PostgreSQL database.
* **Function:** Relies on hard math and velocity thresholds rather than AI to identify comments that are actively going viral in real-time.
* **The Workflow:**
    1. **Discover:** Polls the YouTube `Videos: list` API (filtered by `chart=mostPopular`) to continuously update the roster of trending videos.
    2. **Extract:** Queries the `CommentThreads: list` API (filtered by `order=relevance`) to pull the top-ranked comments for those trending videos.
    3. **Calculate:** Stores the `commentId`, `likeCount`, and timestamp. On the next polling cycle, it calculates the like velocity: $V = \frac{\Delta L}{\Delta t}$.
    4. **Trigger:** If the velocity ($V$) exceeds a hardcoded threshold (e.g., > 100 likes per minute), the system automatically flags the comment as prime real estate and pushes it to the marketplace.
* **Constraint Management:** The polling frequency is strictly orchestrated to avoid burning through the standard 10,000 unit daily quota limit for the YouTube Data API.

### Verification Worker (The Polling Engine)
* **Function:** An isolated background service utilizing Go routines to poll the YouTube API (`comments.list`) for active campaigns. It checks the specific `commentId` every few minutes to verify the user has added the advertiser's exact text, signaling the backend to release the escrow.

### Settlement Layer (Micro-Payments)
* **Function:** Bypasses traditional banking and Stripe fees, which erode margins on small cross-border payments. 
* **Execution:** Utilizes stablecoins (USDC). Advertisers fund campaigns upfront into a non-custodial smart contract or backend escrow. Once the Verification Worker confirms the comment edit, the backend signs a transaction to instantly release the USDC to the commenter's connected wallet.