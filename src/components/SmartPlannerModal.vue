<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { getDistance } from '../utils';

const { t } = useI18n();
const emit = defineEmits(['close', 'create-route']);

const props = defineProps<{
  places: any[];
  userLocation: { lat: number; lng: number } | null;
}>();

const step = ref(1);
const selectedDuration = ref<number | null>(null); // minutes
const selectedInterests = ref<string[]>([]);
const transportMode = ref('driving-car');
const isGenerating = ref(false);

const transportOptions = [
    { id: 'driving-car', label: 'Araç', icon: '🚗' },
    { id: 'foot-walking', label: 'Yürüyerek', icon: '🚶' },
    { id: 'cycling-regular', label: 'Bisiklet', icon: '🚲' }
];

const interests = [
    { id: 'history', label: 'Tarih & Kültür', categories: ['Tarihi', 'Müze', 'Antik Kent'], icon: '🏛️' },
    { id: 'nature', label: 'Doğa & Manzara', categories: ['Doğa', 'Plaj', 'Manzara'], icon: '🌲' },
    { id: 'fun', label: 'Eğlence & Alışveriş', categories: ['Alışveriş', 'Eğlence'], icon: '🛍️' },
    { id: 'mixed', label: 'Sürpriz Karışım', categories: [], icon: '✨' }
];

const durations = [
    { value: 120, label: 'Kısa Tur (2 Saat)', icon: '⚡' },
    { value: 240, label: 'Yarım Gün (4 Saat)', icon: '🌤️' },
    { value: 480, label: 'Tam Gün (8 Saat)', icon: '🎒' }
];

function toggleInterest(id: string) {
    if (id === 'mixed') {
        selectedInterests.value = ['mixed'];
    } else {
        if (selectedInterests.value.includes('mixed')) {
            selectedInterests.value = [];
        }
        if (selectedInterests.value.includes(id)) {
            selectedInterests.value = selectedInterests.value.filter(i => i !== id);
        } else {
            selectedInterests.value.push(id);
        }
    }
}

function generateRoute() {
    if (!props.userLocation) {
        alert(t('map.location_not_found'));
        return;
    }

    isGenerating.value = true;

    setTimeout(() => {
        // 1. Filter by Interest
        let candidates = props.places.filter(p => p.status === 'approved');
        
        if (!selectedInterests.value.includes('mixed')) {
            const allowedCategories = selectedInterests.value.flatMap(i => 
                interests.find(x => x.id === i)?.categories || []
            );
            candidates = candidates.filter(p => allowedCategories.includes(p.category));
        }

        // 2. Filter by reasonable radius based on duration
        // 2h -> 30km, 4h -> 70km, 8h -> 150km
        const maxRadius = (selectedDuration.value || 120) <= 120 ? 30 : 
                          (selectedDuration.value || 120) <= 240 ? 70 : 150;

        const filteredByDist = candidates.map(p => {
            const dist = getDistance(props.userLocation!.lat, props.userLocation!.lng, p.lat, p.lng);
            return { place: p, dist };
        }).filter(item => item.dist <= maxRadius);

        // If no places found in radius, try a larger one but warn or just take closest
        let sourceList = filteredByDist;
        if (sourceList.length < 2) {
            // Fallback: Just take closest places regardless of radius but limit to N
            sourceList = candidates.map(p => {
                const dist = getDistance(props.userLocation!.lat, props.userLocation!.lng, p.lat, p.lng);
                return { place: p, dist };
            }).sort((a, b) => a.dist - b.dist).slice(0, 5);
        }

        if (sourceList.length === 0) {
            alert(t('ui.no_results'));
            isGenerating.value = false;
            return;
        }

        // 3. Select N places based on Duration
        const placeCount = Math.max(2, Math.floor((selectedDuration.value || 120) / 60));
        const selectedToRoute = sourceList.slice(0, placeCount).map(s => s.place);

        // 4. LOGICAL SEQUENCING: Nearest Neighbor Algorithm
        // Start from user location, pick nearest, then from that pick nearest etc.
        const ordered: any[] = [];
        let currentLat = props.userLocation!.lat;
        let currentLng = props.userLocation!.lng;
        const remaining = [...selectedToRoute];

        while (remaining.length > 0) {
            let nearestIdx = 0;
            let minDist = Infinity;

            for (let i = 0; i < remaining.length; i++) {
                const d = getDistance(currentLat, currentLng, remaining[i].lat, remaining[i].lng);
                if (d < minDist) {
                    minDist = d;
                    nearestIdx = i;
                }
            }

            const nextPlace = remaining.splice(nearestIdx, 1)[0];
            ordered.push(nextPlace);
            currentLat = nextPlace.lat;
            currentLng = nextPlace.lng;
        }

        emit('create-route', ordered, transportMode.value);
        emit('close');
    }, 1500); // Fake processing delay for UX
}
</script>

