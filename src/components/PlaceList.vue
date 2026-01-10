<script setup lang="ts">
import { ref, computed, inject } from 'vue';

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
  currentUser: { username: string, role: string } | null;
}>();

const emit = defineEmits<{
  (e: 'select-place', id: number): void;
  (e: 'toggle-theme'): void;
  (e: 'add-click'): void;
  (e: 'delete-place', id: number): void;
  (e: 'edit-place', place: any): void;
  (e: 'open-auth'): void;
  (e: 'logout'): void;
  (e: 'open-admin'): void;
}>();

const isDarkMode = inject('isDarkMode', ref(true));

const searchQuery = ref('');
const selectedCategory = ref('Tümü');
const showProfileMenu = ref(false);

const categories = computed(() => {
  const uniqueCategories = new Set(props.places.map(p => p.category));
  return ['Tümü', ...Array.from(uniqueCategories)];
});

const filteredPlaces = computed(() => {
  return props.places.filter(place => {
    const matchesSearch = place.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          place.city.toLowerCase().includes(searchQuery.value.toLowerCase());
    const matchesCategory = selectedCategory.value === 'Tümü' || place.category === selectedCategory.value;
    return matchesSearch && matchesCategory;
  });
});

function onSelect(id: number) {
  emit('select-place', id);
}

function getCategoryEmoji(category: string) {
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
}
</script>

