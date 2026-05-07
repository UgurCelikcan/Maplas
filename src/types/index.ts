export interface Place {
  id?: number;
  name: Record<string, string>;
  description: Record<string, string>;
  lat: number;
  lng: number;
  category: string;
  city: string;
  imageUrl?: string;
  is_favorite?: boolean;
  status?: string;
}

export interface User {
  username: string;
  role: string;
  email?: string;
  bio?: string;
  avatar_url?: string;
  points?: number;
}

export interface Comment {
  id: number;
  place_id: number;
  content: string;
  rating: number;
  created_at: string;
}
