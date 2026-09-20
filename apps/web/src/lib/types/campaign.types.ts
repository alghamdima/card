export interface BoxItem {
  top: number;
  bottom: number;
  centerX: number;
  maxW: number;
  size: number;
  color?: string;
  weight?: string;
  headColor?: string;
  bodyColor?: string;
}

export interface BoxesConfig {
  to: BoxItem;
  message: BoxItem;
  from: BoxItem;
}

export interface Campaign {
  id: number;
  slug: string;
  title: string;
  lang: string;
  textColor: string;
  headColor: string;
  boxes: BoxesConfig;
  image: string;
  thumb?: string;
  active: boolean;
  createdAt: string;
}

export interface CampaignSummary {
  slug: string;
  title: string;
  lang: string;
  textColor: string;
  headColor: string;
  thumb?: string;
  active: boolean;
  totalCards: number;
  createdAt: string;
}

export interface Card {
  id: number;
  campaignSlug: string;
  from: string;
  to: string;
  message: string;
  heading?: string;
  device: string;
  date: string;
  time: string;
  createdAt: string;
}

export interface DashboardStats {
  totalCampaigns: number;
  activeCampaigns: number;
  totalCards: number;
  cardsToday: number;
  deviceStats: Record<string, number>;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}
