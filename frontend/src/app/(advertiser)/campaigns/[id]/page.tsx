"use client";

import { useEffect, useState, use } from "react";
import {
  useAccount,
  useWriteContract,
  useWaitForTransactionReceipt,
} from "wagmi";
import { useConnectModal } from "@rainbow-me/rainbowkit";
import { parseUnits, stringToHex, keccak256 } from "viem";
import { baseSepolia } from "wagmi/chains";
import { api, type Campaign, type Deal } from "@/lib/api";
import { USDC_ADDRESS_BASE_SEPOLIA, USDC_ABI, ESCROW_ADDRESS, ESCROW_ABI } from "@/lib/wagmi";
import { Button } from "@/components/ui/button";
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

export default function CampaignDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { isConnected, chainId } = useAccount();
  const { openConnectModal } = useConnectModal();
  const [campaign, setCampaign] = useState<Campaign | null>(null);
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
    api.getCampaign(id).then(setCampaign).catch(() => {});
    api
      .listCampaignDeals(id)
      .then(setDeals)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    if (approveConfirmed && campaign) {
      const amount = parseUnits(
        (campaign.budget_cents / 100).toString(),
        6
      );
      const campaignIdBytes = keccak256(stringToHex(campaign.id));

      depositEscrow({
        address: ESCROW_ADDRESS as `0x${string}`,
        abi: ESCROW_ABI,
        functionName: "deposit",
        args: [campaignIdBytes, amount],
      }).catch((err) => {
        setWalletError(
          err instanceof Error ? err.message : "Failed to deposit to escrow"
        );
      });
    }
  }, [approveConfirmed, campaign, depositEscrow]);

  useEffect(() => {
    if (
      depositConfirmed &&
      depositHash &&
      campaign &&
      campaign.status === "draft" &&
      lastFundedTxHash !== depositHash
    ) {
      setLastFundedTxHash(depositHash);
      api
        .fundCampaign(campaign.id, depositHash)
        .then(() => {
          api.getCampaign(id).then(setCampaign);
        })
        .catch((err) => {
          // If backend already switched status, treat it as success and refresh.
          if (err instanceof Error && err.message.includes("already funded")) {
            api.getCampaign(id).then(setCampaign);
            return;
          }
          setWalletError(
            err instanceof Error ? err.message : "Failed to mark campaign as funded"
          );
        });
    }
  }, [depositConfirmed, depositHash, campaign, id, lastFundedTxHash]);

  const handleFund = async () => {
    if (!campaign) return;
    setWalletError(null);

    if (!ESCROW_ADDRESS || !/^0x[a-fA-F0-9]{40}$/.test(ESCROW_ADDRESS)) {
      setWalletError(
        "Escrow contract is not configured. Set NEXT_PUBLIC_ESCROW_CONTRACT in frontend/.env.local and restart the frontend."
      );
      return;
    }

    if (!isConnected) {
      if (!openConnectModal) {
        setWalletError("Wallet connect modal is unavailable");
        return;
      }
      openConnectModal();
      return;
    }

    if (chainId !== baseSepolia.id) {
      setWalletError("Please switch your wallet network to Base Sepolia.");
      return;
    }

    const amount = parseUnits(
      (campaign.budget_cents / 100).toString(),
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

  if (loading || !campaign) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">Loading campaign...</p>
      </div>
    );
  }

  const paidDeals = deals.filter((d) => d.status === "paid").length;
  const spent = paidDeals * campaign.price_per_placement;

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>&ldquo;{campaign.ad_text}&rdquo;</CardTitle>
            <Badge>{campaign.status}</Badge>
          </div>
          <CardDescription>
            Campaign {campaign.id.slice(0, 8)}...
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-3 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold">
                ${(campaign.budget_cents / 100).toFixed(2)}
              </p>
              <p className="text-xs text-muted-foreground">Budget</p>
            </div>
            <div>
              <p className="text-2xl font-bold">
                ${(campaign.price_per_placement / 100).toFixed(2)}
              </p>
              <p className="text-xs text-muted-foreground">Per Placement</p>
            </div>
            <div>
              <p className="text-2xl font-bold">
                ${(spent / 100).toFixed(2)}
              </p>
              <p className="text-xs text-muted-foreground">Spent</p>
            </div>
          </div>

          {campaign.status === "draft" && (
            <>
              <Separator />
              <div className="text-center">
                <p className="mb-3 text-sm text-muted-foreground">
                  Fund this campaign to start placing ads. Deposit USDC on Base
                  Sepolia.
                </p>
                <Button
                  onClick={handleFund}
                  size="lg"
                >
                  {!isConnected
                    ? "Connect wallet first"
                    : approveHash && !approveConfirmed
                    ? "Approving USDC..."
                    : depositHash && !depositConfirmed
                    ? "Depositing..."
                    : `Fund $${(campaign.budget_cents / 100).toFixed(2)} USDC`}
                </Button>
                {walletError && (
                  <p className="mt-2 text-xs text-destructive">{walletError}</p>
                )}
              </div>
            </>
          )}

          {campaign.escrow_tx_hash && (
            <p className="text-center text-xs text-muted-foreground">
              Escrow tx:{" "}
              <a
                href={`https://sepolia.basescan.org/tx/${campaign.escrow_tx_hash}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline"
              >
                {campaign.escrow_tx_hash.slice(0, 16)}...
              </a>
            </p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Deals ({deals.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {deals.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground">
              No deals yet. Fund the campaign and browse the marketplace to
              create deals.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Deal</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Date</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {deals.map((deal) => (
                  <TableRow key={deal.id}>
                    <TableCell className="font-mono text-xs">
                      {deal.id.slice(0, 8)}...
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          deal.status === "paid" ? "default" : "secondary"
                        }
                      >
                        {deal.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {new Date(deal.created_at).toLocaleDateString()}
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
