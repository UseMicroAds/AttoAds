"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, type ViralComment, type Deal, type Transaction } from "@/lib/api";
import { useAccount, useDisconnect } from "wagmi";
import { useConnectModal } from "@rainbow-me/rainbowkit";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export default function CommenterDashboard() {
  const { user, wallet, channel, refresh } = useAuth();
  const { address, isConnected } = useAccount();
  const { disconnect } = useDisconnect();
  const { openConnectModal } = useConnectModal();

  const [comments, setComments] = useState<ViralComment[]>([]);
  const [deals, setDeals] = useState<Deal[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [walletActionLoading, setWalletActionLoading] = useState(false);
  const [walletActionMessage, setWalletActionMessage] = useState<string | null>(
    null
  );
  const [walletActionError, setWalletActionError] = useState<string | null>(
    null
  );

  useEffect(() => {
    if (user) {
      api.listMyComments().then(setComments).catch(() => {});
      api.listMyDeals().then(setDeals).catch(() => {});
      api.listMyTransactions().then(setTransactions).catch(() => {});
    }
  }, [user]);

  const handleConnectWallet = () => {
    setWalletActionMessage(null);
    setWalletActionError(null);
    openConnectModal?.();
  };

  const handleUpdateWalletAddress = async () => {
    setWalletActionMessage(null);
    setWalletActionError(null);

    if (!isConnected || !address) {
      setWalletActionError("Connect a wallet first.");
      return;
    }

    if (isWalletSynced) {
      setWalletActionMessage("Connected wallet is already saved.");
      return;
    }

    setWalletActionLoading(true);
    try {
      await api.linkWallet(address);
      await refresh();
      setWalletActionMessage("Wallet address updated.");
    } catch (err) {
      setWalletActionError(
        err instanceof Error ? err.message : "Failed to update wallet address."
      );
    } finally {
      setWalletActionLoading(false);
    }
  };

  const handleDisconnectWallet = () => {
    setWalletActionMessage(null);
    setWalletActionError(null);
    disconnect();
    setWalletActionMessage("Wallet disconnected. Connect another account.");
  };

  const isWalletSynced =
    !!wallet &&
    !!address &&
    wallet.address.toLowerCase() === address.toLowerCase();

  const totalEarnings = transactions
    .filter((t) => t.status === "confirmed")
    .reduce((sum, t) => sum + t.amount_usdc, 0);

  if (!user) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-muted-foreground">Please sign in to view your dashboard.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Commenter Dashboard</h1>
        <p className="text-muted-foreground">
          Manage your viral comments and earnings
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>YouTube Channel</CardDescription>
            <CardTitle className="text-lg">
              {channel?.channel_title || "Not linked"}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Wallet</CardDescription>
            <CardTitle className="text-lg">
              {wallet ? (
                <span className="font-mono text-sm">
                  {wallet.address.slice(0, 6)}...{wallet.address.slice(-4)}
                </span>
              ) : (
                <span className="text-sm text-muted-foreground">Not linked</span>
              )}
            </CardTitle>
            <CardDescription className="font-mono text-xs">
              {isConnected && address
                ? `Connected: ${address.slice(0, 6)}...${address.slice(-4)}`
                : "No wallet connected"}
            </CardDescription>
            <div className="pt-2">
              {!isConnected ? (
                <Button size="sm" onClick={handleConnectWallet}>
                  Connect Wallet
                </Button>
              ) : (
                <div className="flex gap-2">
                  <Button size="sm" onClick={handleDisconnectWallet}>
                    Disconnect Wallet
                  </Button>
                  <Button
                    size="sm"
                    onClick={handleUpdateWalletAddress}
                    disabled={walletActionLoading}
                  >
                    {walletActionLoading
                      ? "Updating..."
                      : "Update Wallet Address"}
                  </Button>
                </div>
              )}
            </div>
            {walletActionMessage && (
              <p className="pt-2 text-xs text-muted-foreground">
                {walletActionMessage}
              </p>
            )}
            {walletActionError && (
              <p className="pt-2 text-xs text-destructive">{walletActionError}</p>
            )}
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Earnings</CardDescription>
            <CardTitle className="text-lg">
              ${(totalEarnings / 100).toFixed(2)} USDC
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <Tabs defaultValue="comments">
        <TabsList>
          <TabsTrigger value="comments">
            My Comments ({comments.length})
          </TabsTrigger>
          <TabsTrigger value="deals">Deals ({deals.length})</TabsTrigger>
          <TabsTrigger value="earnings">
            Earnings ({transactions.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="comments" className="mt-4">
          {comments.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                No viral comments detected for your channel yet.
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-3">
              {comments.map((c) => (
                <Card key={c.id}>
                  <CardContent className="flex items-center justify-between py-4">
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm">{c.original_text}</p>
                      <p className="text-xs text-muted-foreground">
                        {c.like_count.toLocaleString()} likes
                      </p>
                    </div>
                    <Badge variant={c.status === "available" ? "default" : "secondary"}>
                      {c.status}
                    </Badge>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="deals" className="mt-4">
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
        </TabsContent>

        <TabsContent value="earnings" className="mt-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Tx Hash</TableHead>
                <TableHead>Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {transactions.map((t) => (
                <TableRow key={t.id}>
                  <TableCell>${(t.amount_usdc / 100).toFixed(2)}</TableCell>
                  <TableCell>
                    <Badge
                      variant={
                        t.status === "confirmed" ? "default" : "secondary"
                      }
                    >
                      {t.status}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <a
                      href={`https://sepolia.basescan.org/tx/${t.tx_hash}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-mono text-xs text-primary hover:underline"
                    >
                      {t.tx_hash.slice(0, 10)}...
                    </a>
                  </TableCell>
                  <TableCell>
                    {new Date(t.created_at).toLocaleDateString()}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TabsContent>
      </Tabs>
    </div>
  );
}
