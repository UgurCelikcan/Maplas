<script setup lang="ts">
import { onMounted, ref, watch, shallowRef, inject } from 'vue';
import L from 'leaflet';

interface Place {
  id: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
  imageUrl?: string;
}

const props = defineProps<{
  places: Place[];
  selectedPlaceId: number | null;
}>();

const emit = defineEmits<{
  (e: 'select-place', id: number): void;
  (e: 'view-comments', id: number): void;
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
            createMarker: () => null
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
        
        const btnComments = popupNode.querySelector('.btn-comments');
        if (btnComments) {
            L.DomEvent.on(btnComments, 'click', (ev: any) => {
                L.DomEvent.stopPropagation(ev);
                const id = parseInt(btnComments.dataset.id);
                emit('view-comments', id);
            });
        }
    });

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

watch(isDarkMode, (newVal) => {
    if (tileLayer.value) {
        tileLayer.value.setUrl(getTileLayerUrl(newVal));
    }
});

watch(() => props.places, () => {
  updateMarkers();
}, { deep: true });

watch(() => props.selectedPlaceId, (newId, oldId) => {
  if (map.value && markerClusterGroup.value) {
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

    const imageHtml = place.imageUrl 
        ? `<div class="w-[calc(100%+40px)] -mx-5 -mt-5 mb-3 h-32 rounded-t-xl overflow-hidden"><img src="${place.imageUrl}" alt="${place.name}" class="w-full h-full object-cover" /></div>` 
        : '';

    const popupContent = `
        <div class="custom-popup">
            ${imageHtml}
            <div class="popup-header">
                <span class="popup-category">${getCategoryEmoji(place.category)} ${place.category}</span>
                <span class="popup-city">📍 ${place.city}</span>
            </div>
            <h3>${place.name}</h3>
            <p>${place.description}</p>
            <button class="btn-route w-full mt-3 bg-emerald-500 hover:bg-emerald-600 text-slate-900 border-none py-2 px-3 rounded-lg font-bold cursor-pointer transition-colors flex items-center justify-center gap-2" data-lat="${place.lat}" data-lng="${place.lng}">
                🚗 Yol Tarifi Al
            </button>
            <button class="btn-comments w-full mt-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 text-slate-800 dark:text-white border-none py-2 px-3 rounded-lg font-medium cursor-pointer transition-colors flex items-center justify-center gap-2" data-id="${place.id}">
                💬 Yorumlar & Puan Ver
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
        marker.addTo(map.value!);
    }
    
    markersMap.value.set(place.id, marker);
  });
}
</script>

<template>
  <div class="relative w-full h-full flex-grow">
    <div ref="mapContainer" class="w-full h-full z-0"></div>
    <button class="absolute bottom-5 right-[60px] z-[999] w-11 h-11 bg-emerald-500 border-none rounded-full text-2xl cursor-pointer shadow-lg flex items-center justify-center transition-all hover:scale-110 hover:bg-emerald-600" @click="locateUser" title="Konumumu Bul">
        📍
    </button>
  </div>
</template>
