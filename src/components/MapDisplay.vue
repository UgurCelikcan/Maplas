<script setup lang="ts">
import { onMounted, ref, watch, shallowRef, inject } from 'vue';
import { useI18n } from 'vue-i18n';
import L from 'leaflet';
import { getLocalizedContent } from '../utils';
import { translateText } from '../api';

const { t, locale } = useI18n();

interface Place {
  id?: number;
  name: Record<string, string>;
  description: Record<string, string>;
  lat: number;
  lng: number;
  category: string;
  city: string;
  imageUrl?: string;
  is_favorite?: boolean;
}

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

const mapContainer = ref<HTMLElement | null>(null);
const map = shallowRef<L.Map | null>(null);
const markerClusterGroup = shallowRef<any>(null);
const markersMap = shallowRef<Map<number, L.Marker>>(new Map());
const routingControl = shallowRef<any>(null);
const userLocation = ref<{ lat: number; lng: number } | null>(null);
const userMarker = shallowRef<L.Marker | null>(null);
const routeInfo = ref<{ roads: string[], totalDistance: string, totalTime: string } | null>(null);
const routeWaypoints = ref<Array<{lat: number, lng: number, name: string}>>([]);
const transportMode = ref('driving');

const transportOptions = [
    { id: 'driving', icon: '🚗', label: 'Araba' },
    { id: 'bicycle', icon: '🚲', label: 'Bisiklet' },
    { id: 'foot', icon: '🚶', label: 'Yürüyüş' }
];

watch(transportMode, () => {
    if (routeWaypoints.value.length >= 2) {
        updateRouteDisplay();
    }
});

const getCategoryColor = (category: string) => {
  switch (category) {
    case 'Tarihi': return '#f87171';
    case 'Doğa': return '#4ade80';
    case 'Plaj': return '#60a5fa';
    case 'Müze': return '#c084fc';
    case 'Antik Kent': return '#fbbf24';
    case 'Alışveriş': return '#e91e63';
    default: return '#94a3b8';
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
        
        // Update user location in route if exists, or just store it
        const startIndex = routeWaypoints.value.findIndex(wp => wp.name === t('map.your_location'));
        if (startIndex !== -1) {
            routeWaypoints.value[startIndex] = { lat: e.latlng.lat, lng: e.latlng.lng, name: t('map.your_location') };
            updateRouteDisplay();
        } else if (routeWaypoints.value.length > 0) {
            // User requested location AFTER route was created -> Add to START
            routeWaypoints.value.unshift({ lat: e.latlng.lat, lng: e.latlng.lng, name: t('map.your_location') });
            updateRouteDisplay();
        }
    });

    map.value.once('locationerror', () => {
        alert(t('map.location_not_found'));
    });
}

function addToRoute(lat: number, lng: number, name: string) {
    // If first point and we know user location, add user location as start
    if (routeWaypoints.value.length === 0 && userLocation.value) {
        routeWaypoints.value.push({
            lat: userLocation.value.lat,
            lng: userLocation.value.lng,
            name: t('map.your_location')
        });
    }

    if (routeWaypoints.value.some(wp => wp.lat === lat && wp.lng === lng)) {
        return;
    }
    
    routeWaypoints.value.push({ lat, lng, name });
    updateRouteDisplay();
}

function removeWaypoint(index: number) {
    routeWaypoints.value.splice(index, 1);
    updateRouteDisplay();
}

function moveWaypoint(index: number, direction: 'up' | 'down') {
    if (direction === 'up' && index > 0) {
        const temp = routeWaypoints.value[index];
        // @ts-ignore
        routeWaypoints.value[index] = routeWaypoints.value[index - 1];
        // @ts-ignore
        routeWaypoints.value[index - 1] = temp;
    } else if (direction === 'down' && index < routeWaypoints.value.length - 1) {
        const temp = routeWaypoints.value[index];
        // @ts-ignore
        routeWaypoints.value[index] = routeWaypoints.value[index + 1];
        // @ts-ignore
        routeWaypoints.value[index + 1] = temp;
    }
    updateRouteDisplay();
}

function clearRoute() {
    routeWaypoints.value = [];
    if (routingControl.value) {
        map.value?.removeControl(routingControl.value);
        routingControl.value = null;
    }
    routeInfo.value = null;
}

