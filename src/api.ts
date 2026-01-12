import axios from 'axios';

// Use relative path to leverage Vite's proxy (works for both localhost and IP access)
const API_BASE_URL = '/api';

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
    const response = await api.get<any[]>(`/places?lat=${lat}&lng=${lng}&radius=${radiusKm}`);
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

export const toggleFavorite = async (placeId: number, isFavorite: boolean) => {
    if (isFavorite) {
        await api.delete(`/favorites?place_id=${placeId}`);
    } else {
        await api.post('/favorites', { place_id: placeId });
    }
};

export const getFavorites = async () => {
    const response = await api.get<any[]>('/favorites');
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
