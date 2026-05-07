<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { getLocalizedContent, getCategoryEmoji } from '../../utils';
import type { Place } from '../../types';

const { t, locale } = useI18n();

const props = defineProps<{
    place: Place;
    isSelected: boolean;
    currentUser: { username: string, role: string } | null;
    translatedDescription?: string;
}>();

const emit = defineEmits<{
    (e: 'select', id: number): void;
    (e: 'toggle-favorite', id: number): void;
    (e: 'edit', place: Place): void;
    (e: 'delete', id: number): void;
    (e: 'clear-translation', id: number): void;
}>();

</script>

<template>
    <li 
        class="group bg-white dark:bg-zinc-900/80 rounded-2xl border border-slate-200/60 dark:border-zinc-800 cursor-pointer overflow-hidden transition-all duration-300 hover:shadow-xl hover:shadow-emerald-500/10 hover:-translate-y-1 hover:border-emerald-500/30 dark:hover:border-emerald-500/30 relative"
        :class="{ '!border-emerald-500 ring-2 ring-emerald-500/20 shadow-lg shadow-emerald-500/10': isSelected }"
        @click="$emit('select', place.id as number)"
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
                    {{ translatedDescription || getLocalizedContent(place.description, locale) }}
                </p>
                <button 
                    v-if="translatedDescription"
                    @click.stop="$emit('clear-translation', place.id as number)" 
                    class="absolute bottom-0 right-0 bg-slate-50/90 dark:bg-zinc-900/90 backdrop-blur px-1.5 py-0.5 text-[9px] font-bold text-emerald-600 dark:text-emerald-400 border border-emerald-100 dark:border-emerald-900/30 rounded cursor-pointer hover:bg-emerald-50"
                    title="Otomatik çevrildi. Orijinalini görmek için tıkla."
                >
                    🌐 Çevrildi
                </button>
            </div>

            <div class="flex gap-2 justify-end opacity-0 group-hover:opacity-100 transition-all duration-300 translate-y-2 group-hover:translate-y-0">
                <button class="px-3 py-1.5 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 text-xs font-bold hover:bg-emerald-100 dark:hover:bg-emerald-900/40 transition-colors" @click.stop="$emit('edit', place)">
                    {{ t('common.edit') }}
                </button>
                <button v-if="currentUser?.role === 'admin'" class="px-3 py-1.5 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-xs font-bold hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors" @click.stop="$emit('delete', place.id as number)">
                    {{ t('common.delete') }}
                </button>
            </div>
        </div>
    </li>
</template>
