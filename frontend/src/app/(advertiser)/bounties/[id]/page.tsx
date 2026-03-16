"use client";

import { useEffect, useState, use } from "react";
import Link from "next/link";
import {
  useAccount,
  useWriteContract,
  useWaitForTransactionReceipt,
} from "wagmi";
import { useConnectModal } from "@rainbow-me/rainbowkit";
import { parseUnits, stringToHex, keccak256 } from "viem";
import { baseSepolia } from "wagmi/chains";
import { api, type Bounty, type Deal } from "@/lib/api";
import { USDC_ADDRESS_BASE_SEPOLIA, USDC_ABI, ESCROW_ADDRESS, ESCROW_ABI } from "@/lib/wagmi";
import { Button } from "@/components/ui/button";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Separator } from "@/components/ui/separator";

export default function BountyDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { isConnected, chainId } = useAccount();
  const { openConnectModal } = useConnectModal();
  const [bounty, setBounty] = useState<Bounty | null>(null);
  const [deals, setDeals] = useState<Deal[]>([]);
  const [loading, setLoading] = useState(true);
  const [walletError, setWalletError] = useState<string | null>(null);
  const [lastFundedTxHash, setLastFundedTxHash] = useState<string | null>(null);

  const { writeContractAsync: approveUSDC, data: approveHash } = useWriteContract();
  const { writeContractAsync: depositEscrow, data: depositHash } = useWriteContract();

  const { isSuccess: approveConfirmed } = useWaitForTransactionReceipt({
    hash: approveHash,
  });

  const { isSuccess: depositConfirmed } = useWaitForTransactionReceipt({
    hash: depositHash,
  });

  useEffect(() => {
    api.getBounty(id).then(setBounty).catch(() => {});
    api
      .listBountyDeals(id)
      .then(setDeals)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    if (approveConfirmed && bounty) {
      const amount = parseUnits(
        (bounty.budget_cents / 100).toString(),
        6
      );
      const bountyIdBytes = keccak256(stringToHex(bounty.id));

      depositEscrow({
        address: ESCROW_ADDRESS as `0x${string}`,
        abi: ESCROW_ABI,
        functionName: "deposit",
        args: [bountyIdBytes, amount],
      }).catch((err) => {
        setWalletError(
          err instanceof Error ? err.message : "Failed to deposit to escrow"
        );
      });
    }
  }, [approveConfirmed, bounty, depositEscrow]);

  useEffect(() => {
    if (
      depositConfirmed &&
      depositHash &&
      bounty &&
      bounty.status === "draft" &&
      lastFundedTxHash !== depositHash
    ) {
      setLastFundedTxHash(depositHash);
      api
        .fundBounty(bounty.id, depositHash)
        .then(() => {
          api.getBounty(id).then(setBounty);
        })
        .catch((err) => {
          if (err instanceof Error && err.message.includes("already funded")) {
            api.getBounty(id).then(setBounty);
            return;
          }
          setWalletError(
            err instanceof Error ? err.message : "Failed to mark bounty as funded"
          );
        });
    }
  }, [depositConfirmed, depositHash, bounty, id, lastFundedTxHash]);

  const handleFund = async () => {
    if (!bounty) return;
    setWalletError(null);

    if (!ESCROW_ADDRESS || !/^0x[a-fA-F0-9]{40}$/.test(ESCROW_ADDRESS)) {
      setWalletError(
        "Escrow contract is not configured. Set NEXT_PUBLIC_ESCROW_CONTRACT in frontend/.env.local and restart the frontend."
      );
      return;
    }

    if (!isConnected) {
      if (openConnectModal) openConnectModal();
      else setWalletError("Wallet connect modal is unavailable");
      return;
    }

    if (chainId !== baseSepolia.id) {
      setWalletError("Please switch your wallet network to Base Sepolia.");
      return;
    }

    const amount = parseUnits(
      (bounty.budget_cents / 100).toString(),
      6
    );

    try {
      await approveUSDC({
        address: USDC_ADDRESS_BASE_SEPOLIA as `0x${string}`,
        abi: USDC_ABI,
        functionName: "approve",
        args: [ESCROW_ADDRESS as `0x${string}`, amount],
      });
    } catch (err) {
      setWalletError(
        err instanceof Error ? err.message : "Failed to approve USDC"
      );
    }
  };

  if (loading || !bounty) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">Loading bounty...</p>
      </div>
    );
  }

  const paidDeals = deals.filter((d) => d.status === "paid").length;
  const spent = paidDeals * bounty.amount_per_claim_cents;

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/bounties" className={buttonVariants({ variant: "ghost" })}>
          ← Bounties
        </Link>
      </div>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>&ldquo;{bounty.ad_text}&rdquo;</CardTitle>
            <Badge>{bounty.status}</Badge>
          </div>
          <CardDescription>
            Bounty {bounty.id.slice(0, 8)}...
            {bounty.min_likes > 0 && ` · Min ${bounty.min_likes} likes`}
            {bounty.video_category && ` · Category: ${bounty.video_category}`}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-3 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold">
                ${(bounty.budget_cents / 100).toFixed(2)}
              </p>
              <p className="text-xs text-muted-foreground">Budget</p>
            </div>
            <div>
              <p className="text-2xl font-bold">
                ${(bounty.amount_per_claim_cents / 100).toFixed(2)}
              </p>
              <p className="text-xs text-muted-foreground">Per Claim</p>
            </div>
            <div>
              <p className="text-2xl font-bold">
                ${(spent / 100).toFixed(2)}
              </p>
              <p className="text-xs text-muted-foreground">Paid Out</p>
            </div>
          </div>

          {bounty.status === "draft" && (
            <>
              <Separator />
              <div className="text-center">
                <p className="mb-3 text-sm text-muted-foreground">
                  Fund this bounty so commenters can claim. Deposit USDC on Base Sepolia.
                </p>
                <Button onClick={handleFund} size="lg">
                  {!isConnected
                    ? "Connect wallet first"
                    : approveHash && !approveConfirmed
                    ? "Approving USDC..."
                    : depositHash && !depositConfirmed
                    ? "Depositing..."
                    : `Fund $${(bounty.budget_cents / 100).toFixed(2)} USDC`}
                </Button>
                {walletError && (
                  <p className="mt-2 text-xs text-destructive">{walletError}</p>
                )}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Claims</CardTitle>
          <CardDescription>Deals from commenters who hunted this bounty</CardDescription>
        </CardHeader>
        <CardContent>
          {deals.length === 0 ? (
            <p className="text-sm text-muted-foreground">No claims yet.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Deal ID</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {deals.map((d) => (
                  <TableRow key={d.id}>
                    <TableCell className="font-mono text-xs">
                      {d.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={d.status === "paid" ? "default" : "secondary"}
                      >
                        {d.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {new Date(d.created_at).toLocaleDateString()}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
