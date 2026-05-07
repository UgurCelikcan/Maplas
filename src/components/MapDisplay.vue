<script setup lang="ts">
import { onMounted, ref, watch, shallowRef, inject } from 'vue';
import { useI18n } from 'vue-i18n';
import L from 'leaflet';
import { getLocalizedContent, getCategoryColor, getCategoryEmoji } from '../utils';
import { translateText } from '../api';
import type { Place } from '../types';
import RoutePanel from './map/RoutePanel.vue';
import { useMapRouting } from '../composables/useMapRouting';

const { t, locale } = useI18n();

const props = defineProps<{
  places: Place[];
  selectedPlaceId: number | null;
}>();

const emit = defineEmits<{
  (e: 'select-place', id: number): void;
  (e: 'view-comments', id: number): void;
  (e: 'toggle-favorite', id: number): void;
  (e: 'location-updated', location: { lat: number; lng: number }): void;
}>();

const isDarkMode = inject('isDarkMode', ref(true));

// Map & Markers State
const mapContainer = ref<HTMLElement | null>(null);
const map = shallowRef<L.Map | null>(null);
const markerClusterGroup = shallowRef<any>(null);
const markersMap = shallowRef<Map<number, L.Marker>>(new Map());
const userLocation = ref<{ lat: number; lng: number } | null>(null);
const userMarker = shallowRef<L.Marker | null>(null);
const transportMode = ref('driving');

// Routing Logic (Extracted)
const {
    routeWaypoints,
    routeInfo,
    clearRoute,
    removeWaypoint,
    moveWaypoint,
    updateRouteDisplay,
    shareRoute
} = useMapRouting(map, transportMode);

watch(transportMode, () => {
    if (routeWaypoints.value.length >= 2) {
        updateRouteDisplay();
    }
});

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

async function locateUser() {
    if (!map.value) return;
    map.value.locate({ setView: true, maxZoom: 13 });

    map.value.once('locationfound', (e: any) => {
        userLocation.value = e.latlng;
        emit('location-updated', { lat: e.latlng.lat, lng: e.latlng.lng });
        
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
            userMarker.value.bindPopup(t('map.your_location'));
        }
        
        const startIndex = routeWaypoints.value.findIndex(wp => wp.name === t('map.your_location'));
        if (startIndex !== -1) {
            routeWaypoints.value[startIndex] = { lat: e.latlng.lat, lng: e.latlng.lng, name: t('map.your_location') };
            updateRouteDisplay();
        } else if (routeWaypoints.value.length > 0) {
            routeWaypoints.value.unshift({ lat: e.latlng.lat, lng: e.latlng.lng, name: t('map.your_location') });
            updateRouteDisplay();
        }
    });

    map.value.once('locationerror', () => alert(t('map.location_not_found')));
}

function addToRoute(lat: number, lng: number, name: string) {
    if (routeWaypoints.value.length === 0 && userLocation.value) {
        routeWaypoints.value.push({ lat: userLocation.value.lat, lng: userLocation.value.lng, name: t('map.your_location') });
    }
    if (routeWaypoints.value.some(wp => wp.lat === lat && wp.lng === lng)) return;
    routeWaypoints.value.push({ lat, lng, name });
    updateRouteDisplay();
}

function updateMarkers() {
  if (!map.value) return;
  const newPlaceIds = new Set(props.places.map(p => p.id));
  
  markersMap.value.forEach((marker, id) => {
    if (!newPlaceIds.has(id)) {
        if (markerClusterGroup.value) markerClusterGroup.value.removeLayer(marker);
        else marker.remove();
        markersMap.value.delete(id);
    }
  });

  props.places.forEach(place => {
    const existingMarker = markersMap.value.get(place.id as number);
    const isSelected = props.selectedPlaceId === place.id;
    const customIcon = createCustomIcon(place.category, isSelected);
    const popupContent = generatePopupContent(place);

    if (existingMarker) {
        existingMarker.setIcon(customIcon);
        const popup = existingMarker.getPopup();
        if (popup && popup.getContent() !== popupContent) {
            popup.setContent(popupContent);
            if (popup.isOpen()) {
                setTimeout(() => {
                    // @ts-ignore
                    if (popup._contentNode) attachPopupEvents(popup._contentNode, place);
                }, 50);
            }
        }
    } else {
        const marker = L.marker([place.lat, place.lng], { icon: customIcon });
        (marker as any).placeId = place.id;
        marker.bindPopup(popupContent, { closeButton: false, className: 'modern-popup' });
        marker.on('click', (e) => {
            L.DomEvent.stopPropagation(e);
            emit('select-place', place.id as number);
        });
        if (markerClusterGroup.value) markerClusterGroup.value.addLayer(marker);
        else marker.addTo(map.value!);
        markersMap.value.set(place.id as number, marker);
    }
  });
}

