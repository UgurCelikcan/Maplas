import axios from 'axios';

// Get the current hostname (e.g., 'localhost' or '192.168.1.35')
const hostname = window.location.hostname;

// Backend is running on port 8080 on the same machine
const API_BASE_URL = `http://${hostname}:8080/api`;

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

export default api;
