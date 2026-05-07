import axios from 'axios';
import type { Place, Comment } from './types';

// Use environment variable for API URL in production, fallback to /api for local proxy
const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Add Authorization header if token exists
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 Unauthorized globally
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      // Dispatch a custom event that App.vue can listen to
      window.dispatchEvent(new CustomEvent('auth-error'));
    }
    return Promise.reject(error);
  }
);

export const uploadImage = async (file: File) => {
  const formData = new FormData();
  formData.append('image', file);

  const response = await api.post('/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
  return response.data.url;
};

export const getNearbyPlaces = async (lat: number, lng: number, radiusKm: number = 10) => {
    const response = await api.get<Place[]>(`/places?lat=${lat}&lng=${lng}&radius=${radiusKm}`);
    return response.data;
};

export const getUserComments = async () => {
    const response = await api.get<any[]>('/user?action=comments');
    return response.data;
};

export const getAdminStats = async () => {
    const response = await api.get<any>('/admin?action=stats');
    return response.data;
};

export const setFavoriteStatus = async (placeId: number, shouldBeFavorite: boolean) => {
    if (!shouldBeFavorite) {
        await api.delete(`/favorites?place_id=${placeId}`);
    } else {
        await api.post('/favorites', { place_id: placeId });
    }
};

export const getFavorites = async () => {
    const response = await api.get<Place[]>('/favorites');
    return response.data;
};

export const getLeaderboard = async () => {
    const response = await api.get<any[]>('/leaderboard');
    return response.data;
};

export const translateText = async (text: string, from: string, to: string) => {
    try {
        const response = await axios.get(`https://lingva.dialectapp.org/api/v1/${from}/${to}/${encodeURIComponent(text)}`);
        return response.data.translation;
    } catch (error) {
        console.error('Translation error:', error);
        return null;
    }
};

export default api;
