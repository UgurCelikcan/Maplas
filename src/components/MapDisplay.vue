<script setup lang="ts">
import { onMounted, ref, watch, shallowRef, inject } from 'vue';
import L from 'leaflet';

// Leaflet marker icon fix for Webpack/Vite
// ... (icons are custom divIcons now)

interface Place {
  id: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
}

const props = defineProps<{
  places: Place[];
  selectedPlaceId: number | null;
}>();

const emit = defineEmits<{
  (e: 'select-place', id: number): void;
}>();

const isDarkMode = inject('isDarkMode', ref(true));

const mapContainer = ref<HTMLElement | null>(null);
const map = shallowRef<L.Map | null>(null);
const markerClusterGroup = shallowRef<any>(null);
const markersMap = shallowRef<Map<number, L.Marker>>(new Map());
const tileLayer = shallowRef<L.TileLayer | null>(null);
const routingControl = shallowRef<any>(null);
const userLocation = ref<{ lat: number; lng: number } | null>(null);
const userMarker = shallowRef<L.Marker | null>(null);

// Kategoriye göre renk belirle (Değişmedi)
const getCategoryColor = (category: string) => {
  switch (category) {
    case 'Tarihi': return '#f87171'; // Red 400
    case 'Doğa': return '#4ade80';   // Green 400
    case 'Plaj': return '#60a5fa';   // Blue 400
    case 'Müze': return '#c084fc';   // Purple 400
    case 'Antik Kent': return '#fbbf24'; // Amber 400
    case 'Alışveriş': return '#e91e63'; // Pink
    default: return '#94a3b8';       // Slate 400
  }
};

const getCategoryEmoji = (category: string) => {
    switch (category) {
        case 'Tarihi': return '🏛️';
        case 'Doğa': return '🌲';
        case 'Plaj': return '🏖️';
        case 'Müze': return '🎨';
        case 'Antik Kent': return '🏺';
        case 'Alışveriş': return '🛍️';
        case 'Manzara': return '🌄';
        default: return '📍';
    }
};

function createCustomIcon(category: string, isSelected: boolean = false) {
  const color = getCategoryColor(category);
  const emoji = getCategoryEmoji(category);
  
  return L.divIcon({
    className: 'custom-marker-wrapper',
    html: `
      <div class="marker-pin ${isSelected ? 'selected' : ''}" style="--marker-color: ${color}">
        <span class="marker-emoji">${emoji}</span>
      </div>
    `,
    iconSize: [40, 40],
    iconAnchor: [20, 40],
    popupAnchor: [0, -40]
  });
}

function getTileLayerUrl(isDark: boolean) {
    if (isDark) {
        return 'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png';
    } else {
        return 'https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png';
    }
}

async function locateUser() {
    if (!map.value) return;

    map.value.locate({ setView: true, maxZoom: 13 });

    map.value.once('locationfound', (e: any) => {
        userLocation.value = e.latlng;
        
        if (userMarker.value) {
            userMarker.value.setLatLng(e.latlng);
        } else {
            userMarker.value = L.marker(e.latlng, {
                icon: L.divIcon({
                    className: 'user-marker',
                    html: '<div class="user-dot"></div><div class="user-pulse"></div>',
                    iconSize: [20, 20]
                })
            }).addTo(map.value!);
            userMarker.value.bindPopup("Konumunuz");
        }
    });

    map.value.once('locationerror', () => {
        alert("Konumunuz alınamadı. Lütfen izinleri kontrol edin.");
    });
}

function drawRoute(toLat: number, toLng: number) {
    if (!map.value || !userLocation.value) {
        alert("Lütfen önce konumunuzu belirleyin.");
        locateUser();
        return;
    }

    if (routingControl.value) {
        map.value.removeControl(routingControl.value);
    }

    // @ts-ignore
    if (L.Routing) {
        // @ts-ignore
        routingControl.value = L.Routing.control({
            waypoints: [
                L.latLng(userLocation.value.lat, userLocation.value.lng),
                L.latLng(toLat, toLng)
            ],
            routeWhileDragging: false,
            addWaypoints: false,
            fitSelectedRoutes: true,
            showAlternatives: false,
            lineOptions: {
                styles: [{ color: '#42b883', opacity: 0.8, weight: 6 }],
                extendToWaypoints: true,
                missingRouteTolerance: 0
            },
            // @ts-ignore
            createMarker: () => null // Rota başlangıç/bitiş markerlarını gizle (bizimkiler var)
        }).addTo(map.value);
    }
}

onMounted(() => {
  if (mapContainer.value) {
    map.value = L.map(mapContainer.value, {
        zoomControl: false 
    }).setView([39.0, 35.0], 6);

    tileLayer.value = L.tileLayer(getTileLayerUrl(isDarkMode.value), {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
    }).addTo(map.value);

    L.control.zoom({ position: 'bottomright' }).addTo(map.value);

    // Popup içindeki butona tıklamayı dinle
    map.value.on('popupopen', (e: any) => {
        const popupNode = e.popup._contentNode;
        const btn = popupNode.querySelector('.btn-route');
        if (btn) {
            L.DomEvent.on(btn, 'click', (ev: any) => {
                L.DomEvent.stopPropagation(ev);
                const lat = parseFloat(btn.dataset.lat);
                const lng = parseFloat(btn.dataset.lng);
                drawRoute(lat, lng);
            });
        }
    });

    // Cluster Group'u oluştur
    // @ts-ignore
    if (L.markerClusterGroup) {
        // @ts-ignore
        markerClusterGroup.value = L.markerClusterGroup({
            showCoverageOnHover: false,
            zoomToBoundsOnClick: true,
            spiderfyOnMaxZoom: true,
            iconCreateFunction: function(cluster: any) {
                const count = cluster.getChildCount();
                return L.divIcon({
                    html: `<div class="custom-cluster"><span>${count}</span></div>`,
                    className: 'cluster-wrapper',
                    iconSize: L.point(40, 40)
                });
            }
        });
        map.value.addLayer(markerClusterGroup.value);
    }

    updateMarkers();
  }
});