<template>
  <div class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-900 w-full max-w-lg rounded-3xl shadow-2xl border border-slate-200 dark:border-zinc-800 overflow-hidden animate-in zoom-in-95 duration-200 relative">
        
        <!-- Decoration -->
        <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-emerald-400 via-teal-500 to-cyan-500"></div>

        <div class="p-8">
            <!-- Header -->
            <div class="text-center mb-8">
                <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-50 dark:bg-emerald-900/30 text-3xl mb-4 shadow-sm text-emerald-600">
                    🤖
                </div>
                <h2 class="text-2xl font-bold text-slate-900 dark:text-white m-0">Akıllı Gezi Asistanı</h2>
                <p class="text-slate-500 dark:text-zinc-400 mt-2 text-sm">Sana özel bir rota oluşturmam için birkaç soruya cevap ver.</p>
            </div>

            <!-- Steps -->
            <div v-if="!isGenerating">
                
                <!-- Step 1: Duration -->
                <div v-if="step === 1" class="animate-in slide-in-from-right-8 fade-in duration-300">
                    <h3 class="text-lg font-semibold mb-4 text-slate-800 dark:text-zinc-200">Ne kadar vaktin var? ⏳</h3>
                    <div class="grid grid-cols-1 gap-3">
                        <button 
                            v-for="dur in durations" 
                            :key="dur.value"
                            @click="selectedDuration = dur.value; step = 2"
                            class="flex items-center p-4 rounded-xl border-2 transition-all cursor-pointer text-left group hover:scale-[1.02]"
                            :class="selectedDuration === dur.value ? 'border-emerald-500 bg-emerald-50 dark:bg-emerald-900/20' : 'border-slate-100 dark:border-zinc-800 hover:border-emerald-200 dark:hover:border-zinc-700'"
                        >
                            <span class="text-3xl mr-4 group-hover:scale-110 transition-transform">{{ dur.icon }}</span>
                            <div>
                                <span class="block font-bold text-slate-900 dark:text-white">{{ dur.label }}</span>
                                <span class="text-xs text-slate-500">Ortalama {{ Math.floor(dur.value/60) }} mekan</span>
                            </div>
                        </button>
                    </div>
                </div>

                <!-- Step 2: Interests & Transport -->
                <div v-else-if="step === 2" class="animate-in slide-in-from-right-8 fade-in duration-300">
                    <h3 class="text-lg font-semibold mb-3 text-slate-800 dark:text-zinc-200">Nelerden hoşlanırsın? ❤️</h3>
                    <div class="grid grid-cols-2 gap-2 mb-6">
                        <button 
                            v-for="int in interests" 
                            :key="int.id"
                            @click="toggleInterest(int.id)"
                            class="flex flex-col items-center justify-center p-3 rounded-xl border-2 transition-all cursor-pointer h-24 gap-1 hover:scale-105"
                            :class="selectedInterests.includes(int.id) ? 'border-emerald-500 bg-emerald-50 dark:bg-emerald-900/20 shadow-md scale-105' : 'border-slate-100 dark:border-zinc-800 bg-slate-50 dark:bg-zinc-800/50 hover:border-emerald-200'"
                        >
                            <span class="text-2xl">{{ int.icon }}</span>
                            <span class="font-bold text-xs text-center text-slate-700 dark:text-zinc-300">{{ int.label }}</span>
                        </button>
                    </div>

                    <h3 class="text-lg font-semibold mb-3 text-slate-800 dark:text-zinc-200">Nasıl gezeceksin? 🚲</h3>
                    <div class="flex gap-2 mb-8">
                        <button 
                            v-for="opt in transportOptions" 
                            :key="opt.id"
                            @click="transportMode = opt.id"
                            class="flex-1 flex flex-col items-center py-3 rounded-xl border-2 transition-all cursor-pointer gap-1"
                            :class="transportMode === opt.id ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-slate-100 dark:border-zinc-800 hover:border-blue-200'"
                        >
                            <span class="text-xl">{{ opt.icon }}</span>
                            <span class="text-[10px] font-bold uppercase tracking-tight">{{ opt.label }}</span>
                        </button>
                    </div>
                    
                    <div class="flex gap-3">
                        <button @click="step = 1" class="flex-1 py-3 rounded-xl font-bold text-slate-500 hover:bg-slate-100 dark:hover:bg-zinc-800 transition-colors">Geri</button>
                        <button 
                            @click="generateRoute" 
                            :disabled="selectedInterests.length === 0"
                            class="flex-[2] py-3 rounded-xl font-bold text-white bg-emerald-500 shadow-lg shadow-emerald-500/30 transition-all hover:bg-emerald-600 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            ✨ Rota Oluştur
                        </button>
                    </div>
                </div>

            </div>

            <!-- Generating Animation -->
            <div v-else class="text-center py-8 animate-in fade-in zoom-in duration-500">
                <div class="relative w-24 h-24 mx-auto mb-6">
                    <div class="absolute inset-0 border-4 border-slate-100 dark:border-zinc-800 rounded-full"></div>
                    <div class="absolute inset-0 border-4 border-emerald-500 rounded-full border-t-transparent animate-spin"></div>
                    <div class="absolute inset-0 flex items-center justify-center text-4xl animate-pulse">🤖</div>
                </div>
                <h3 class="text-xl font-bold text-slate-800 dark:text-white mb-2">Mükemmel Rota Hazırlanıyor...</h3>
                <p class="text-slate-500 dark:text-zinc-400">Yapay zeka senin için en iyi durakları seçiyor.</p>
            </div>

        </div>
    </div>
  </div>
</template>