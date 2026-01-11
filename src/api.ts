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

export default api;
