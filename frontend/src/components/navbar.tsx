"use client";

import Image from "next/image";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import { buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

export function Navbar() {
  const { user, logout, loading } = useAuth();

  return (
    <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4">
        <Link href="/" className="flex items-center gap-2">
          <Image
            src="/attoads-horizontal-logo.png"
            alt="AttoAds logo"
            width={200}
            height={60}
            className="h-10 w-auto"
            priority
          />
        </Link>

        <nav className="hidden items-center gap-6 md:flex">
          <Link
            href="/marketplace"
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Marketplace
          </Link>
          <Link
            href="/test/my-comments"
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Test Comments
          </Link>
          {user?.role === "advertiser" && (
            <>
              <Link
                href="/campaigns"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                Campaigns
              </Link>
              <Link
                href="/bounties"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                Bounties
              </Link>
              <Link
                href="/performance"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                Performance
              </Link>
            </>
          )}
          {user?.role === "commenter" && (
            <>
              <Link
                href="/dashboard"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                Dashboard
              </Link>
              <Link
                href="/bounty-hunt"
                className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                Bounty Hunt
              </Link>
            </>
          )}
        </nav>

        <div className="flex items-center gap-3">
          {loading ? null : user ? (
            <DropdownMenu>
              <DropdownMenuTrigger className="relative flex h-9 w-9 items-center justify-center rounded-full outline-none hover:bg-muted">
                <Avatar className="h-9 w-9">
                  <AvatarFallback>
                    {user.email.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56">
                <div className="px-2 py-1.5">
                  <p className="text-sm font-medium">{user.email}</p>
                  <p className="text-xs text-muted-foreground capitalize">
                    {user.role}
                  </p>
                </div>
                <DropdownMenuSeparator />
                {user.role === "commenter" && (
                  <>
                    <DropdownMenuItem>
                      <Link href="/dashboard">Dashboard</Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Link href="/bounty-hunt">Bounty Hunt</Link>
                    </DropdownMenuItem>
                  </>
                )}
                {user.role === "advertiser" && (
                  <>
                    <DropdownMenuItem>
                      <Link href="/campaigns">Campaigns</Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Link href="/bounties">Bounties</Link>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <Link href="/performance">Performance</Link>
                    </DropdownMenuItem>
                  </>
                )}
                <DropdownMenuItem>
                  <Link href="/test/my-comments">Test Comments</Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={logout}>Log out</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <Link href="/login" className={buttonVariants({ size: "sm" })}>
              Sign In
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