function updateRouteDisplay() {
    if (routeWaypoints.value.length < 2) {
        if (routingControl.value) {
            map.value?.removeControl(routingControl.value);
            routingControl.value = null;
            routeInfo.value = null;
        }
        return;
    }

    if (routingControl.value) {
        map.value?.removeControl(routingControl.value);
    }

    // @ts-ignore
    if (L.Routing) {
        let profile = 'car';
        if (transportMode.value === 'bicycle') profile = 'bike';
        if (transportMode.value === 'foot') profile = 'foot';

        // @ts-ignore
        const control = L.Routing.control({
            waypoints: routeWaypoints.value.map(wp => L.latLng(wp.lat, wp.lng)),
            router: L.Routing.osrmv1({
                serviceUrl: 'https://router.project-osrm.org/route/v1',
                profile: profile
            }),
            routeWhileDragging: false,
            addWaypoints: false,
            fitSelectedRoutes: true,
            showAlternatives: false,
            show: false,
            lineOptions: {
                styles: [{ 
                    color: transportMode.value === 'bicycle' ? '#3b82f6' : (transportMode.value === 'foot' ? '#f97316' : '#42b883'), 
                    opacity: 0.8, 
                    weight: 6 
                }],
                extendToWaypoints: true,
                missingRouteTolerance: 0
            },
            // @ts-ignore
            createMarker: () => null
        });

        control.addTo(map.value!);
        routingControl.value = control;

        control.on('routesfound', function(e: any) {
            const routes = e.routes;
            const summary = routes[0].summary;
            const distanceMeters = summary.totalDistance;
            
            // Manual Speed Calculation (since OSRM demo server might ignore profile)
            // Car: ~50 km/h = ~833 m/min
            // Bike: ~15 km/h = ~250 m/min
            // Foot: ~5 km/h = ~83 m/min
            let speedMetersPerMin = 833; 
            if (transportMode.value === 'bicycle') speedMetersPerMin = 250;
            if (transportMode.value === 'foot') speedMetersPerMin = 83;

            // Use OSRM time for car (it's accurate with traffic/roads), but manual for others
            let totalMinutes = Math.round(summary.totalTime / 60);
            
            if (transportMode.value !== 'driving') {
                totalMinutes = Math.round(distanceMeters / speedMetersPerMin);
            }

            const hours = Math.floor(totalMinutes / 60);
            const mins = totalMinutes % 60;
            const timeStr = hours > 0 ? `${hours} ${t('map.hours')} ${mins} ${t('map.minutes')}` : `${mins} ${t('map.minutes')}`;
            
            const modeIcon = transportOptions.find(t => t.id === transportMode.value)?.icon || '🚗';
            const displayTime = `${modeIcon} ${timeStr}`;

            const distKm = (distanceMeters / 1000).toFixed(1);

            const instructions = routes[0].instructions;
            const roads: string[] = instructions
                .map((i: any) => i.road)
                .filter((r: string) => {
                    if (!r || r.trim().length === 0 || r === '{road}') return false;
                    const name = r.toLocaleUpperCase('tr-TR');
                    const isHighway = name.match(/^(O|D|E)\s?-?\s?\d+/) || name.includes('OTOYOL') || name.includes('ÇEVRE YOLU');
                    return isHighway;
                })
                .reduce((acc: string[], curr: string) => {
                    if (acc.length === 0 || acc[acc.length - 1] !== curr) {
                        acc.push(curr);
                    }
                    return acc;
                }, []);

            routeInfo.value = {
                roads: roads.length > 0 ? roads : [t('map.main_roads')],
                totalDistance: `${distKm} km`,
                totalTime: displayTime
            };
        });
    }
}

