<script setup lang="ts">
const props = defineProps<{
    routeWaypoints: Array<{lat: number, lng: number, name: string}>;
    transportMode: string;
    routeInfo: { roads: string[], totalDistance: string, totalTime: string } | null;
}>();

const emit = defineEmits<{
    (e: 'clear-route'): void;
    (e: 'update-mode', mode: string): void;
    (e: 'remove-waypoint', index: number): void;
    (e: 'move-waypoint', index: number, direction: 'up' | 'down'): void;
    (e: 'share-route'): void;
    (e: 'open-transit'): void;
    (e: 'close-info'): void;
}>();

const transportOptions = [
    { id: 'driving', icon: '🚗', label: 'Araba' },
    { id: 'bicycle', icon: '🚲', label: 'Bisiklet' },
    { id: 'foot', icon: '🚶', label: 'Yürüyüş' }
];
</script>

<template>
    <div class="absolute top-16 md:top-4 left-4 z-[999] flex flex-col gap-3 w-[280px]">
        
        <!-- Planner Panel -->
        <div v-if="routeWaypoints.length > 0" class="bg-white dark:bg-zinc-800 p-4 shadow-xl border border-slate-200 dark:border-zinc-700 rounded-xl animate-in fade-in slide-in-from-left-4">
            <div class="flex justify-between items-center mb-3 pb-2 border-b border-slate-100 dark:border-zinc-700">
                <span class="font-bold text-slate-900 dark:text-white flex items-center gap-2">🚦 Rota Planlayıcı <span class="bg-emerald-500 text-white text-[10px] px-1.5 py-0.5 rounded-full">{{ routeWaypoints.length }}</span></span>
                <button class="text-xs text-red-500 hover:text-red-600 bg-transparent border-none cursor-pointer" @click="$emit('clear-route')">Temizle</button>
            </div>
            
            <!-- Transport Modes -->
            <div class="flex gap-1 bg-slate-100 dark:bg-zinc-700/50 p-1 rounded-lg mb-3">
                <button 
                    v-for="mode in transportOptions" 
                    :key="mode.id"
                    @click="$emit('update-mode', mode.id)"
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
                        <button class="text-[10px] leading-none text-slate-400 hover:text-emerald-500 bg-transparent border-none cursor-pointer p-0" @click.stop="$emit('move-waypoint', index, 'up')" v-if="index > 0">▲</button>
                        <button class="text-[10px] leading-none text-slate-400 hover:text-emerald-500 bg-transparent border-none cursor-pointer p-0" @click.stop="$emit('move-waypoint', index, 'down')" v-if="index < routeWaypoints.length - 1">▼</button>
                    </div>

                    <button class="text-slate-400 hover:text-red-500 bg-transparent border-none cursor-pointer opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity ml-1" @click.stop="$emit('remove-waypoint', index)">✕</button>
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
                    <button class="bg-blue-500 hover:bg-blue-600 text-white border-none rounded p-1.5 cursor-pointer text-xs flex items-center gap-1 transition-colors" @click="$emit('open-transit')" title="Google Haritalar'da Toplu Taşıma ile Aç">
                        🚌 <span class="hidden sm:inline">Toplu Taşıma</span>
                    </button>
                    <button class="bg-transparent border-none text-slate-400 hover:text-slate-600 cursor-pointer text-lg" @click="$emit('close-info')">✕</button>
                </div>
            </div>
            
            <div class="max-h-[150px] overflow-y-auto space-y-1 pr-1">
                <div v-for="(road, index) in routeInfo.roads" :key="index" class="flex items-center gap-2 text-sm text-slate-700 dark:text-zinc-300">
                    <span class="w-1.5 h-1.5 bg-emerald-500 rounded-full flex-shrink-0"></span>
                    <span class="truncate">{{ road }}</span>
                </div>
            </div>

            <button @click="$emit('share-route')" class="w-full mt-3 bg-indigo-500 hover:bg-indigo-600 text-white border-none py-2 rounded-lg font-bold cursor-pointer transition-colors flex items-center justify-center gap-2 text-xs">
                🔗 Rotayı Paylaş
            </button>
        </div>

    </div>
</template>