function setRoute(places: Place[], mode?: string) {
    if (mode) transportMode.value = mode;
    clearRoute();
    places.forEach(p => addToRoute(p.lat, p.lng, getLocalizedContent(p.name, locale.value)));
}

async function handleEmergency() {
    if (!navigator.geolocation) {
        alert("Konum servisi desteklenmiyor.");
        return;
    }
    const confirmSOS = confirm("🚨 ACİL DURUM MODU 🚨\n\nEn yakın sağlık kuruluşuna rota oluşturulsun mu?");
    if (!confirmSOS) return;

    navigator.geolocation.getCurrentPosition(async (pos) => {
        const { latitude, longitude } = pos.coords;
        userLocation.value = { lat: latitude, lng: longitude };
        if (map.value) {
             map.value.setView([latitude, longitude], 14);
             if (!userMarker.value) {
                 userMarker.value = L.marker([latitude, longitude], {
                    icon: L.divIcon({ className: 'user-marker', html: '<div class="user-dot"></div>', iconSize: [20, 20] })
                 }).addTo(map.value);
             } else userMarker.value.setLatLng([latitude, longitude]);
        }
        try {
            const query = `[out:json];(node["amenity"~"hospital|clinic|doctors"](around:5000,${latitude},${longitude});way["amenity"~"hospital|clinic|doctors"](around:5000,${latitude},${longitude});relation["amenity"~"hospital|clinic|doctors"](around:5000,${latitude},${longitude}););out center;`;
            const res = await fetch(`https://overpass-api.de/api/interpreter?data=${encodeURIComponent(query)}`);
            const data = await res.json();
            if (data.elements && data.elements.length > 0) {
                let closest: any = null;
                let minDist = Infinity;
                data.elements.forEach((el: any) => {
                    const elLat = el.lat || el.center.lat;
                    const elLng = el.lon || el.center.lon;
                    const dist = Math.sqrt(Math.pow(elLat - latitude, 2) + Math.pow(elLng - longitude, 2));
                    if (dist < minDist) { minDist = dist; closest = { ...el, lat: elLat, lng: elLng }; }
                });
                if (closest) {
                    clearRoute();
                    const name = closest.tags.name || (closest.tags.amenity === 'hospital' ? 'Hastane' : 'Sağlık Merkezi');
                    routeWaypoints.value = [{ lat: latitude, lng: longitude, name: '🚨 Konumum' }, { lat: closest.lat, lng: closest.lng, name: `🏥 ${name}` }];
                    updateRouteDisplay();
                    alert(`🚨 En yakın sağlık kuruluşu bulundu: ${name}\nRota oluşturuldu!`);
                }
            } else alert("Yakınlarda (5km) sağlık kuruluşu bulunamadı. Lütfen 112'yi arayın!");
        } catch (e) { alert("Hata oluştu. Lütfen manuel arama yapın."); }
    }, () => alert("Konum alınamadı."));
}

function openGoogleMapsTransit() {
    if (routeWaypoints.value.length < 2) return;
    const start = routeWaypoints.value[0];
    const end = routeWaypoints.value[routeWaypoints.value.length - 1];
    if (start && end) {
        window.open(`https://www.google.com/maps/dir/?api=1&origin=${start.lat},${start.lng}&destination=${end.lat},${end.lng}&travelmode=transit`, '_blank');
    }
}