onMounted(() => {
  if (mapContainer.value) {
    map.value = L.map(mapContainer.value, {
        zoomControl: false 
    }).setView([39.0, 35.0], 6);

    const cartoLight = L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
        attribution: '&copy; OpenStreetMap &copy; CARTO'
    });
    const cartoDark = L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
        attribution: '&copy; OpenStreetMap &copy; CARTO'
    });
    const osm = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; OpenStreetMap'
    });
    const esriSatellite = L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
        attribution: '&copy; Esri'
    });

    if (isDarkMode.value) {
        cartoDark.addTo(map.value);
    } else {
        cartoLight.addTo(map.value);
    }

    watch(isDarkMode, (newVal) => {
        if (map.value) {
            if (newVal) {
                map.value.removeLayer(cartoLight);
                map.value.addLayer(cartoDark);
            } else {
                map.value.removeLayer(cartoDark);
                map.value.addLayer(cartoLight);
            }
        }
    });

    // Re-render markers (and thus popups) when locale changes to update text
    watch(locale, () => {
        updateMarkers();
    });

    // Overlay Layers (Optional Layers)
    const googleTraffic = L.tileLayer('https://mt0.google.com/vt?lyrs=h@159000000,traffic|seconds_into_week:-1&style=3&x={x}&y={y}&z={z}', {
        attribution: '&copy; Google'
    });

    // Note: Replace {API_KEY} with your actual OpenWeatherMap API Key
    const OWM_API_KEY = '9de243494c0b295cca9337e1e96b00e2'; // Demo Key (May not work or have limits)
    
    const rainLayer = L.tileLayer(`https://tile.openweathermap.org/map/precipitation_new/{z}/{x}/{y}.png?appid=${OWM_API_KEY}`, {
        attribution: '&copy; OpenWeatherMap'
    });

    const cloudLayer = L.tileLayer(`https://tile.openweathermap.org/map/clouds_new/{z}/{x}/{y}.png?appid=${OWM_API_KEY}`, {
        attribution: '&copy; OpenWeatherMap'
    });

    const baseMaps = {
        "🌍 Sokak (Açık)": cartoLight,
        "🌑 Karanlık": cartoDark,
        "🗺️ Standart (OSM)": osm,
        "🛰️ Uydu": esriSatellite
    };

    const overlayMaps = {
        "🚦 Trafik": googleTraffic,
        "🌧️ Yağış": rainLayer,
        "☁️ Bulutlar": cloudLayer
    };

    L.control.layers(baseMaps, overlayMaps, { position: 'bottomleft' }).addTo(map.value);
    L.control.zoom({ position: 'bottomright' }).addTo(map.value);

    map.value.on('popupopen', (e: any) => {
        const popupNode = e.popup._contentNode;
        const marker = e.popup._source;
        // Access placeId attached to marker (we need to ensure we attach it in updateMarkers)
        const placeId = (marker as any).placeId;
        
        if (placeId) {
            const place = props.places.find(p => p.id === placeId);
            if (place && popupNode) {
                 // Use timeout to ensure DOM is ready
                 setTimeout(() => attachPopupEvents(popupNode, place), 50);
            }
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
        map.value.flyTo([place.lat, place.lng], 15, { duration: 1.5 });
        markerClusterGroup.value.zoomToShowLayer(marker, () => {
            marker.openPopup();
        });
      }
    }
  }
});

