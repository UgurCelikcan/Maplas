import L from 'leaflet';

// Default: OpenStreetMap tiles (no API key needed).
// To use another provider, set these in .env (Vite reads them at build time):
//   VITE_MAP_TILE_URL=https://.../{z}/{x}/{y}.png?key=...
//   VITE_MAP_TILE_URL_DARK=https://.../{z}/{x}/{y}.png?key=...   (optional)
//   VITE_MAP_TILE_ATTRIBUTION=&copy; Provider
const OSM_URL = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
const OSM_ATTRIBUTION = '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>';

const env = import.meta.env;
const lightUrl: string = env.VITE_MAP_TILE_URL || OSM_URL;
const darkUrl: string = env.VITE_MAP_TILE_URL_DARK || '';
const attribution: string = env.VITE_MAP_TILE_ATTRIBUTION || OSM_ATTRIBUTION;

export function createLightLayer(): L.TileLayer {
  return L.tileLayer(lightUrl, { attribution, maxZoom: 19 });
}

// With no dark tile URL, the light tiles are darkened with a CSS filter (.map-tiles-dark in style.css).
export function createDarkLayer(): L.TileLayer {
  if (darkUrl) return L.tileLayer(darkUrl, { attribution, maxZoom: 19 });
  return L.tileLayer(lightUrl, { attribution, maxZoom: 19, className: 'map-tiles-dark' });
}