function generatePopupContent(place: Place) {
    const placeName = getLocalizedContent(place.name, locale.value);
    const placeDesc = getLocalizedContent(place.description, locale.value);
    const imageHtml = place.imageUrl ? `<div class="w-[calc(100%+40px)] -mx-5 -mt-5 mb-3 h-32 rounded-t-xl overflow-hidden"><img src="${place.imageUrl}" alt="${placeName}" class="w-full h-full object-cover" /></div>` : '';
    return `<div class="custom-popup">${imageHtml}<div class="popup-header"><span class="popup-category">${getCategoryEmoji(place.category)} ${t(`categories.${place.category}`)}</span><button class="btn-favorite bg-transparent border-none cursor-pointer p-1 transition-transform active:scale-75" data-id="${place.id}"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="${place.is_favorite ? '#ef4444' : 'none'}" stroke="${place.is_favorite ? '#ef4444' : 'currentColor'}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path></svg></button></div><div class="flex justify-between items-center mb-1"><span class="popup-city text-[10px] text-slate-400">📍 ${place.city}</span></div><h3>${placeName}</h3><p>${placeDesc}</p><div class="weather-info mt-2 text-xs text-slate-500 flex items-center gap-1"><span class="weather-loading">🌤️ Hava durumu yükleniyor...</span></div><button class="btn-add-route w-full mt-3 bg-emerald-500 hover:bg-emerald-600 text-white border-none py-2.5 px-3 rounded-lg font-bold cursor-pointer transition-colors flex items-center justify-center gap-2" data-lat="${place.lat}" data-lng="${place.lng}" data-name="${placeName}">🚩 Rotaya Ekle</button><button class="btn-comments w-full mt-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 text-slate-800 dark:text-white border-none py-2 px-3 rounded-lg font-medium cursor-pointer transition-colors flex items-center justify-center gap-2" data-id="${place.id}">💬 ${t('map.view_comments')}</button></div>`;
}

function attachPopupEvents(popupNode: HTMLElement, place: Place) {
    const btnAddRoute = popupNode.querySelector('.btn-add-route') as HTMLElement;
    const btnComments = popupNode.querySelector('.btn-comments') as HTMLElement;
    const btnFavorite = popupNode.querySelector('.btn-favorite') as HTMLElement;
    const weatherContainer = popupNode.querySelector('.weather-info');
    const descContainer = popupNode.querySelector('p');

    if (descContainer) {
        const originalText = descContainer.innerText;
        translateText(originalText, 'auto', locale.value).then(translated => {
            if (translated && translated !== originalText) descContainer.innerHTML = `${translated} <span class="text-[9px] text-emerald-500 font-bold ml-1 bg-emerald-50 px-1 rounded border border-emerald-100 cursor-help" title="Otomatik Çevrildi">🌐</span>`;
        }).catch(() => {});
    }

    if (btnAddRoute) {
        if (weatherContainer) {
            fetch(`https://api.open-meteo.com/v1/forecast?latitude=${place.lat}&longitude=${place.lng}&current_weather=true&daily=weathercode,temperature_2m_max,temperature_2m_min&timezone=auto`)
                .then(res => res.json()).then(data => {
                    const temp = Math.round(data.current_weather.temperature);
                    const wind = data.current_weather.windspeed;
                    weatherContainer.innerHTML = `<div class="flex items-center gap-2 mb-1"><span class="text-2xl">🌡️</span><div><span class="font-bold text-slate-800 dark:text-white text-sm">${temp}°C</span><span class="text-[10px] text-slate-500 block">💨 ${wind} km/s</span></div></div>`;
                }).catch(() => weatherContainer.innerHTML = '<span class="text-red-400 text-[10px]">Hava durumu yok</span>');
        }
        L.DomEvent.on(btnAddRoute, 'click', (ev) => { L.DomEvent.stopPropagation(ev); addToRoute(place.lat, place.lng, getLocalizedContent(place.name, locale.value)); });
    }
    if (btnComments) L.DomEvent.on(btnComments, 'click', (ev) => { L.DomEvent.stopPropagation(ev); emit('view-comments', place.id as number); });
    if (btnFavorite) L.DomEvent.on(btnFavorite, 'click', (ev) => { L.DomEvent.stopPropagation(ev); emit('toggle-favorite', place.id as number); });
}