function generatePopupContent(place: Place) {
    const placeName = getLocalizedContent(place.name, locale.value);
    const placeDesc = getLocalizedContent(place.description, locale.value);

    const imageHtml = place.imageUrl 
        ? `<div class="w-[calc(100%+40px)] -mx-5 -mt-5 mb-3 h-32 rounded-t-xl overflow-hidden"><img src="${place.imageUrl}" alt="${placeName}" class="w-full h-full object-cover" /></div>` 
        : '';

    return `
        <div class="custom-popup">
            ${imageHtml}
            <div class="popup-header">
                <span class="popup-category">${getCategoryEmoji(place.category)} ${t(`categories.${place.category}`)}</span>
                <button class="btn-favorite bg-transparent border-none cursor-pointer p-1 transition-transform active:scale-75" data-id="${place.id}">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="${place.is_favorite ? '#ef4444' : 'none'}" stroke="${place.is_favorite ? '#ef4444' : 'currentColor'}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
                    </svg>
                </button>
            </div>
            <div class="flex justify-between items-center mb-1">
                <span class="popup-city text-[10px] text-slate-400">📍 ${place.city}</span>
            </div>
            <h3>${placeName}</h3>
            <p>${placeDesc}</p>
            <div class="weather-info mt-2 text-xs text-slate-500 flex items-center gap-1">
                <span class="weather-loading">🌤️ Hava durumu yükleniyor...</span>
            </div>
            <button class="btn-add-route w-full mt-3 bg-emerald-500 hover:bg-emerald-600 text-white border-none py-2.5 px-3 rounded-lg font-bold cursor-pointer transition-colors flex items-center justify-center gap-2" data-lat="${place.lat}" data-lng="${place.lng}" data-name="${placeName}">
                🚩 Rotaya Ekle
            </button>
            <button class="btn-comments w-full mt-2 bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 text-slate-800 dark:text-white border-none py-2 px-3 rounded-lg font-medium cursor-pointer transition-colors flex items-center justify-center gap-2" data-id="${place.id}">
                💬 ${t('map.view_comments')}
            </button>
        </div>
    `;
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
                    if (translated && translated !== originalText) {
                         descContainer.innerHTML = `${translated} <span class="text-[9px] text-emerald-500 font-bold ml-1 bg-emerald-50 px-1 rounded border border-emerald-100 cursor-help" title="Otomatik Çevrildi">🌐</span>`;
                    }
                }).catch(() => {});
            }

            if (btnAddRoute) {
                // Weather Logic (Only if weatherContainer exists)
                if (weatherContainer) {
                     const lat = place.lat;
                     const lng = place.lng;
                     
                     fetch(`https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lng}&current_weather=true&daily=weathercode,temperature_2m_max,temperature_2m_min&timezone=auto`)
                        .then(res => res.json())
                        .then(data => {
                            const weather = data.current_weather;
                            const temp = Math.round(weather.temperature);
                            const wind = weather.windspeed;
                            const code = weather.weathercode;
                            
                            const getIcon = (c: number) => {
                                if (c === 0) return '☀️';
                                if (c === 1 || c === 2 || c === 3) return '⛅';
                                if (c === 45 || c === 48) return '🌫️';
                                if (c >= 51 && c <= 67) return '🌧️';
                                if (c >= 71 && c <= 77) return '❄️';
                                if (c >= 80 && c <= 82) return '🌦️';
                                if (c >= 95) return '⛈️';
                                return '🌡️';
                            };

                            const currentIcon = getIcon(code);

                            let forecastHtml = '';
                            if (data.daily && data.daily.time) {
                                forecastHtml = '<div class="mt-2 pt-2 border-t border-slate-100 dark:border-zinc-700 grid grid-cols-3 gap-1">';
                                for (let i = 1; i <= 3; i++) {
                                    if (data.daily.time[i]) {
                                        const date = new Date(data.daily.time[i]);
                                        const dayName = date.toLocaleDateString(locale.value, { weekday: 'short' });
                                        const maxTemp = Math.round(data.daily.temperature_2m_max[i]);
                                        const dailyIcon = getIcon(data.daily.weathercode[i]);
                                        
                                        forecastHtml += `
                                            <div class="text-center">
                                                <div class="text-[9px] text-slate-400 font-bold uppercase">${dayName}</div>
                                                <div class="text-lg leading-none my-0.5">${dailyIcon}</div>
                                                <div class="text-[10px] font-bold text-slate-700 dark:text-zinc-300">${maxTemp}°</div>
                                            </div>
                                        `;
                                    }
                                }
                                forecastHtml += '</div>';
                            }

                            weatherContainer.innerHTML = `
                                <div class="flex items-center gap-2 mb-1">
                                    <span class="text-2xl">${currentIcon}</span>
                                    <div>
                                        <span class="font-bold text-slate-800 dark:text-white text-sm">${temp}°C</span>
                                        <span class="text-[10px] text-slate-500 block">💨 ${wind} km/s</span>
                                    </div>
                                </div>
                                ${forecastHtml}
                            `;
                        })
                        .catch(err => {
                            console.error('Weather error:', err);
                            weatherContainer.innerHTML = '<span class="text-red-400 text-[10px]">Hava durumu yok</span>';
                        });
                }
                
                // Route Button
                // @ts-ignore
                L.DomEvent.off(btnAddRoute, 'click');
                L.DomEvent.on(btnAddRoute, 'click', (ev: any) => {
                    L.DomEvent.stopPropagation(ev);
                    L.DomEvent.preventDefault(ev);
                    const name = getLocalizedContent(place.name, locale.value);
                    addToRoute(place.lat, place.lng, name);
                });
            }
            
            if (btnComments) {
                // @ts-ignore
                L.DomEvent.off(btnComments, 'click');
                L.DomEvent.on(btnComments, 'click', (ev: any) => {
                    L.DomEvent.stopPropagation(ev);
                    L.DomEvent.preventDefault(ev);
                    emit('view-comments', place.id as number);
                });
            }

            if (btnFavorite) {
                // @ts-ignore
                L.DomEvent.off(btnFavorite, 'click');
                L.DomEvent.on(btnFavorite, 'click', (ev: any) => {
                    L.DomEvent.stopPropagation(ev);
                    L.DomEvent.preventDefault(ev);
                    emit('toggle-favorite', place.id as number);
                });
            }
}

