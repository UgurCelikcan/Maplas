<script setup lang="ts">
import { ref, onMounted, shallowRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import L from 'leaflet';
import { uploadImage } from '../api';

const { t } = useI18n();

interface Place {
  id?: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
  imageUrl?: string;
}

const props = defineProps<{
  initialData?: Place;
}>();

const emit = defineEmits<{
  (e: 'save-place', place: Place): void;
  (e: 'close'): void;
}>();

const form = ref({
  name: '',
  description: '',
  lat: 0,
  lng: 0,
  category: 'Tarihi',
  city: '',
  imageUrl: ''
});

watch(() => props.initialData, (newVal: Place | undefined) => {
  if (newVal) {
    // const isTurkish = navigator.language.startsWith('tr') || localStorage.getItem('lang') === 'tr';
    form.value = {
      ...newVal,
      name: typeof newVal.name === 'string' ? newVal.name : (newVal.name['tr'] || Object.values(newVal.name)[0]),
      description: typeof newVal.description === 'string' ? newVal.description : (newVal.description['tr'] || Object.values(newVal.description)[0])
    };
  }
}, { immediate: true });

const mapContainer = ref<HTMLElement | null>(null);
const map = shallowRef<L.Map | null>(null);
const marker = shallowRef<L.Marker | null>(null);
const searchQuery = ref('');
const isSearching = ref(false);
const isUploading = ref(false);

onMounted(() => {
  if (props.initialData) {
    form.value = { ...props.initialData };
  }

  initMap();
});

function initMap() {
    if (mapContainer.value) {
        const startLat = form.value.lat || 39.0;
        const startLng = form.value.lng || 35.0;
        const startZoom = form.value.lat ? 13 : 5;

        map.value = L.map(mapContainer.value).setView([startLat, startLng], startZoom);
        
        L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
            attribution: '&copy; OpenStreetMap'
        }).addTo(map.value);

        map.value.on('click', (e: L.LeafletMouseEvent) => {
            updateLocation(e.latlng.lat, e.latlng.lng);
        });

        if (form.value.lat && form.value.lng) {
            updateLocation(form.value.lat, form.value.lng);
        }

        setTimeout(() => {
            map.value?.invalidateSize();
        }, 200);
    }
}

function updateLocation(lat: number, lng: number, pan: boolean = false) {
    form.value.lat = parseFloat(lat.toFixed(6));
    form.value.lng = parseFloat(lng.toFixed(6));

    if (marker.value) {
        marker.value.setLatLng([lat, lng]);
    } else if (map.value) {
        marker.value = L.marker([lat, lng]).addTo(map.value);
    }

    if (pan && map.value) {
        map.value.setView([lat, lng], 13);
    }
}

