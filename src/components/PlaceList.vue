<script setup lang="ts">
import { computed, inject, ref, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { getLocalizedContent } from '../utils';
import { translateText } from '../api';
import { getUserPoints, getUserRank } from '../gamification';

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
  places: Place[]; // All places for filter options
  visiblePlaces: Place[]; // Filtered places to display
  selectedPlaceId: number | null;
  currentUser: { username: string, role: string } | null;
  isOpen: boolean;
  searchQuery: string;
  selectedCategories: string[];
  selectedCity: string;
  showFavoritesOnly?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select-place', id: number): void;
  (e: 'toggle-favorite', id: number): void;
  (e: 'toggle-theme'): void;
  (e: 'add-click'): void;
  (e: 'delete-place', id: number): void;
  (e: 'edit-place', place: any): void;
  (e: 'open-auth'): void;
  (e: 'logout'): void;
  (e: 'open-admin'): void;
  (e: 'open-profile'): void;
  (e: 'open-leaderboard'): void;
  (e: 'open-smart-planner'): void;
  (e: 'close-sidebar'): void;
  (e: 'open-about'): void;
  (e: 'search-nearby'): void;
  (e: 'update:searchQuery', value: string): void;
  (e: 'update:selectedCategories', value: string[]): void;
  (e: 'update:selectedCity', value: string): void;
  (e: 'update:showFavoritesOnly', value: boolean): void;
}>();

const isDarkMode = inject('isDarkMode', ref(true));

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

// Proxies for v-model
const localSearchQuery = computed({
  get: () => props.searchQuery,
  set: (val) => emit('update:searchQuery', val)
});

const localSelectedCity = computed({
  get: () => props.selectedCity,
  set: (val) => emit('update:selectedCity', val)
});

const categories = computed(() => {
  const uniqueCategories = new Set(props.places.map(p => p.category));
  return Array.from(uniqueCategories);
});

const cities = computed(() => {
  const uniqueCities = new Set(props.places.map(p => {
    // Normalize city name for display in filter
    return p.city.toLocaleLowerCase('tr-TR').split(' ').map(word => word.charAt(0).toLocaleUpperCase('tr-TR') + word.slice(1)).join(' ');
  }));
  return Array.from(uniqueCities).sort();
});

function toggleCategory(cat: string) {
  const newCategories = [...props.selectedCategories];
  if (newCategories.includes(cat)) {
    const index = newCategories.indexOf(cat);
    newCategories.splice(index, 1);
  } else {
    newCategories.push(cat);
  }
  emit('update:selectedCategories', newCategories);
}