function updateMarkers() {
  if (!map.value) return;

  const newPlaceIds = new Set(props.places.map(p => p.id));
  
  // 1. Remove markers that are no longer in the list
  markersMap.value.forEach((marker, id) => {
    if (!newPlaceIds.has(id)) {
        if (markerClusterGroup.value) {
            markerClusterGroup.value.removeLayer(marker);
        } else {
            marker.remove();
        }
        markersMap.value.delete(id);
    }
  });

  // 2. Add new markers or update existing ones
  props.places.forEach(place => {
    const existingMarker = markersMap.value.get(place.id as number);
    const isSelected = props.selectedPlaceId === place.id;
    const customIcon = createCustomIcon(place.category, isSelected);
    const popupContent = generatePopupContent(place);

    if (existingMarker) {
        // Update existing marker
        existingMarker.setIcon(customIcon);
        
        // Update popup content if changed
        const popup = existingMarker.getPopup();
        if (popup) {
             // Only update content if it's different to avoid flicker
             if (popup.getContent() !== popupContent) {
                 popup.setContent(popupContent);
                 
                 // If popup is currently open, we must re-attach events because the DOM node is replaced
                 if (popup.isOpen()) {
                     // Wait for the content to be rendered in DOM
                     setTimeout(() => {
                        // @ts-ignore
                        if (popup._contentNode) {
                            // @ts-ignore
                            attachPopupEvents(popup._contentNode, place);
                        }
                     }, 50);
                 }
             }
        }
    } else {
        // Create new marker
        const marker = L.marker([place.lat, place.lng], {
          icon: customIcon
        });
        (marker as any).placeId = place.id;

        marker.bindPopup(popupContent, {
            closeButton: false,
            className: 'modern-popup'
        });
        
        marker.on('click', (e) => {
            L.DomEvent.stopPropagation(e);
            emit('select-place', place.id as number);
        });

        if (markerClusterGroup.value) {
            markerClusterGroup.value.addLayer(marker);
        } else {
            marker.addTo(map.value!);
        }
        
        markersMap.value.set(place.id as number, marker);
    }
  });
}

