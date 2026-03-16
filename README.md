# MicroAds — Viral Comment Monetization Platform

A micro-sponsorship marketplace connecting brands with YouTube's top commenters. Advertisers pay for plain-text shoutouts on viral comments. Commenters get paid instantly in USDC on Base.

## Campaigns vs Bounties

- **Campaigns**: Advertiser picks specific comments from the marketplace and pays for edits. One deal per comment; verifier appends ad text and releases payment after verification.
- **Bounties**: Advertiser sets rules (min likes, optional category) and funds a pool. Commenters **hunt** bounties by claiming with their own qualifying comments. First claimants within budget get paid; when the pool is exhausted the bounty is marked **completed** and no longer shown to commenters. The verifier edits the comment and pays out the same way as campaigns (deal status: `pending` → `edit_pending` → `verified` → `paid`).

## Test Page & Bounty Hunt

- **Test page** (`/test/my-comments`): Enter a video ID to load comments. **Advertisers** can create campaign deals or create/fund test bounties. **Commenters** see only their channel’s comments; they can hunt bounties even when a comment isn’t in the marketplace yet (the comment is registered on first claim). Comments that already have a deal (claimed/edited) are hidden so they can’t be claimed again.
- **Bounty Hunt** (`/bounty-hunt`): Commenters see active bounties and their eligible comments; only bounties with remaining budget are listed. Completed bounties are excluded for commenters and shown as completed for advertisers.

## Deal Performance (advertisers)

- **Performance** (`/performance`): Advertisers see all their deals (from campaigns and bounties) in one list. Each row shows the deal status, a **View comment on YouTube** link to the edited comment, current likes and velocity (likes/min), and an optional **velocity over time** graph (likes per minute since the comment was edited). Metrics are recorded when the verifier edits the comment and by a **separate performance worker** that polls YouTube for like counts on edited deals (independent of the discovery engine).

## Architecture

| Component | Tech | Directory |
|-----------|------|-----------|
| **Frontend** | Next.js 16, Tailwind CSS, Shadcn UI, wagmi/viem | `frontend/` |
| **API Server** | Go, Chi router, PostgreSQL, Redis | `backend/cmd/api/` |
| **Discovery Engine** | Go worker, YouTube Data API v3 | `backend/cmd/discovery/` |
| **Verification Worker** | Go worker, YouTube comment polling | `backend/cmd/verifier/` |
| **Performance Worker** | Go worker, polls YouTube for deal comment likes/velocity | `backend/cmd/performance/` |
| **Smart Contract** | Solidity 0.8.24, Foundry, Base chain | `contracts/` |

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- Docker & Docker Compose
- [Foundry](https://book.getfoundry.sh/getting-started/installation) (for smart contracts)

### 1. Start local databases

```bash
docker compose up -d
```

### 2. Run database migrations

```bash
# Install golang-migrate: https://github.com/golang-migrate/migrate
migrate -path backend/migrations -database "postgres://microads:microads_dev@localhost:5432/microads?sslmode=disable" up
```

### 3. Configure environment

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
# Fill in the required values (Google OAuth, YouTube API key, JWT secret, etc.)
```

### 4. Start the Go API

```bash
cd backend && go run ./cmd/api
```

### 5. Start the frontend

```bash
cd frontend && npm run dev
```

### 6. Start workers (separate terminals)

```bash
cd backend && go run ./cmd/discovery    # Discovery engine
cd backend && go run ./cmd/verifier     # Verification worker
cd backend && go run ./cmd/performance  # Deal performance metrics (optional; uses PERFORMANCE_POLL_INTERVAL_MIN, default 5 min)
```

## Smart Contract

The `MicroAdsEscrow` contract on Base holds USDC for campaigns and bounties.

```bash
cd contracts
forge build      # Compile
forge test -vv   # Run tests
```

### Deploy to Base Sepolia

```bash
OPERATOR_ADDRESS=0x... forge script script/Deploy.s.sol --rpc-url base_sepolia --broadcast --private-key $DEPLOYER_KEY
```

## Deployment

- **Frontend**: Deploy `frontend/` to Vercel. Set `NEXT_PUBLIC_API_URL` to the API server URL.
- **Backend**: Deploy via Railway using the `Dockerfile`. Attach managed PostgreSQL and Redis.
  - API: set `CMD` to `/bin/api`
  - Discovery: set `CMD` to `/bin/discovery`
  - Verifier: set `CMD` to `/bin/verifier`
  - Performance: set `CMD` to `/bin/performance` (for deal metrics; optional)

## Environment Variables

See `backend/.env.example` and `frontend/.env.example` for all required configuration.
