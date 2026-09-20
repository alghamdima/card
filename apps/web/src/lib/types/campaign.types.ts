export interface TextFieldConfig {
  id: string;
  name: string;
  label: string;
  placeholder?: string;
  x: number;
  y: number;
  width: number;
  height: number;
  fontSize: number;
  color: string;
  weight?: 'bold' | 'regular';
  align?: 'center' | 'right' | 'left';
  maxChars?: number;
  required?: boolean;
  order: number;
}

export interface TemplateVariant {
  image: string;
  fields: TextFieldConfig[];
}

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
  templateAR?: TemplateVariant;
  templateEN?: TemplateVariant;
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
  lang?: string;
  fieldValues?: Record<string, any>;
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

export interface CampaignAnalytics {
  slug: string;
  title: string;
  totalCards: number;
  cardsToday: number;
  lastCardAt?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}
