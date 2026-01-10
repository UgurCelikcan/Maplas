<script setup lang="ts">
import { ref, onMounted } from 'vue';
import L from 'leaflet';

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

const form = ref<Place>({
  name: '',
  description: '',
  lat: 0,
  lng: 0,
  category: '',
  city: '',
  imageUrl: ''
});

const mapContainer = ref<HTMLElement | null>(null);
const map = ref<L.Map | null>(null);
const marker = ref<L.Marker | null>(null);
const searchQuery = ref('');
const isSearching = ref(false);

onMounted(() => {
  if (props.initialData) {
    form.value = { ...props.initialData };
  }

  initMap();
});

function initMap() {
    if (mapContainer.value) {
        // Default center (Turkey)
        const startLat = form.value.lat || 39.0;
        const startLng = form.value.lng || 35.0;
        const startZoom = form.value.lat ? 13 : 5;

        map.value = L.map(mapContainer.value).setView([startLat, startLng], startZoom);
        
        L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
            attribution: '&copy; OpenStreetMap'
        }).addTo(map.value);

        // Click handler to pick location
        map.value.on('click', (e: L.LeafletMouseEvent) => {
            updateLocation(e.latlng.lat, e.latlng.lng);
        });

        // Set initial marker if editing
        if (form.value.lat && form.value.lng) {
            updateLocation(form.value.lat, form.value.lng);
        }

        // Force map resize after modal animation
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
            
            // Auto fill city if empty and available in address (simplified)
            // Note: Nominatim display_name is complex, this is just a helper logic
            if (!form.value.city && result.display_name) {
               const parts = result.display_name.split(',');
               // Usually city is near the end or middle, hard to guess perfectly without 'addressdetails=1'
               // Let's just update the view for now.
            }
        } else {
            alert('Konum bulunamadı.');
        }
    } catch (error) {
        console.error('Search error:', error);
        alert('Arama sırasında hata oluştu.');
    } finally {
        isSearching.value = false;
    }
}

function handleSubmit() {
    if (form.value.lat === 0 && form.value.lng === 0) {
        alert('Lütfen haritadan bir konum seçin.');
        return;
    }
    emit('save-place', form.value);
}
</script>

<template>
  <div class="fixed inset-0 w-full h-full bg-black/60 flex justify-center items-center z-[2000] backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-800 p-6 rounded-2xl w-[90%] max-w-[500px] max-h-[90vh] overflow-y-auto shadow-2xl border border-slate-200 dark:border-zinc-700 text-slate-800 dark:text-white transition-colors duration-300">
      <div class="flex justify-between items-center mb-5">
        <h3 class="m-0 text-xl font-semibold">{{ initialData ? 'Yeri Düzenle' : 'Yeni Yer Ekle' }}</h3>
        <button class="bg-transparent border-none text-slate-400 hover:text-slate-600 dark:text-zinc-400 dark:hover:text-zinc-200 cursor-pointer text-xl" @click="$emit('close')">✕</button>
      </div>
      
      <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
        <!-- Location Picker Section (Moved to top for importance) -->
        <div class="flex flex-col gap-2">
            <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">Konum Seç (Haritaya Tıklayın)</label>
            <div class="flex gap-2">
                <input v-model="searchQuery" @keydown.enter.prevent="searchLocation" placeholder="Adres veya yer ara (örn: Ayasofya)..." 
                       class="flex-grow p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
                <button type="button" @click="searchLocation" 
                        class="px-4 py-2 bg-emerald-100 dark:bg-emerald-900 text-emerald-700 dark:text-emerald-300 rounded-lg font-medium border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-200 dark:hover:bg-emerald-800 transition-colors disabled:opacity-50">
                    {{ isSearching ? '...' : 'Bul' }}
                </button>
            </div>
            
            <div class="relative w-full h-64 rounded-lg overflow-hidden border border-slate-300 dark:border-zinc-700 z-0">
                <div ref="mapContainer" class="w-full h-full z-0"></div>
                <div v-if="!form.lat" class="absolute inset-0 bg-black/10 flex items-center justify-center pointer-events-none z-[400]">
                    <span class="bg-white/80 dark:bg-black/60 px-3 py-1 rounded text-xs font-semibold backdrop-blur-sm">Konum İşaretleyin</span>
                </div>
            </div>
            <div class="flex justify-between text-xs text-slate-400">
                <span>{{ form.lat ? `${form.lat}, ${form.lng}` : 'Konum seçilmedi' }}</span>
                <span v-if="form.lat" class="text-emerald-500">✓ Seçildi</span>
            </div>
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">Yer Adı</label>
          <input v-model="form.name" required placeholder="Örn: Galata Kulesi" 
                 class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
        </div>

        <div class="flex gap-3">
            <div class="flex flex-col gap-1.5 flex-1">
                <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">Şehir</label>
                <input v-model="form.city" required placeholder="Örn: İstanbul" 
                       class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
            </div>
            <div class="flex flex-col gap-1.5 flex-1">
                <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">Kategori</label>
                <select v-model="form.category" required 
                        class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors">
                    <option value="" disabled>Seçiniz</option>
                    <option value="Tarihi">Tarihi</option>
                    <option value="Doğa">Doğa</option>
                    <option value="Plaj">Plaj</option>
                    <option value="Müze">Müze</option>
                    <option value="Antik Kent">Antik Kent</option>
                    <option value="Alışveriş">Alışveriş</option>
                    <option value="Manzara">Manzara</option>
                    <option value="Diğer">Diğer</option>
                </select>
            </div>
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">Resim URL (İsteğe bağlı)</label>
          <input v-model="form.imageUrl" placeholder="https://example.com/image.jpg" 
                 class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
        </div>

        <div class="flex flex-col gap-1.5">
          <label class="text-sm font-medium text-slate-500 dark:text-zinc-400">Açıklama</label>
          <textarea v-model="form.description" rows="3" required placeholder="Kısa bir açıklama..." 
                    class="p-2.5 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white text-sm focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors font-sans"></textarea>
        </div>

        <div class="flex justify-end gap-2.5 mt-2">
          <button type="button" class="px-4 py-2 rounded-lg border border-slate-300 dark:border-zinc-700 bg-transparent text-slate-600 dark:text-zinc-300 cursor-pointer hover:bg-slate-100 dark:hover:bg-zinc-700 transition-colors" @click="$emit('close')">İptal</button>
          <button type="submit" class="px-5 py-2 rounded-lg border-none bg-emerald-500 text-white font-semibold cursor-pointer hover:bg-emerald-400 transition-colors shadow-lg shadow-emerald-500/20">{{ initialData ? 'Güncelle' : 'Kaydet' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>