<template>
  <div class="w-[420px] bg-white dark:bg-zinc-900 text-slate-800 dark:text-zinc-100 flex flex-col h-full border-r border-slate-200 dark:border-zinc-800 shadow-xl z-50 transition-colors duration-300">
    <div class="p-6 bg-slate-50 dark:bg-zinc-800 border-b border-slate-200 dark:border-zinc-700 transition-colors duration-300">
      
      <!-- Top Bar: Logo & User Actions -->
      <div class="flex justify-between items-start mb-5">
        <div class="flex items-center gap-3">
            <span class="text-3xl">🗺️</span>
            <div>
                <h2 class="m-0 text-xl font-bold text-slate-900 dark:text-white tracking-tight">Türkiye Rehberi</h2>
                <p class="m-0 mt-0.5 opacity-60 text-sm">Keşfedilecek Rotalar</p>
            </div>
        </div>
        
        <div class="flex gap-2 relative">
            <button class="w-10 h-10 flex items-center justify-center bg-transparent border border-slate-300 dark:border-zinc-700 rounded-lg text-slate-700 dark:text-white cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-white/10 hover:border-emerald-500" @click="$emit('toggle-theme')" :title="isDarkMode ? 'Aydınlık Mod' : 'Karanlık Mod'">
                {{ isDarkMode ? '☀️' : '🌙' }}
            </button>
            
            <button v-if="!currentUser" class="w-10 h-10 flex items-center justify-center bg-transparent border border-slate-300 dark:border-zinc-700 rounded-lg text-slate-700 dark:text-white cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-white/10 hover:border-emerald-500" @click="$emit('open-auth')" title="Giriş Yap">
                👤
            </button>

            <div v-else class="relative">
                <button class="w-10 h-10 flex items-center justify-center bg-emerald-100 dark:bg-emerald-900 border border-emerald-300 dark:border-emerald-700 rounded-lg text-emerald-700 dark:text-emerald-300 cursor-pointer font-bold" @click="showProfileMenu = !showProfileMenu">
                    {{ currentUser.username.charAt(0).toUpperCase() }}
                </button>
                
                <!-- Profile Dropdown -->
                <div v-if="showProfileMenu" class="absolute right-0 top-12 w-48 bg-white dark:bg-zinc-800 border border-slate-200 dark:border-zinc-700 rounded-xl shadow-xl z-[100] overflow-hidden flex flex-col" @click="showProfileMenu = false">
                    <div class="p-3 border-b border-slate-100 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900/50">
                        <p class="m-0 text-sm font-bold truncate">{{ currentUser.username }}</p>
                        <p class="m-0 text-xs opacity-60">{{ currentUser.role === 'admin' ? 'Yönetici 🛡️' : 'Kullanıcı' }}</p>
                    </div>
                    <button v-if="currentUser.role === 'admin'" @click="$emit('open-admin')" class="text-left px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors flex items-center gap-2">
                        🛡️ Yönetici Paneli
                    </button>
                    <button @click="$emit('logout')" class="text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors flex items-center gap-2">
                        🚪 Çıkış Yap
                    </button>
                </div>
            </div>
            <!-- Click outside handler could be added here for perfection -->
        </div>
      </div>
      
      <div class="flex flex-col gap-4">
        <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 opacity-50 text-sm">🔍</span>
            <input 
            v-model="searchQuery" 
            type="text" 
            placeholder="Şehir veya yer ara..." 
            class="w-full py-3 pl-10 pr-3 rounded-xl border border-slate-300 dark:border-zinc-700 bg-white dark:bg-zinc-900 text-slate-900 dark:text-white text-sm transition-all outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20"
            />
        </div>
        
        <div class="flex gap-2 overflow-x-auto pb-2 scrollbar-hide">
            <button class="px-3 py-2 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 text-xs font-bold whitespace-nowrap hover:bg-emerald-500/20 transition-colors cursor-pointer" @click="$emit('add-click')">
                + Yeni Ekle
            </button>
            <div class="w-[1px] bg-slate-300 dark:bg-zinc-700 mx-1 h-8 self-center"></div>
            <button 
                v-for="cat in categories" 
                :key="cat"
                class="px-4 py-2 rounded-full border border-transparent whitespace-nowrap text-xs font-medium cursor-pointer transition-all"
                :class="[
                selectedCategory === cat 
                    ? 'bg-emerald-500 text-slate-900 font-semibold shadow-lg shadow-emerald-500/30' 
                    : 'bg-slate-100 dark:bg-zinc-700 text-slate-600 dark:text-zinc-300 hover:bg-slate-200 dark:hover:bg-zinc-600 hover:text-slate-900 dark:hover:text-white'
                ]"
                @click="selectedCategory = cat"
            >
                {{ cat }}
            </button>
        </div>
      </div>
    </div>

    <div class="flex-grow overflow-y-auto bg-slate-50 dark:bg-zinc-900 p-4 transition-colors duration-300" @click="showProfileMenu = false">
      <ul v-if="filteredPlaces.length > 0" class="m-0 p-0 list-none flex flex-col gap-3">
        <li 
          v-for="place in filteredPlaces" 
          :key="place.id" 
          class="group bg-white dark:bg-zinc-800 rounded-2xl border border-slate-200 dark:border-zinc-700 cursor-pointer overflow-hidden transition-all duration-200 hover:-translate-y-0.5 hover:bg-slate-50 dark:hover:bg-zinc-700 hover:shadow-lg dark:hover:shadow-black/30 hover:border-slate-300 dark:hover:border-zinc-600"
          :class="{ '!border-emerald-500 !bg-emerald-50/50 dark:!bg-zinc-800 ring-2 ring-emerald-500/40': place.id === selectedPlaceId }"
          @click="onSelect(place.id)"
        >
          <div v-if="place.imageUrl" class="w-full h-40 overflow-hidden bg-slate-200 dark:bg-zinc-700">
            <img :src="place.imageUrl" :alt="place.name" loading="lazy" class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105" />
          </div>
          <div class="p-4">
            <div class="mb-3">
                <div class="flex justify-between items-start mb-1.5">
                    <div class="flex items-center gap-2 flex-grow overflow-hidden">
                        <h3 class="m-0 text-base font-semibold text-slate-900 dark:text-white leading-tight truncate">{{ place.name }}</h3>
                        <span class="flex items-center justify-center w-8 h-8 rounded-full bg-slate-100 dark:bg-white/5 flex-shrink-0 text-lg" :title="place.category">{{ getCategoryEmoji(place.category) }}</span>
                    </div>
                    <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                        <button class="p-1 rounded bg-transparent border-none cursor-pointer hover:bg-emerald-500/10 hover:text-emerald-500" @click.stop="$emit('edit-place', place)" title="Düzenle">
                            ✏️
                        </button>
                        <!-- Only show delete for admins -->
                        <button v-if="currentUser?.role === 'admin'" class="p-1 rounded bg-transparent border-none cursor-pointer hover:bg-red-500/10 hover:text-red-500" @click.stop="$emit('delete-place', place.id)" title="Sil">
                            🗑️
                        </button>
                    </div>
                </div>
                <div class="flex items-center gap-2.5">
                    <span class="text-xs font-medium text-slate-400 dark:text-zinc-400 flex items-center">📍 {{ place.city }}</span>
                    <span class="text-[10px] bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 px-2 py-0.5 rounded-md font-bold uppercase tracking-wider">{{ place.category }}</span>
                </div>
            </div>
            <p class="m-0 text-sm text-slate-500 dark:text-zinc-400 leading-relaxed line-clamp-2">{{ place.description }}</p>
          </div>
        </li>
      </ul>
      <div v-else class="py-16 px-5 text-center text-slate-500 dark:text-zinc-500">
        <p>Aradığınız kriterlere uygun yer bulunamadı.</p>
        <button class="mt-4 bg-transparent border border-slate-400 dark:border-zinc-600 text-slate-500 dark:text-zinc-500 px-4 py-2 rounded-lg cursor-pointer hover:border-slate-600 dark:hover:border-zinc-400 hover:text-slate-700 dark:hover:text-zinc-300 transition-colors" @click="searchQuery = ''; selectedCategory = 'Tümü'">Filtreleri Temizle</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
    display: none;
}
.scrollbar-hide {
    -ms-overflow-style: none;
    scrollbar-width: none;
}
</style>