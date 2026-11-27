import axios from 'axios';
import type { Place } from './types';

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

// --- Translation with cache ---
// Each text is translated once per language, remembered in localStorage,
// and at most 3 requests run at the same time.
const TRANSLATION_STORE_KEY = 'translations-v1';
const MAX_STORED_TRANSLATIONS = 500;
const MAX_PARALLEL_TRANSLATIONS = 3;

const storedTranslations: Record<string, string> = (() => {
    try { return JSON.parse(localStorage.getItem(TRANSLATION_STORE_KEY) || '{}'); } catch { return {}; }
})();
const pendingTranslations = new Map<string, Promise<string | null>>();
let activeTranslations = 0;
const translationQueue: Array<() => void> = [];

function saveTranslation(key: string, value: string) {
    storedTranslations[key] = value;
    const keys = Object.keys(storedTranslations);
    for (const old of keys.slice(0, Math.max(0, keys.length - MAX_STORED_TRANSLATIONS))) {
        delete storedTranslations[old];
    }
    try { localStorage.setItem(TRANSLATION_STORE_KEY, JSON.stringify(storedTranslations)); } catch { /* storage full */ }
}

async function withTranslationSlot<T>(task: () => Promise<T>): Promise<T> {
    if (activeTranslations >= MAX_PARALLEL_TRANSLATIONS) {
        await new Promise<void>(resolve => translationQueue.push(resolve));
    }
    activeTranslations++;
    try {
        return await task();
    } finally {
        activeTranslations--;
        translationQueue.shift()?.();
    }
}

export const translateText = (text: string, from: string, to: string): Promise<string | null> => {
    if (!text || !text.trim()) return Promise.resolve(null);
    const key = `${from}|${to}|${text}`;
    if (key in storedTranslations) return Promise.resolve(storedTranslations[key] ?? null);
    const pending = pendingTranslations.get(key);
    if (pending) return pending;

    const request = withTranslationSlot(async () => {
        try {
            const response = await axios.get(`https://lingva.dialectapp.org/api/v1/${from}/${to}/${encodeURIComponent(text)}`, { timeout: 8000 });
            const translation: string | null = response.data?.translation ?? null;
            if (translation) saveTranslation(key, translation);
            return translation;
        } catch (error) {
            console.error('Translation error:', error);
            return null;
        } finally {
            pendingTranslations.delete(key);
        }
    });
    pendingTranslations.set(key, request);
    return request;
};

export default api;
