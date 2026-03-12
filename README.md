# MicroAds — Viral Comment Monetization Platform

A micro-sponsorship marketplace connecting brands with YouTube's top commenters. Advertisers pay for plain-text shoutouts on viral comments. Commenters get paid instantly in USDC on Base.

## Architecture

| Component | Tech | Directory |
|-----------|------|-----------|
| **Frontend** | Next.js 16, Tailwind CSS, Shadcn UI, wagmi/viem | `frontend/` |
| **API Server** | Go, Chi router, PostgreSQL, Redis | `backend/cmd/api/` |
| **Discovery Engine** | Go worker, YouTube Data API v3 | `backend/cmd/discovery/` |
| **Verification Worker** | Go worker, YouTube comment polling | `backend/cmd/verifier/` |
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
cd backend && go run ./cmd/discovery   # Discovery engine
cd backend && go run ./cmd/verifier    # Verification worker
```

## Smart Contract

The `MicroAdsEscrow` contract on Base holds USDC for advertising campaigns.

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

## Environment Variables

See `backend/.env.example` and `frontend/.env.example` for all required configuration.
