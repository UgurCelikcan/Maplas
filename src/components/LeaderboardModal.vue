<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { getLeaderboard } from '../api';
import { getUserRank } from '../gamification';

// const { t } = useI18n();
const emit = defineEmits(['close']);

interface LeaderboardUser {
    id: number;
    username: string;
    avatar_url: string;
    points: number;
}

const users = ref<LeaderboardUser[]>([]);
const loading = ref(true);

onMounted(async () => {
    try {
        const data = await getLeaderboard();
        users.value = data || [];
    } catch (e) {
        console.error(e);
    } finally {
        loading.value = false;
    }
});
</script>

<template>
  <div class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-900 w-full max-w-md rounded-3xl shadow-2xl border border-slate-200 dark:border-zinc-800 overflow-hidden animate-in zoom-in-95 duration-200">
      
      <!-- Header -->
      <div class="relative h-32 bg-gradient-to-br from-yellow-400 to-orange-500 flex items-center justify-center overflow-hidden">
        <div class="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-20"></div>
        <div class="relative text-center text-white">
            <div class="text-4xl mb-1">🏆</div>
            <h2 class="text-2xl font-bold m-0 tracking-tight text-white drop-shadow-md">Gezgin Liderler</h2>
            <p class="text-white/90 text-xs font-medium uppercase tracking-widest mt-1">En Çok Keşfedenler</p>
        </div>
        <button class="absolute top-4 right-4 w-8 h-8 flex items-center justify-center bg-black/20 hover:bg-black/30 text-white rounded-full transition-colors cursor-pointer border-none" @click="$emit('close')">✕</button>
      </div>

      <!-- List -->
      <div class="p-0 max-h-[60vh] overflow-y-auto">
        <div v-if="loading" class="p-8 text-center text-slate-400">
            <div class="animate-spin inline-block w-6 h-6 border-2 border-current border-t-transparent rounded-full mb-2"></div>
            <p>Sıralama yükleniyor...</p>
        </div>

        <div v-else-if="users.length === 0" class="p-8 text-center text-slate-400">
            Henüz sıralama oluşmadı.
        </div>

        <ul v-else class="m-0 p-0 list-none divide-y divide-slate-100 dark:divide-zinc-800">
            <li v-for="(user, index) in users" :key="user.id" class="flex items-center gap-3 p-4 hover:bg-slate-50 dark:hover:bg-zinc-800/50 transition-colors">
                
                <!-- Rank -->
                <div class="w-8 h-8 flex items-center justify-center font-bold text-lg rounded-full flex-shrink-0"
                    :class="{
                        'bg-yellow-100 text-yellow-600': index === 0,
                        'bg-slate-100 text-slate-500 dark:bg-zinc-800 dark:text-zinc-400': index > 0
                    }">
                    <span v-if="index === 0">🥇</span>
                    <span v-else-if="index === 1">🥈</span>
                    <span v-else-if="index === 2">🥉</span>
                    <span v-else>{{ index + 1 }}</span>
                </div>

                <!-- Avatar -->
                <div class="w-10 h-10 rounded-full bg-slate-200 dark:bg-zinc-700 overflow-hidden flex-shrink-0 border-2 border-white dark:border-zinc-800 shadow-sm">
                    <img v-if="user.avatar_url" :src="user.avatar_url" class="w-full h-full object-cover" />
                    <div v-else class="w-full h-full flex items-center justify-center text-slate-400 font-bold bg-slate-100 dark:bg-zinc-800">
                        {{ user.username.charAt(0).toUpperCase() }}
                    </div>
                </div>

                <!-- Info -->
                <div class="flex-grow min-w-0">
                    <div class="flex items-center gap-2">
                        <span class="font-bold text-slate-900 dark:text-white truncate">{{ user.username }}</span>
                        <span class="text-xs" :title="getUserRank(user.points).title">{{ getUserRank(user.points).icon }}</span>
                    </div>
                    <div class="text-xs text-slate-500 dark:text-zinc-500 truncate">{{ getUserRank(user.points).title }}</div>
                </div>

                <!-- Points -->
                <div class="font-mono font-bold text-emerald-600 dark:text-emerald-400 text-right">
                    {{ user.points }} XP
                </div>
            </li>
        </ul>
      </div>

      <!-- Footer -->
      <div class="p-4 bg-slate-50 dark:bg-zinc-900/50 text-center text-[10px] text-slate-400 border-t border-slate-100 dark:border-zinc-800">
        Mekan ekleyerek (+50 XP) ve yorum yaparak (+10 XP) sıralamada yüksel!
      </div>
    </div>
  </div>
</template>