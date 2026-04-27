const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("token") : null;

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

export const api = {
  // Auth
  googleCallback: (code: string, role: string) =>
    request<{ token: string; user: User }>("/api/auth/google/callback", {
      method: "POST",
      body: JSON.stringify({ code, role }),
    }),

  getMe: () => request<{ user: User; wallet: Wallet | null; channel: YouTubeChannel | null }>("/api/me"),

  linkWallet: (address: string) =>
    request<Wallet>("/api/wallet/link", {
      method: "POST",
      body: JSON.stringify({ address }),
    }),

  // Marketplace
  listComments: (limit = 20, offset = 0) =>
    request<ViralComment[]>(`/api/marketplace/comments?limit=${limit}&offset=${offset}`),

  getComment: (id: string) =>
    request<ViralComment>(`/api/marketplace/comments/${id}`),

  listCommentsByChannel: (channelId: string) =>
    request<ViralComment[]>(
      `/api/marketplace/comments/by-channel/${encodeURIComponent(channelId)}`
    ),

  listAllCommentsByChannel: (channelId: string) =>
    request<ChannelAuthoredComment[]>(
      `/api/marketplace/comments/by-channel/${encodeURIComponent(channelId)}/all`
    ),

  listAllCommentsByVideo: (videoId: string) =>
    request<ChannelAuthoredComment[]>(
      `/api/marketplace/comments/by-video/${encodeURIComponent(videoId)}/all`
    ),

  registerCommentForTesting: (data: RegisterCommentForTestingInput) =>
    request<ViralComment>("/api/marketplace/comments/register-test", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  listTrendingVideos: (limit = 20) =>
    request<TrendingVideo[]>(`/api/marketplace/videos?limit=${limit}`),

  // Campaigns (advertiser)
  createCampaign: (data: CreateCampaignInput) =>
    request<Campaign>("/api/campaigns", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  listCampaigns: () => request<Campaign[]>("/api/campaigns"),

  getCampaign: (id: string) => request<Campaign>(`/api/campaigns/${id}`),

  fundCampaign: (id: string, txHash: string) =>
    request<{ status: string }>(`/api/campaigns/${id}/fund`, {
      method: "POST",
      body: JSON.stringify({ tx_hash: txHash }),
    }),

  listCampaignDeals: (id: string) =>
    request<Deal[]>(`/api/campaigns/${id}/deals`),

  createDeal: (campaignId: string, commentId: string) =>
    request<Deal>("/api/deals", {
      method: "POST",
      body: JSON.stringify({ campaign_id: campaignId, comment_id: commentId }),
    }),

  // Commenter
  listMyComments: () => request<ViralComment[]>("/api/my/comments"),
  listMyDeals: () => request<Deal[]>("/api/my/deals"),
  listMyTransactions: () => request<Transaction[]>("/api/my/transactions"),
};

// Types
export interface User {
  id: string;
  email: string;
  role: "commenter" | "advertiser";
  created_at: string;
}

export interface YouTubeChannel {
  id: string;
  user_id: string;
  channel_id: string;
  channel_title: string;
  created_at: string;
}

export interface Wallet {
  id: string;
  user_id: string;
  address: string;
  chain: string;
  created_at: string;
}

export interface TrendingVideo {
  id: string;
  video_id: string;
  title: string;
  channel_title: string;
  thumbnail_url: string;
  view_count: number;
  discovered_at: string;
}

export interface ViralComment {
  id: string;
  video_id: string;
  comment_id: string;
  author_channel_id: string;
  author_display_name: string;
  original_text: string;
  like_count: number;
  velocity: number;
  status: "available" | "claimed" | "expired";
  published_at?: string | null;
  updated_at?: string | null;
  first_seen: string;
}

export interface Campaign {
  id: string;
  advertiser_id: string;
  ad_text: string;
  budget_cents: number;
  price_per_placement: number;
  status: "draft" | "funded" | "active" | "completed";
  escrow_tx_hash?: string;
  created_at: string;
}

export interface Deal {
  id: string;
  campaign_id: string;
  comment_id: string;
  commenter_id: string;
  status: "pending" | "edit_pending" | "verified" | "paid" | "failed";
  created_at: string;
}

export interface ChannelAuthoredComment {
  comment_id: string;
  video_id: string;
  author_channel_id: string;
  author_display_name: string;
  text: string;
  like_count: number;
  published_at?: string | null;
  updated_at?: string | null;
  marketplace_comment_id?: string;
  marketplace_status?: ViralComment["status"];
}

export interface Transaction {
  id: string;
  deal_id: string;
  tx_hash: string;
  amount_usdc: number;
  status: "pending" | "confirmed" | "failed";
  created_at: string;
}

export interface CreateCampaignInput {
  ad_text: string;
  budget_cents: number;
  price_per_placement: number;
}

export interface RegisterCommentForTestingInput {
  comment_id: string;
  video_id: string;
  author_channel_id: string;
  author_display_name: string;
  text: string;
  like_count: number;
}