async function searchLocation() {
    if (!searchQuery.value) return;
    isSearching.value = true;

    try {
        const response = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(searchQuery.value)}`);
        const data = await response.json();

        if (data && data.length > 0) {
            const result = data[0];
            const lat = parseFloat(result.lat);
            const lng = parseFloat(result.lon);
            updateLocation(lat, lng, true);
        } else {
            alert(t('map.route_not_found'));
        }
    } catch (error) {
        console.error('Search error:', error);
        alert(t('place.save_error'));
    } finally {
        isSearching.value = false;
    }
}

async function handleFileChange(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
        const file = target.files[0];
        if (!file) return;
        
        if (file.size > 10 * 1024 * 1024) {
            alert('File too large (Max 10MB)');
            return;
        }

        isUploading.value = true;
        try {
            const url = await uploadImage(file);
            form.value.imageUrl = url;
        } catch (error) {
            console.error('Upload failed:', error);
            alert(t('place.save_error'));
        } finally {
            isUploading.value = false;
        }
    }
}

function handleSubmit() {
    if (form.value.lat === 0 && form.value.lng === 0) {
        alert(t('place.select_on_map'));
        return;
    }
    emit('save-place', form.value);
}
</script>

<template>
  <div class="fixed inset-0 w-full h-full bg-black/60 flex justify-center items-center z-[2000] backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-800 p-6 rounded-2xl w-[90%] max-w-[500px] max-h-[90vh] overflow-y-auto shadow-2xl border border-slate-200 dark:border-zinc-700 text-slate-800 dark:text-white transition-colors duration-300">
      <div class="flex justify-between items-center mb-5">
        <h3 class="m-0 text-xl font-semibold">{{ initialData ? t('place.edit_title') : t('place.add_title') }}</h3>
        <button class="bg-transparent border-none text-slate-400 hover:text-slate-600 dark:text-zinc-400 dark:hover:text-zinc-200 cursor-pointer text-xl" @click="$emit('close')">✕</button>
      </div>
      
      <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
        <!-- Location Picker Section -->
        <div class="flex flex-col gap-2">
            <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">{{ t('place.select_location') }}</label>
            <div class="flex gap-2">
                <input v-model="searchQuery" @keydown.enter.prevent="searchLocation" :placeholder="t('place.search_placeholder')" 
                       class="flex-grow p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
                <button type="button" @click="searchLocation" 
                        class="px-4 py-2 bg-emerald-100 dark:bg-emerald-900 text-emerald-700 dark:text-emerald-300 rounded-lg font-medium border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-200 dark:hover:bg-emerald-800 transition-colors disabled:opacity-50">
                    {{ isSearching ? '...' : t('common.search') }}
                </button>
            </div>
            
            <div class="relative w-full h-64 rounded-lg overflow-hidden border border-slate-300 dark:border-zinc-700 z-0">
                <div ref="mapContainer" class="w-full h-full z-0"></div>
                <div v-if="!form.lat" class="absolute inset-0 bg-black/10 flex items-center justify-center pointer-events-none z-[400]">
                    <span class="bg-white/80 dark:bg-black/60 px-3 py-1 rounded text-xs font-semibold backdrop-blur-sm">{{ t('place.mark_location') }}</span>
                </div>
            </div>
            <div class="flex justify-between text-xs text-slate-400">
                <span>{{ form.lat ? `${form.lat}, ${form.lng}` : t('place.location_not_selected') }}</span>
                <span v-if="form.lat" class="text-emerald-500">✓ {{ t('place.location_selected') }}</span>
            </div>
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">{{ t('place.name') }}</label>
          <input v-model="form.name" required :placeholder="t('place.name_placeholder')" 
                 class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
        </div>

        <div class="flex gap-3">
            <div class="flex flex-col gap-1.5 flex-1">
                <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">{{ t('place.city') }}</label>
                <input v-model="form.city" required :placeholder="t('place.city_placeholder')" 
                       class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
            </div>
            <div class="flex flex-col gap-1.5 flex-1">
                <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">{{ t('place.category') }}</label>
                <select v-model="form.category" required 
                        class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors">
                    <option value="" disabled>{{ t('categories.select') }}</option>
                    <option v-for="cat in ['Tarihi', 'Doğa', 'Plaj', 'Müze', 'Antik Kent', 'Alışveriş', 'Manzara', 'Diğer']" :key="cat" :value="cat">
                        {{ t(`categories.${cat}`) }}
                    </option>
                </select>
            </div>
        </div>

        <!-- Image Upload Section -->
        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">{{ t('place.image_url') }}</label>
          
          <div class="flex items-center gap-2">
               <input type="file" accept="image/*" @change="handleFileChange" class="hidden" id="file-upload" />
               <label for="file-upload" class="cursor-pointer px-4 py-2.5 bg-slate-100 dark:bg-zinc-700 text-slate-600 dark:text-zinc-300 rounded-lg text-sm font-medium hover:bg-slate-200 dark:hover:bg-zinc-600 transition-colors flex-grow text-center border border-slate-200 dark:border-zinc-600">
                    {{ isUploading ? t('common.loading') : (form.imageUrl ? '📸 ' + t('common.update') : '📷 ' + t('ui.add_new')) }}
               </label>
          </div>

          <!-- Preview -->
          <div v-if="form.imageUrl" class="relative mt-2 w-full h-32 rounded-lg overflow-hidden border border-slate-200 dark:border-zinc-700 bg-slate-50 dark:bg-black/20 group">
              <img :src="form.imageUrl" class="w-full h-full object-cover" />
              <button type="button" @click="form.imageUrl = ''" class="absolute top-2 right-2 bg-black/50 text-white rounded-full w-6 h-6 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity hover:bg-red-500">✕</button>
          </div>
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">{{ t('place.description') }}</label>
          <textarea v-model="form.description" rows="3" required :placeholder="t('place.description_placeholder')" 
                    class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors font-sans"></textarea>
        </div>

        <div class="flex justify-end gap-2.5 mt-2">
          <button type="button" class="px-4 py-2 rounded-lg border border-slate-300 dark:border-zinc-700 bg-transparent text-slate-600 dark:text-zinc-300 cursor-pointer hover:bg-slate-100 dark:hover:bg-zinc-700 transition-colors" @click="$emit('close')">{{ t('common.cancel') }}</button>
          <button type="submit" :disabled="isUploading" class="px-5 py-2 rounded-lg border-none bg-emerald-500 text-white font-semibold cursor-pointer hover:bg-emerald-400 transition-colors shadow-lg shadow-emerald-500/20 disabled:opacity-50 disabled:cursor-not-allowed">{{ initialData ? t('common.update') : t('common.save') }}</button>
        </div>
      </form>
    </div>
  </div>
</template>