onMounted(() => {
  if (mapContainer.value) {
    map.value = L.map(mapContainer.value, { zoomControl: false }).setView([39.0, 35.0], 6);
    const cartoLight = L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', { attribution: '&copy; CARTO' });
    const cartoDark = L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', { attribution: '&copy; CARTO' });
    
    if (isDarkMode.value) cartoDark.addTo(map.value);
    else cartoLight.addTo(map.value);

    watch(isDarkMode, (newVal) => {
        if (map.value) {
            if (newVal) { map.value.removeLayer(cartoLight); map.value.addLayer(cartoDark); }
            else { map.value.removeLayer(cartoDark); map.value.addLayer(cartoLight); }
        }
    });

    watch(locale, () => updateMarkers());

    L.control.zoom({ position: 'bottomright' }).addTo(map.value);

    map.value.on('popupopen', (e: any) => {
        const popupNode = e.popup._contentNode;
        const placeId = (e.popup._source as any).placeId;
        if (placeId) {
            const place = props.places.find(p => p.id === placeId);
            if (place && popupNode) setTimeout(() => attachPopupEvents(popupNode, place), 50);
        }
    });

    // @ts-ignore
    if (L.markerClusterGroup) {
        // @ts-ignore
        markerClusterGroup.value = L.markerClusterGroup({
            showCoverageOnHover: false,
            zoomToBoundsOnClick: true,
            iconCreateFunction: (cluster: any) => L.divIcon({ html: `<div class="custom-cluster"><span>${cluster.getChildCount()}</span></div>`, className: 'cluster-wrapper', iconSize: L.point(40, 40) })
        });
        map.value.addLayer(markerClusterGroup.value);
    }
    updateMarkers();
  }
});

watch(() => props.places, () => updateMarkers(), { deep: true });
watch(() => props.selectedPlaceId, (newId, oldId) => {
  if (map.value && markerClusterGroup.value) {
    if (oldId) {
        const oldMarker = markersMap.value.get(oldId);
        const oldPlace = props.places.find(p => p.id === oldId);
        if (oldMarker && oldPlace) oldMarker.setIcon(createCustomIcon(oldPlace.category, false));
    }
    if (newId) {
      const place = props.places.find(p => p.id === newId);
      const marker = markersMap.value.get(newId);
      if (place && marker) {
        marker.setIcon(createCustomIcon(place.category, true));
        map.value.flyTo([place.lat, place.lng], 15, { duration: 1.5 });
        markerClusterGroup.value.zoomToShowLayer(marker, () => marker.openPopup());
      }
    }
  }
});

defineExpose({ setRoute });
</script>

<template>
  <div class="relative w-full h-full flex-grow">
    <div ref="mapContainer" class="w-full h-full z-0"></div>
    <button class="absolute bottom-5 right-[60px] z-[999] w-11 h-11 bg-emerald-500 border-none rounded-full text-2xl cursor-pointer shadow-lg flex items-center justify-center transition-all hover:scale-110 hover:bg-emerald-600 text-white" @click="locateUser" :title="t('map.locate_me')">📍</button>
    <button class="absolute bottom-5 right-[115px] z-[999] w-11 h-11 bg-red-600 border-2 border-white rounded-full text-lg font-bold text-white shadow-xl cursor-pointer flex items-center justify-center transition-all hover:scale-110 hover:bg-red-700 animate-pulse" @click="handleEmergency" title="Acil Durum (Hastane/Eczane)">SOS</button>

    <RoutePanel 
        :route-waypoints="routeWaypoints"
        :transport-mode="transportMode"
        :route-info="routeInfo"
        @clear-route="clearRoute"
        @update-mode="(m) => transportMode = m"
        @remove-waypoint="removeWaypoint"
        @move-waypoint="moveWaypoint"
        @share-route="shareRoute"
        @open-transit="openGoogleMapsTransit"
        @close-info="routeInfo = null"
    />
  </div>
</template>