// Watch theme change to update map tiles
watch(isDarkMode, (newVal) => {
    if (tileLayer.value) {
        tileLayer.value.setUrl(getTileLayerUrl(newVal));
    }
    // Popup style changes via body class, no JS needed here for popups
});

watch(() => props.places, () => {
  updateMarkers();
}, { deep: true });

watch(() => props.selectedPlaceId, (newId, oldId) => {
  if (map.value && markerClusterGroup.value) {
    // Eski seçili ikonunu sıfırla
    if (oldId) {
        const oldMarker = markersMap.value.get(oldId);
        const oldPlace = props.places.find(p => p.id === oldId);
        if (oldMarker && oldPlace) {
            oldMarker.setIcon(createCustomIcon(oldPlace.category, false));
        }
    }

    if (newId) {
      const place = props.places.find(p => p.id === newId);
      const marker = markersMap.value.get(newId);
      if (place && marker) {
        marker.setIcon(createCustomIcon(place.category, true));
        markerClusterGroup.value.zoomToShowLayer(marker, () => {
            marker.openPopup();
        });
      }
    }
  }
});

function updateMarkers() {
  if (!map.value) return;

  // Cluster grubu varsa temizle, yoksa map üzerinden markerları temizle
  if (markerClusterGroup.value) {
    markerClusterGroup.value.clearLayers();
  } else {
    markersMap.value.forEach(marker => marker.remove());
  }
  
  markersMap.value.clear();

  props.places.forEach(place => {
    const marker = L.marker([place.lat, place.lng], {
      icon: createCustomIcon(place.category, props.selectedPlaceId === place.id)
    });

    // Not: Popup içeriği string olduğu için burada class değişikliği yapmak zor.
    // CSS tarafında body.light-theme .custom-popup ile yöneteceğiz.
    const popupContent = `
        <div class="custom-popup">
            <div class="popup-header">
                <span class="popup-category">${getCategoryEmoji(place.category)} ${place.category}</span>
                <span class="popup-city">📍 ${place.city}</span>
            </div>
            <h3>${place.name}</h3>
            <p>${place.description}</p>
            <button class="btn-route" data-lat="${place.lat}" data-lng="${place.lng}">
                🚗 Yol Tarifi Al
            </button>
        </div>
    `;

    marker.bindPopup(popupContent, {
        closeButton: false,
        className: 'modern-popup'
    });
    
    marker.on('click', (e) => {
        L.DomEvent.stopPropagation(e);
        emit('select-place', place.id);
    });

    if (markerClusterGroup.value) {
        markerClusterGroup.value.addLayer(marker);
    } else {
        // Fallback
        marker.addTo(map.value!);
    }
    
    markersMap.value.set(place.id, marker);
  });
}
</script>

<template>
  <div class="map-wrapper">
    <div ref="mapContainer" class="map-container"></div>
    <button class="locate-btn" @click="locateUser" title="Konumumu Bul">
        📍
    </button>
  </div>
</template>

<style>
/* ... (Previous Styles) ... */

/* Route Button in Popup */
.btn-route {
    margin-top: 12px;
    width: 100%;
    background-color: #42b883;
    color: #1a1a1a;
    border: none;
    padding: 8px 12px;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
}
.btn-route:hover {
    background-color: #3aa876;
}

/* User Location Marker */
.user-marker {
    display: flex;
    align-items: center;
    justify-content: center;
}
.user-dot {
    width: 12px;
    height: 12px;
    background-color: #3b82f6;
    border: 2px solid white;
    border-radius: 50%;
    z-index: 2;
}
.user-pulse {
    position: absolute;
    width: 30px;
    height: 30px;
    background-color: rgba(59, 130, 246, 0.4);
    border-radius: 50%;
    animation: pulse 2s infinite;
}
@keyframes pulse {
    0% { transform: scale(0.5); opacity: 1; }
    100% { transform: scale(1.5); opacity: 0; }
}

/* Leaflet Routing Machine Overrides */
.leaflet-routing-container {
    background-color: #18181b !important;
    color: #fff !important;
    border: 1px solid #3f3f46 !important;
    padding: 10px !important;
    border-radius: 12px !important;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5) !important;
}
body.light-theme .leaflet-routing-container {
    background-color: #fff !important;
    color: #0f172a !important;
    border-color: #e2e8f0 !important;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1) !important;
}
.leaflet-routing-alt {
    max-height: 200px;
    overflow-y: auto;
}
.leaflet-routing-alt tr:hover {
    background-color: #27272a !important;
    cursor: pointer;
}
body.light-theme .leaflet-routing-alt tr:hover {
    background-color: #f1f5f9 !important;
}
</style>

<style scoped>
.map-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
    flex-grow: 1;
}
.map-container {
  width: 100%;
  height: 100%;
}
.locate-btn {
    position: absolute;
    bottom: 20px;
    right: 60px; /* Zoom control'ün solunda */
    z-index: 999;
    width: 44px;
    height: 44px;
    background-color: #42b883;
    border: none;
    border-radius: 50%;
    font-size: 1.5rem;
    cursor: pointer;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.2s, background-color 0.2s;
}
.locate-btn:hover {
    transform: scale(1.1);
    background-color: #3aa876;
}
</style>