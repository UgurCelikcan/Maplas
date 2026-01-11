<script setup lang="ts">
import { ref, computed, inject } from 'vue';
import { useI18n } from 'vue-i18n';

const { t, locale } = useI18n();

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
  places: Place[];
  selectedPlaceId: number | null;
  currentUser: { username: string, role: string } | null;
  isOpen: boolean;
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
  (e: 'open-profile'): void;
  (e: 'close-sidebar'): void;
  (e: 'open-about'): void;
  (e: 'search-nearby'): void;
}>();

const isDarkMode = inject('isDarkMode', ref(true));

const searchQuery = ref('');
const selectedCategories = ref<string[]>([]);
const selectedCity = ref<string>('');

const showProfileMenu = ref(false);
const showLangMenu = ref(false);

const languages = [
    { code: 'tr', name: 'Türkçe', flag: '🇹🇷' },
    { code: 'en', name: 'English', flag: '🇺🇸' },
    { code: 'el', name: 'Ελληνικά', flag: '🇬🇷' },
    { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
    { code: 'fr', name: 'Français', flag: '🇫🇷' },
    { code: 'es', name: 'Español', flag: '🇪🇸' },
    { code: 'it', name: 'Italiano', flag: '🇮🇹' },
    { code: 'pt', name: 'Português', flag: '🇵🇹' },
    { code: 'ru', name: 'Русский', flag: '🇷🇺' },
    { code: 'ja', name: '日本語', flag: '🇯🇵' },
    { code: 'zh-CN', name: '简体中文', flag: '🇨🇳' },
    { code: 'ko', name: '한국어', flag: '🇰🇷' },
    { code: 'ar', name: 'العربية', flag: '🇸🇦' }
];

function changeLanguage(code: string) {
    locale.value = code;
    localStorage.setItem('lang', code);
    showLangMenu.value = false;
}

const categories = computed(() => {
  const uniqueCategories = new Set(props.places.map(p => p.category));
  return Array.from(uniqueCategories);
});

const cities = computed(() => {
  const uniqueCities = new Set(props.places.map(p => p.city));
  return Array.from(uniqueCities).sort();
});

const filteredPlaces = computed(() => {
  return props.places.filter(place => {
    const matchesSearch = place.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          place.city.toLowerCase().includes(searchQuery.value.toLowerCase());
    
    const matchesCategory = selectedCategories.value.length === 0 || selectedCategories.value.includes(place.category);
    
    const matchesCity = selectedCity.value === '' || place.city === selectedCity.value;

    return matchesSearch && matchesCategory && matchesCity;
  });
});

function toggleCategory(cat: string) {
  if (selectedCategories.value.includes(cat)) {
    selectedCategories.value = selectedCategories.value.filter(c => c !== cat);
  } else {
    selectedCategories.value.push(cat);
  }
}

function onSelect(id: number) {
  emit('select-place', id);
  // On mobile, close sidebar after selection to show map
  if (window.innerWidth < 768) {
      emit('close-sidebar');
  }
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
  <div 
    class="fixed inset-y-0 left-0 w-full md:w-[420px] md:relative bg-white dark:bg-zinc-900 text-slate-800 dark:text-zinc-100 flex flex-col h-full border-r border-slate-200 dark:border-zinc-800 shadow-xl z-[1000] transition-transform duration-300 transform"
    :class="isOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'"
  >
    <div class="p-6 bg-slate-50 dark:bg-zinc-800 border-b border-slate-200 dark:border-zinc-700 transition-colors duration-300">
      
      <!-- Top Bar: Logo & User Actions -->
      <div class="flex justify-between items-start mb-5 gap-2">
        <div class="flex items-center gap-3 min-w-0">
            <img src="/Maplas.png" alt="Maplas Logo" class="w-10 h-10 object-contain rounded-lg flex-shrink-0" />
            <div class="min-w-0">
                <h2 class="m-0 text-xl font-bold text-slate-900 dark:text-white tracking-tight truncate">{{ t('ui.title') }}</h2>
                <p class="m-0 mt-0.5 opacity-60 text-sm truncate">{{ t('ui.subtitle') }}</p>
            </div>
        </div>
        
        <div class="flex gap-1.5 relative flex-shrink-0">
             <!-- Mobile Close Button -->
             <button class="md:hidden w-10 h-10 flex items-center justify-center bg-transparent border border-slate-300 dark:border-zinc-700 rounded-lg text-slate-700 dark:text-white cursor-pointer transition-colors" @click="$emit('close-sidebar')">
                ✕
            </button>

            <!-- Language Switcher -->
            <div class="relative">
                <button class="w-10 h-10 flex items-center justify-center bg-transparent border border-slate-300 dark:border-zinc-700 rounded-lg text-lg cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-white/10" @click="showLangMenu = !showLangMenu" :title="locale === 'tr' ? 'Dil Seç' : 'Select Language'">
                    {{ languages.find(l => l.code === locale)?.flag }}
                </button>
                <div v-if="showLangMenu" class="absolute right-0 top-12 w-40 max-h-[300px] overflow-y-auto bg-white dark:bg-zinc-800 border border-slate-200 dark:border-zinc-700 rounded-xl shadow-xl z-[100] flex flex-col scrollbar-thin">
                    <button v-for="lang in languages" :key="lang.code" @click="changeLanguage(lang.code)" class="px-4 py-2 text-sm text-left hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors flex items-center gap-2 flex-shrink-0">
                        <span class="text-lg">{{ lang.flag }}</span> <span>{{ lang.name }}</span>
                    </button>
                </div>
            </div>

            <button class="w-10 h-10 flex items-center justify-center bg-transparent border border-slate-300 dark:border-zinc-700 rounded-lg text-slate-700 dark:text-white cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-white/10 hover:border-emerald-500" @click="$emit('open-about')" :title="t('about.title')">
                ℹ️
            </button>

            <button class="w-10 h-10 flex items-center justify-center bg-transparent border border-slate-300 dark:border-zinc-700 rounded-lg text-slate-700 dark:text-white cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-white/10 hover:border-emerald-500" @click="$emit('toggle-theme')" :title="isDarkMode ? t('ui.light_mode') : t('ui.dark_mode')">
                {{ isDarkMode ? '☀️' : '🌙' }}
            </button>
            
            <button v-if="!currentUser" class="w-10 h-10 flex items-center justify-center bg-transparent border border-slate-300 dark:border-zinc-700 rounded-lg text-slate-700 dark:text-white cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-white/10 hover:border-emerald-500" @click="$emit('open-auth')" :title="t('ui.login')">
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
                        <p class="m-0 text-xs opacity-60">{{ currentUser.role === 'admin' ? `${t('ui.admin')} 🛡️` : t('ui.user') }}</p>
                    </div>
                    <button v-if="currentUser.role === 'admin'" @click="$emit('open-admin')" class="text-left px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors flex items-center gap-2">
                        🛡️ {{ t('ui.admin_panel') }}
                    </button>
                    <button @click="$emit('open-profile')" class="text-left px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-zinc-700 transition-colors flex items-center gap-2">
                        👤 {{ t('profile.title', 'Profilim') }}
                    </button>
                    <button @click="$emit('logout')" class="text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors flex items-center gap-2">
                        🚪 {{ t('ui.logout') }}
                    </button>
                </div>
            </div>
        </div>
      </div>
      
      <div class="flex flex-col gap-3">
        <div class="flex gap-2">
            <div class="relative flex-grow">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 opacity-50 text-sm">🔍</span>
                <input 
                v-model="searchQuery" 
                type="text" 
                :placeholder="t('ui.search_placeholder')" 
                class="w-full py-2.5 pl-9 pr-3 rounded-xl border border-slate-300 dark:border-zinc-700 bg-white dark:bg-zinc-900 text-slate-900 dark:text-white text-sm transition-all outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20"
                />
            </div>
            
            <div class="relative w-1/3 min-w-[120px]">
                <select 
                    v-model="selectedCity" 
                    class="w-full h-full py-2.5 px-3 rounded-xl border border-slate-300 dark:border-zinc-700 bg-white dark:bg-zinc-900 text-slate-900 dark:text-white text-sm appearance-none outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 cursor-pointer"
                >
                    <option value="">{{ t('common.all_cities', 'Tüm Şehirler') }}</option>
                    <option v-for="city in cities" :key="city" :value="city">{{ city }}</option>
                </select>
                <span class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none opacity-50 text-xs">▼</span>
            </div>

            <button class="w-11 h-11 flex items-center justify-center bg-emerald-100 dark:bg-emerald-900 border border-emerald-300 dark:border-emerald-700 rounded-xl text-emerald-700 dark:text-emerald-300 cursor-pointer transition-colors hover:bg-emerald-200 dark:hover:bg-emerald-800 flex-shrink-0" @click="$emit('search-nearby')" :title="t('map.locate_me')">
                📍
            </button>
        </div>
        
        <div class="flex gap-2 overflow-x-auto pb-2 scrollbar-hide items-center">
            <button class="px-3 py-2 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20 text-xs font-bold whitespace-nowrap hover:bg-emerald-500/20 transition-colors cursor-pointer flex-shrink-0" @click="$emit('add-click')">
                + {{ t('ui.add_new') }}
            </button>
            <div class="w-[1px] bg-slate-300 dark:bg-zinc-700 mx-1 h-6 self-center flex-shrink-0"></div>
            
            <button 
                v-for="cat in categories" 
                :key="cat"
                class="px-3 py-1.5 rounded-full border border-transparent whitespace-nowrap text-xs font-medium cursor-pointer transition-all flex-shrink-0 select-none"
                :class="[
                selectedCategories.includes(cat) 
                    ? 'bg-emerald-500 text-slate-900 font-semibold shadow-lg shadow-emerald-500/30 ring-2 ring-emerald-500/50 ring-offset-1 dark:ring-offset-zinc-800' 
                    : 'bg-slate-100 dark:bg-zinc-700 text-slate-600 dark:text-zinc-300 hover:bg-slate-200 dark:hover:bg-zinc-600'
                ]"
                @click="toggleCategory(cat)"
            >
                {{ t(`categories.${cat}`) }}
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
          @click="onSelect(place.id as number)"
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
                        <button class="p-1 rounded bg-transparent border-none cursor-pointer hover:bg-emerald-500/10 hover:text-emerald-500" @click.stop="$emit('edit-place', place)" :title="t('common.edit')">
                            ✏️
                        </button>
                        <!-- Only show delete for admins -->
                        <button v-if="currentUser?.role === 'admin'" class="p-1 rounded bg-transparent border-none cursor-pointer hover:bg-red-500/10 hover:text-red-500" @click.stop="$emit('delete-place', place.id as number)" :title="t('common.delete')">
                            🗑️
                        </button>
                    </div>
                </div>
                <div class="flex items-center gap-2.5">
                    <span class="text-xs font-medium text-slate-400 dark:text-zinc-400 flex items-center">📍 {{ place.city }}</span>
                    <span class="text-[10px] bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 px-2 py-0.5 rounded-md font-bold uppercase tracking-wider">{{ t(`categories.${place.category}`) }}</span>
                </div>
            </div>
            <p class="m-0 text-sm text-slate-500 dark:text-zinc-400 leading-relaxed line-clamp-2">{{ place.description }}</p>
          </div>
        </li>
      </ul>
      <div v-else class="py-16 px-5 text-center text-slate-500 dark:text-zinc-500">
        <p>{{ t('ui.no_results') }}</p>
        <button class="mt-4 bg-transparent border border-slate-400 dark:border-zinc-600 text-slate-500 dark:text-zinc-500 px-4 py-2 rounded-lg cursor-pointer hover:border-slate-600 dark:hover:border-zinc-400 hover:text-slate-700 dark:hover:text-zinc-300 transition-colors" @click="searchQuery = ''; selectedCategories = []; selectedCity = ''">{{ t('ui.clear_filters') }}</button>
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