function clearFilters() {
    emit('update:searchQuery', '');
    emit('update:selectedCategories', []);
    emit('update:selectedCity', '');
    emit('update:showFavoritesOnly', false);
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

const categorySlider = ref<HTMLElement | null>(null);
const translatedDescriptions = ref<Record<number, string>>({});
const isTranslating = ref<Record<number, boolean>>({});

// Auto-translate visible places when list changes OR locale changes
watch([() => props.visiblePlaces, locale], async ([newPlaces, newLocale]) => {
    // Clear old translations on locale change to force re-fetch
    // We compare with a stored 'lastLocale' or just clear if the watcher triggers primarily due to locale
    // Ideally we just re-translate everything visible.
    
    for (const place of newPlaces as Place[]) {
        if (!place.id) continue;
        
        const originalText = getLocalizedContent(place.description, newLocale as string);
        
        // Skip if empty
        if (!originalText) continue;

        try {
            isTranslating.value[place.id] = true;
            // Translate from auto to current locale
            const translated = await translateText(originalText, 'auto', newLocale as string);
            
            // Only update if different (avoid flashing or redundant updates)
            if (translated && translated !== originalText) {
                translatedDescriptions.value[place.id] = translated;
            } else {
                // If translation is same as original (e.g. source was same language), 
                // we might want to clear the 'translated' state so it shows as original.
                delete translatedDescriptions.value[place.id];
            }
        } catch (e) {
            // silent fail
        } finally {
            isTranslating.value[place.id] = false;
        }
    }
}, { immediate: true, deep: true });

onMounted(() => {
    const slider = categorySlider.value;
    if (!slider) return;

    let isDown = false;
    let startX = 0;
    let scrollLeft = 0;

    slider.addEventListener('mousedown', (e) => {
        isDown = true;
        slider.classList.add('active');
        startX = e.pageX - slider.offsetLeft;
        scrollLeft = slider.scrollLeft;
    });

    slider.addEventListener('mouseleave', () => {
        isDown = false;
        slider.classList.remove('active');
    });

    slider.addEventListener('mouseup', () => {
        isDown = false;
        slider.classList.remove('active');
    });

    slider.addEventListener('mousemove', (e) => {
        if (!isDown) return;
        e.preventDefault();
        const x = e.pageX - slider.offsetLeft;
        const walk = (x - startX) * 2; // scroll-fast
        slider.scrollLeft = scrollLeft - walk;
    });
});
</script>

<template>
  <div 
    class="fixed inset-y-0 left-0 w-full md:w-[420px] md:relative bg-white dark:bg-zinc-900 text-slate-800 dark:text-zinc-100 flex flex-col h-full border-r border-slate-200 dark:border-zinc-800 shadow-2xl z-[1000] transition-transform duration-300 transform"
    :class="isOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'"
  >
    <!-- Glassmorphic Sticky Header -->
    <div class="sticky top-0 z-20 px-6 py-4 bg-white/90 dark:bg-zinc-900/90 backdrop-blur-xl border-b border-slate-200/50 dark:border-zinc-700/50 transition-colors duration-300">
      
      <!-- Top Bar: Logo & User Actions -->
      <div class="flex justify-between items-center mb-5 gap-2">
        <div class="flex items-center gap-3 min-w-0 group cursor-default">
            <div class="relative">
                <div class="absolute inset-0 bg-emerald-500 blur opacity-20 group-hover:opacity-40 transition-opacity rounded-xl"></div>
                <img src="/Maplas.png" alt="Maplas Logo" class="relative w-10 h-10 object-contain rounded-xl shadow-sm" />
            </div>
            <div class="min-w-0">
                <h2 class="m-0 text-xl font-bold text-slate-900 dark:text-white tracking-tight truncate group-hover:text-emerald-500 transition-colors">{{ t('ui.title') }}</h2>
                <p class="m-0 text-[11px] font-medium uppercase tracking-widest text-slate-400 dark:text-zinc-500">{{ t('ui.subtitle') }}</p>
            </div>
        </div>
        
        <div class="flex gap-2 relative flex-shrink-0">
             <!-- Mobile Close Button -->
             <button class="md:hidden w-9 h-9 flex items-center justify-center bg-slate-100 dark:bg-zinc-800 rounded-full text-slate-700 dark:text-zinc-400 transition-all active:scale-90" @click="$emit('close-sidebar')">
                ✕
            </button>

            <!-- Language Switcher -->
            <div class="relative">
                <button class="w-9 h-9 flex items-center justify-center rounded-full text-lg cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10 active:scale-95" @click="showLangMenu = !showLangMenu" :title="locale === 'tr' ? 'Dil Seç' : 'Select Language'">
                    {{ languages.find(l => l.code === locale)?.flag }}
                </button>
                <div v-if="showLangMenu" class="absolute left-1/2 -translate-x-1/2 top-11 w-44 max-h-[300px] overflow-y-auto bg-white/95 dark:bg-zinc-800/95 backdrop-blur-md border border-slate-200 dark:border-zinc-700 rounded-2xl shadow-xl z-[100] flex flex-col py-1">
                    <button v-for="lang in languages" :key="lang.code" @click="changeLanguage(lang.code)" class="px-3 py-2.5 text-sm text-left hover:bg-emerald-50 dark:hover:bg-emerald-900/20 hover:text-emerald-600 dark:hover:text-emerald-400 transition-colors flex items-center gap-3 flex-shrink-0">
                        <span class="text-xl leading-none shadow-sm rounded-sm">{{ lang.flag }}</span> <span class="font-medium">{{ lang.name }}</span>
                    </button>
                </div>
            </div>

            <button class="w-9 h-9 flex items-center justify-center rounded-full text-slate-500 dark:text-zinc-400 cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10 hover:text-emerald-500 dark:hover:text-emerald-400 active:scale-95" @click="$emit('open-about')" :title="t('about.title')">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
            </button>

            <button class="w-9 h-9 flex items-center justify-center rounded-full text-slate-500 dark:text-zinc-400 cursor-pointer transition-all hover:bg-slate-100 dark:hover:bg-white/10 hover:text-yellow-500 active:scale-95" @click="$emit('toggle-theme')" :title="isDarkMode ? t('ui.light_mode') : t('ui.dark_mode')">
                <span v-if="isDarkMode">☀️</span>
                <span v-else>🌙</span>
            </button>
            
            <button v-if="!currentUser" class="w-9 h-9 flex items-center justify-center bg-emerald-500 text-white rounded-full shadow-lg shadow-emerald-500/30 cursor-pointer transition-all hover:bg-emerald-600 hover:shadow-emerald-500/40 active:scale-95" @click="$emit('open-auth')" :title="t('ui.login')">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
            </button>

            <div v-else class="relative">
                <button class="w-9 h-9 flex items-center justify-center bg-emerald-100 dark:bg-emerald-900/50 border border-emerald-200 dark:border-emerald-800 rounded-full text-emerald-700 dark:text-emerald-400 cursor-pointer font-bold shadow-sm transition-all hover:ring-2 hover:ring-emerald-500/20" @click="showProfileMenu = !showProfileMenu">
                    {{ currentUser.username.charAt(0).toUpperCase() }}
                </button>
                
                <!-- Profile Dropdown -->
                <div v-if="showProfileMenu" class="absolute right-0 top-11 w-56 bg-white/95 dark:bg-zinc-800/95 backdrop-blur-md border border-slate-200 dark:border-zinc-700 rounded-2xl shadow-xl z-[100] overflow-hidden flex flex-col py-1" @click="showProfileMenu = false">
                    <div class="px-4 py-3 border-b border-slate-100 dark:border-zinc-700/50 bg-slate-50/50 dark:bg-zinc-900/50">
                        <div class="flex items-center justify-between">
                            <p class="m-0 text-sm font-bold truncate text-slate-900 dark:text-white">{{ currentUser.username }}</p>
                            <span class="text-xs font-bold" :class="getUserRank(getUserPoints(currentUser)).color">
                                {{ getUserRank(getUserPoints(currentUser)).icon }}
                            </span>
                        </div>
                        <p class="m-0 text-[10px] font-medium text-slate-500 dark:text-zinc-500 mt-0.5 flex justify-between">
                            <span>{{ getUserRank(getUserPoints(currentUser)).title }}</span>
                            <span class="font-mono text-emerald-600 dark:text-emerald-400">{{ getUserPoints(currentUser) }} XP</span>
                        </p>
                    </div>
                    <button v-if="currentUser.role === 'admin'" @click="$emit('open-admin')" class="text-left px-4 py-2.5 text-sm hover:bg-emerald-50 dark:hover:bg-emerald-900/10 transition-colors flex items-center gap-2.5 text-slate-700 dark:text-zinc-200">
                        <span>🛡️</span> {{ t('ui.admin_panel') }}
                    </button>
                    <button @click="$emit('open-profile')" class="text-left px-4 py-2.5 text-sm hover:bg-emerald-50 dark:hover:bg-emerald-900/10 transition-colors flex items-center gap-2.5 text-slate-700 dark:text-zinc-200">
                        <span>👤</span> {{ t('profile.title', 'Profilim') }}
                    </button>
                    <button @click="$emit('open-leaderboard')" class="text-left px-4 py-2.5 text-sm hover:bg-emerald-50 dark:hover:bg-emerald-900/10 transition-colors flex items-center gap-2.5 text-slate-700 dark:text-zinc-200">
                        <span>🏆</span> Liderlik Tablosu
                    </button>
                    <div class="h-px bg-slate-100 dark:bg-zinc-700/50 my-1"></div>
                    <button @click="$emit('logout')" class="text-left px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors flex items-center gap-2.5">
                        <span>🚪</span> {{ t('ui.logout') }}
                    </button>
                </div>
            </div>
        </div>
      </div>
      
      <div class="flex flex-col gap-3">
        <div class="flex gap-2">
            <div class="relative flex-grow group">
                <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-emerald-500 transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                </span>
                <input 
                v-model="localSearchQuery" 
                type="text" 
                :placeholder="t('ui.search_placeholder')" 
                class="w-full py-2.5 pl-10 pr-4 rounded-xl border border-slate-200 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-800/50 text-slate-900 dark:text-white text-sm transition-all outline-none focus:bg-white dark:focus:bg-zinc-800 focus:border-emerald-500 focus:ring-4 focus:ring-emerald-500/10 shadow-sm"
                />
            </div>
            
            <div class="relative w-1/3 min-w-[120px] group">
                <select 
                    v-model="localSelectedCity" 
                    class="w-full h-full py-2.5 pl-3 pr-8 rounded-xl border border-slate-200 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-800/50 text-slate-900 dark:text-white text-sm appearance-none outline-none focus:bg-white dark:focus:bg-zinc-800 focus:border-emerald-500 focus:ring-4 focus:ring-emerald-500/10 shadow-sm cursor-pointer transition-all"
                >
                    <option value="">{{ t('common.all_cities', 'Tüm Şehirler') }}</option>
                    <option v-for="city in cities" :key="city" :value="city">{{ city }}</option>
                </select>
                <span class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-slate-400 group-hover:text-emerald-500 transition-colors text-xs">▼</span>
            </div>
            
            <button class="w-11 h-11 flex items-center justify-center bg-white dark:bg-zinc-800 border border-slate-200 dark:border-zinc-700 rounded-xl text-slate-500 dark:text-zinc-400 cursor-pointer transition-all hover:bg-emerald-500 hover:text-white hover:border-emerald-500 dark:hover:border-emerald-500 shadow-sm active:scale-95 flex-shrink-0 group" @click="$emit('search-nearby')" :title="t('map.locate_me')">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="group-hover:animate-pulse"><polygon points="3 11 22 2 13 21 11 13 3 11"></polygon></svg>
            </button>
        </div>
        
                        <div class="relative mt-1">
        
                            <div ref="categorySlider" class="flex gap-2 overflow-x-auto pb-3 -mb-3 pt-1 items-center snap-x snap-mandatory touch-pan-x" id="category-slider">
        
                                <button  
        
                            class="px-4 py-2 rounded-full bg-emerald-500 text-white shadow-lg shadow-emerald-500/20 text-xs font-bold whitespace-nowrap hover:bg-emerald-600 hover:scale-105 transition-all cursor-pointer flex-shrink-0 flex items-center gap-1.5 snap-start" @click="$emit('add-click')">
        
                            <span>+</span> {{ t('ui.add_new') }}
        
                        </button>

                        <button  
                            class="px-4 py-2 rounded-full bg-gradient-to-r from-violet-500 to-fuchsia-500 text-white shadow-lg shadow-fuchsia-500/20 text-xs font-bold whitespace-nowrap hover:brightness-110 hover:scale-105 transition-all cursor-pointer flex-shrink-0 flex items-center gap-1.5 snap-start" @click="$emit('open-smart-planner')">
                            <span>✨</span> {{ t('ui.smart_planner', 'Asistan') }}
                        </button>
                
                <div class="w-px bg-slate-200 dark:bg-zinc-700 mx-1 h-5 self-center flex-shrink-0"></div>
                
                <button 
                    v-if="searchQuery || selectedCategories.length > 0 || selectedCity || showFavoritesOnly"
                    @click="clearFilters"
                    class="px-4 py-2 rounded-full bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 border border-red-100 dark:border-red-800/30 text-[10px] font-bold uppercase tracking-wider flex-shrink-0 hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors cursor-pointer snap-start"
                >
                    ✕ {{ t('ui.clear_filters') }}
                </button>

                <button 
                    v-if="currentUser"
                    class="px-4 py-2 rounded-full border border-transparent whitespace-nowrap text-xs font-medium cursor-pointer transition-all flex-shrink-0 select-none active:scale-95 snap-start"
                    :class="[
                    showFavoritesOnly 
                        ? 'bg-red-500 text-white shadow-lg shadow-red-500/20' 
                        : 'bg-red-50 dark:bg-red-900/10 text-red-500 hover:bg-red-100'
                    ]"
                    @click="$emit('update:showFavoritesOnly', !showFavoritesOnly)"
                >
                    ❤️ {{ t('ui.favorites') }}
                </button>
                
                <button 
                    v-for="cat in categories" 
                    :key="cat"
                    class="px-4 py-2 rounded-full border border-transparent whitespace-nowrap text-xs font-medium cursor-pointer transition-all flex-shrink-0 select-none active:scale-95 snap-start"
                    :class="[
                    selectedCategories.includes(cat) 
                        ? 'bg-slate-800 dark:bg-white text-white dark:text-slate-900 shadow-lg' 
                        : 'bg-slate-100 dark:bg-zinc-800 text-slate-600 dark:text-zinc-300 hover:bg-slate-200 dark:hover:bg-zinc-700'
                    ]"
                    @click="toggleCategory(cat)"
                >
                    {{ t(`categories.${cat}`) }}
                </button>
                <!-- Extra space at end for better scroll experience -->
                <div class="w-8 h-1 flex-shrink-0"></div>
            </div>
            
            <!-- Fade Indicators -->
            <div class="absolute inset-y-0 left-0 w-4 bg-gradient-to-r from-white dark:from-zinc-900 to-transparent pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity"></div>
            <div class="absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-white dark:from-zinc-900 to-transparent pointer-events-none"></div>
        </div>
      </div>
    </div>

    <!-- Place List -->
    <div class="flex-grow overflow-y-auto bg-slate-50 dark:bg-zinc-950 p-4 transition-colors duration-300" @click="showProfileMenu = false">
      <ul v-if="visiblePlaces.length > 0" class="m-0 p-0 list-none flex flex-col gap-4 pb-20 md:pb-4">
        <li 
          v-for="place in visiblePlaces" 
          :key="place.id" 
          class="group bg-white dark:bg-zinc-900/80 rounded-2xl border border-slate-200/60 dark:border-zinc-800 cursor-pointer overflow-hidden transition-all duration-300 hover:shadow-xl hover:shadow-emerald-500/10 hover:-translate-y-1 hover:border-emerald-500/30 dark:hover:border-emerald-500/30 relative"
          :class="{ '!border-emerald-500 ring-2 ring-emerald-500/20 shadow-lg shadow-emerald-500/10': place.id === selectedPlaceId }"
          @click="onSelect(place.id as number)"
        >
          <div v-if="place.imageUrl" class="w-full h-48 overflow-hidden bg-slate-100 dark:bg-zinc-800 relative">
            <img :src="place.imageUrl" :alt="getLocalizedContent(place.name, locale)" loading="lazy" class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-110" />
            <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-60"></div>
            <div class="absolute bottom-3 left-4 right-4 flex justify-between items-end">
                 <span class="text-white font-bold text-lg drop-shadow-md truncate">{{ getLocalizedContent(place.name, locale) }}</span>
                 <span class="text-xs bg-white/20 backdrop-blur-md text-white px-2 py-1 rounded-md border border-white/20 shadow-sm">{{ getCategoryEmoji(place.category) }}</span>
            </div>
          </div>
          
          <div class="p-4 pt-3">
            <div v-if="!place.imageUrl" class="flex justify-between items-start mb-2">
                 <h3 class="m-0 text-lg font-bold text-slate-900 dark:text-white leading-tight truncate">{{ getLocalizedContent(place.name, locale) }}</h3>
                 <span class="text-xl">{{ getCategoryEmoji(place.category) }}</span>
            </div>

            <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-3">
                    <span class="text-xs font-semibold text-slate-500 dark:text-zinc-400 flex items-center gap-1">
                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg>
                        {{ place.city }}
                    </span>
                    <span class="w-1 h-1 rounded-full bg-slate-300 dark:bg-zinc-700"></span>
                    <span class="text-[10px] text-emerald-600 dark:text-emerald-400 font-bold uppercase tracking-wider">{{ t(`categories.${place.category}`) }}</span>
                </div>
                
                <button 
                    @click.stop="$emit('toggle-favorite', place.id as number)" 
                    class="w-8 h-8 flex items-center justify-center rounded-full transition-all hover:bg-red-50 dark:hover:bg-red-900/20 active:scale-75 group/heart"
                    :title="place.is_favorite ? t('ui.remove_favorite', 'Favorilerden Çıkar') : t('ui.add_favorite', 'Favorilere Ekle')"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" :viewBox="place.is_favorite ? '0 0 24 24' : '0 0 24 24'" :fill="place.is_favorite ? '#ef4444' : 'none'" :stroke="place.is_favorite ? '#ef4444' : 'currentColor'" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="transition-colors" :class="place.is_favorite ? '' : 'text-slate-300 group-hover/heart:text-red-400'">
                        <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
                    </svg>
                </button>
            </div>
            
            <div class="relative">
                <p class="m-0 text-sm text-slate-600 dark:text-zinc-400 leading-relaxed line-clamp-3 mb-2 transition-all">
                    {{ translatedDescriptions[place.id as number] || getLocalizedContent(place.description, locale) }}
                </p>
                <!-- Translation Indicator / Toggle -->
                <button 
                    v-if="translatedDescriptions[place.id as number]"
                    @click.stop="delete translatedDescriptions[place.id as number]" 
                    class="absolute bottom-0 right-0 bg-slate-50/90 dark:bg-zinc-900/90 backdrop-blur px-1.5 py-0.5 text-[9px] font-bold text-emerald-600 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/30 rounded cursor-pointer hover:bg-emerald-50"
                    title="Otomatik çevrildi. Orijinalini görmek için tıkla."
                >
                    🌐 Çevrildi
                </button>
            </div>

            <div class="flex gap-2 justify-end opacity-0 group-hover:opacity-100 transition-all duration-300 translate-y-2 group-hover:translate-y-0">
                <button class="px-3 py-1.5 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 text-xs font-bold hover:bg-emerald-100 dark:hover:bg-emerald-900/40 transition-colors" @click.stop="$emit('edit-place', place)">
                    {{ t('common.edit') }}
                </button>
                <button v-if="currentUser?.role === 'admin'" class="px-3 py-1.5 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-xs font-bold hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors" @click.stop="$emit('delete-place', place.id as number)">
                    {{ t('common.delete') }}
                </button>
            </div>
          </div>
        </li>
      </ul>
      <div v-else class="h-full flex flex-col items-center justify-center text-center text-slate-400 dark:text-zinc-500 pb-20">
        <div class="w-16 h-16 bg-slate-100 dark:bg-zinc-800 rounded-full flex items-center justify-center mb-4 text-3xl opacity-50">
            🔍
        </div>
        <p class="font-medium mb-1">{{ t('ui.no_results') }}</p>
        <p class="text-xs max-w-[200px] mx-auto opacity-60 mb-6">Arama kriterlerinizi değiştirerek tekrar deneyin.</p>
        <button class="bg-white dark:bg-zinc-800 border border-slate-200 dark:border-zinc-700 text-slate-600 dark:text-zinc-300 px-5 py-2.5 rounded-xl text-sm font-semibold shadow-sm hover:shadow-md hover:border-emerald-500/50 transition-all active:scale-95" @click="clearFilters">{{ t('ui.clear_filters') }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Scoped styles if needed */
</style>