function setRoute(places: Place[]) {
    clearRoute();
    places.forEach(p => {
        addToRoute(p.lat, p.lng, getLocalizedContent(p.name, locale.value));
    });
    // Add user location as start if available
    if (userLocation.value) {
        routeWaypoints.value.unshift({
            lat: userLocation.value.lat,
            lng: userLocation.value.lng,
            name: t('map.your_location')
        });
        updateRouteDisplay();
    }
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

        // Add user marker
        if (map.value) {
             map.value.setView([latitude, longitude], 14);
             if (!userMarker.value) {
                 userMarker.value = L.marker([latitude, longitude], {
                    icon: L.divIcon({ className: 'user-marker', html: '<div class="user-dot"></div>', iconSize: [20, 20] })
                 }).addTo(map.value);
             } else {
                 userMarker.value.setLatLng([latitude, longitude]);
             }
        }

        try {
            // Find nearest hospital or clinic using Overpass API
            const query = `
                [out:json];
                (
                  node["amenity"~"hospital|clinic|doctors"](around:5000,${latitude},${longitude});
                  way["amenity"~"hospital|clinic|doctors"](around:5000,${latitude},${longitude});
                  relation["amenity"~"hospital|clinic|doctors"](around:5000,${latitude},${longitude});
                );
                out center;
            `;
            const url = `https://overpass-api.de/api/interpreter?data=${encodeURIComponent(query)}`;
            const res = await fetch(url);
            const data = await res.json();

            if (data.elements && data.elements.length > 0) {
                // Find closest (simple Euclidean approximation is enough for short distance)
                let closest: any = null;
                let minDist = Infinity;

                data.elements.forEach((el: any) => {
                    const elLat = el.lat || el.center.lat;
                    const elLng = el.lon || el.center.lon;
                    const dist = Math.sqrt(Math.pow(elLat - latitude, 2) + Math.pow(elLng - longitude, 2));
                    if (dist < minDist) {
                        minDist = dist;
                        closest = { ...el, lat: elLat, lng: elLng };
                    }
                });

                if (closest) {
                    // Create Route
                    clearRoute();
                    const name = closest.tags.name || (closest.tags.amenity === 'hospital' ? 'Hastane' : 'Sağlık Merkezi');
                    const icon = closest.tags.amenity === 'hospital' ? '🏥' : '🩺';
                    
                    routeWaypoints.value = [
                        { lat: latitude, lng: longitude, name: '🚨 Konumum' },
                        { lat: closest.lat, lng: closest.lng, name: `${icon} ${name}` }
                    ];
                    updateRouteDisplay();
                    alert(`🚨 En yakın sağlık kuruluşu bulundu: ${name}\nRota oluşturuldu!`);
                }
            } else {
                alert("Yakınlarda (5km) sağlık kuruluşu bulunamadı. Lütfen 112'yi arayın!");
            }
        } catch (e) {
            console.error(e);
            alert("Hata oluştu. Lütfen manuel arama yapın.");
        }

    }, () => {
        alert("Konum alınamadı.");
    });
}

function openGoogleMapsTransit() {
    if (routeWaypoints.value.length < 2) return;
    
    const start = routeWaypoints.value[0];
    const end = routeWaypoints.value[routeWaypoints.value.length - 1];
    
    if (!start || !end) return;

    const url = `https://www.google.com/maps/dir/?api=1&origin=${start.lat},${start.lng}&destination=${end.lat},${end.lng}&travelmode=transit`;
    window.open(url, '_blank');
}

defineExpose({ setRoute });
</script>

<template>
  <div class="relative w-full h-full flex-grow">
    <div ref="mapContainer" class="w-full h-full z-0"></div>
    <button class="absolute bottom-5 right-[60px] z-[999] w-11 h-11 bg-emerald-500 border-none rounded-full text-2xl cursor-pointer shadow-lg flex items-center justify-center transition-all hover:scale-110 hover:bg-emerald-600 text-white" @click="locateUser" :title="t('map.locate_me')">
        📍
    </button>
    
    <!-- SOS Button -->
    <button class="absolute bottom-5 right-[115px] z-[999] w-11 h-11 bg-red-600 border-2 border-white rounded-full text-lg font-bold text-white shadow-xl cursor-pointer flex items-center justify-center transition-all hover:scale-110 hover:bg-red-700 animate-pulse" @click="handleEmergency" title="Acil Durum (Hastane/Eczane)">
        SOS
    </button>

    <!-- Route Panels Container -->
    <div class="absolute top-16 md:top-4 left-4 z-[999] flex flex-col gap-3 w-[280px]">
        
        <!-- Planner Panel -->
        <div v-if="routeWaypoints.length > 0" class="bg-white dark:bg-zinc-800 p-4 shadow-xl border border-slate-200 dark:border-zinc-700 rounded-xl animate-in fade-in slide-in-from-left-4">
            <div class="flex justify-between items-center mb-3 pb-2 border-b border-slate-100 dark:border-zinc-700">
                <span class="font-bold text-slate-900 dark:text-white flex items-center gap-2">🚦 Rota Planlayıcı <span class="bg-emerald-500 text-white text-[10px] px-1.5 py-0.5 rounded-full">{{ routeWaypoints.length }}</span></span>
                <button class="text-xs text-red-500 hover:text-red-600 bg-transparent border-none cursor-pointer" @click="clearRoute">Temizle</button>
            </div>
            
            <!-- Transport Modes -->
            <div class="flex gap-1 bg-slate-100 dark:bg-zinc-700/50 p-1 rounded-lg mb-3">
                <button 
                    v-for="mode in transportOptions" 
                    :key="mode.id"
                    @click="transportMode = mode.id"
                    class="flex-1 flex items-center justify-center gap-1 py-1.5 rounded-md text-xs font-bold transition-all cursor-pointer border-none"
                    :class="transportMode === mode.id ? 'bg-white dark:bg-zinc-600 text-slate-900 dark:text-white shadow-sm' : 'bg-transparent text-slate-400 hover:text-slate-600 dark:hover:text-zinc-300'"
                    :title="mode.label"
                >
                    <span class="text-sm">{{ mode.icon }}</span>
                </button>
            </div>
            
            <ul class="m-0 p-0 list-none space-y-2 max-h-[250px] overflow-y-auto pr-1">
                <li v-for="(wp, index) in routeWaypoints" :key="index" class="flex items-center gap-2 bg-slate-50 dark:bg-zinc-700/50 p-2 rounded-lg text-sm group">
                    <span class="w-5 h-5 flex items-center justify-center bg-emerald-500 text-white rounded-full text-xs font-bold flex-shrink-0">{{ index + 1 }}</span>
                    <span class="truncate flex-grow text-slate-700 dark:text-zinc-200">{{ wp.name }}</span>
                    
                    <div class="flex flex-col gap-0.5 opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity">
                        <button class="text-[10px] leading-none text-slate-400 hover:text-emerald-500 bg-transparent border-none cursor-pointer p-0" @click.stop="moveWaypoint(index, 'up')" v-if="index > 0">▲</button>
                        <button class="text-[10px] leading-none text-slate-400 hover:text-emerald-500 bg-transparent border-none cursor-pointer p-0" @click.stop="moveWaypoint(index, 'down')" v-if="index < routeWaypoints.length - 1">▼</button>
                    </div>

                    <button class="text-slate-400 hover:text-red-500 bg-transparent border-none cursor-pointer opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity ml-1" @click.stop="removeWaypoint(index)">✕</button>
                </li>
            </ul>

            <div v-if="routeWaypoints.length < 2" class="mt-3 text-[11px] text-slate-400 text-center italic">
                En az bir durak daha ekleyerek rotayı görün.
            </div>
        </div>

        <!-- Info Panel -->
        <div v-if="routeInfo" class="bg-white dark:bg-zinc-800 p-4 rounded-xl shadow-xl border border-slate-200 dark:border-zinc-700 animate-in fade-in slide-in-from-top-2">
            <div class="flex justify-between items-center mb-3 pb-2 border-b border-slate-100 dark:border-zinc-700">
                <div>
                    <span class="text-2xl font-bold text-slate-900 dark:text-white">{{ routeInfo.totalTime }}</span>
                    <span class="text-sm text-slate-500 dark:text-zinc-400 ml-2">({{ routeInfo.totalDistance }})</span>
                </div>
                <div class="flex gap-2">
                    <button class="bg-blue-500 hover:bg-blue-600 text-white border-none rounded p-1.5 cursor-pointer text-xs flex items-center gap-1 transition-colors" @click="openGoogleMapsTransit" title="Google Haritalar'da Toplu Taşıma ile Aç">
                        🚌 <span class="hidden sm:inline">Toplu Taşıma</span>
                    </button>
                    <button class="bg-transparent border-none text-slate-400 hover:text-slate-600 cursor-pointer text-lg" @click="routeInfo = null">✕</button>
                </div>
            </div>
            
            <div class="max-h-[150px] overflow-y-auto space-y-1 pr-1">
                <div v-for="(road, index) in routeInfo.roads" :key="index" class="flex items-center gap-2 text-sm text-slate-700 dark:text-zinc-300">
                    <span class="w-1.5 h-1.5 bg-emerald-500 rounded-full flex-shrink-0"></span>
                    <span class="truncate">{{ road }}</span>
                </div>
            </div>
        </div>

    </div>
  </div>
</template>