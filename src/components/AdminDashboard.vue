<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import api, { getAdminStats } from '../api';

const { t } = useI18n();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'place-approved'): void;
}>();

const activeTab = ref<'pending' | 'users' | 'stats'>('stats');
const pendingPlaces = ref<any[]>([]);
const users = ref<any[]>([]);
const stats = ref<any>(null);
const loading = ref(false);

const fetchPending = async () => {
  loading.value = true;
  try {
    const response = await api.get('/admin?action=pending');
    pendingPlaces.value = response.data;
  } catch (error) {
    console.error('Error fetching pending places:', error);
  } finally {
    loading.value = false;
  }
};

const fetchUsers = async () => {
  loading.value = true;
  try {
    const response = await api.get('/admin?action=users');
    users.value = response.data;
  } catch (error) {
    console.error('Error fetching users:', error);
  } finally {
    loading.value = false;
  }
};

const fetchStats = async () => {
    loading.value = true;
    try {
        stats.value = await getAdminStats();
    } catch (error) {
        console.error('Error fetching stats:', error);
    } finally {
        loading.value = false;
    }
};

const approvePlace = async (id: number) => {
  try {
    await api.post('/admin?action=approve', { id });
    pendingPlaces.value = pendingPlaces.value.filter(p => p.id !== id);
    emit('place-approved');
    fetchStats(); // Refresh stats
  } catch (error) {
    alert(t('admin.approve_error'));
  }
};

const rejectPlace = async (id: number) => {
  if (!confirm(t('admin.reject_confirm'))) return;
  try {
    await api.post('/admin?action=reject', { id });
    pendingPlaces.value = pendingPlaces.value.filter(p => p.id !== id);
    fetchStats(); // Refresh stats
  } catch (error) {
    alert(t('admin.delete_error'));
  }
};

onMounted(() => {
  fetchStats();
});

function switchTab(tab: 'pending' | 'users' | 'stats') {
    activeTab.value = tab;
    if (tab === 'pending') fetchPending();
    if (tab === 'users') fetchUsers();
    if (tab === 'stats') fetchStats();
}
</script>

<template>
  <div class="fixed inset-0 w-full h-full bg-black/60 flex justify-center items-center z-[2000] backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-900 p-6 rounded-2xl w-[90%] max-w-4xl h-[85vh] shadow-2xl border border-slate-200 dark:border-zinc-800 text-slate-800 dark:text-white transition-colors duration-300 flex flex-col">
      
      <div class="flex justify-between items-center mb-6 flex-shrink-0">
        <h2 class="m-0 text-2xl font-bold flex items-center gap-2">🛡️ {{ t('ui.admin_panel') }}</h2>
        <button class="bg-transparent border-none text-slate-400 hover:text-slate-600 dark:hover:text-zinc-200 cursor-pointer text-xl" @click="$emit('close')">✕</button>
      </div>

      <!-- Navigation Tabs -->
      <div class="flex gap-4 mb-6 border-b border-slate-200 dark:border-zinc-800 pb-1 flex-shrink-0">
        <button 
            @click="switchTab('stats')" 
            class="pb-2 px-2 font-medium transition-colors border-b-2"
            :class="activeTab === 'stats' ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400' : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-zinc-400 dark:hover:text-zinc-200'"
        >
            📊 İstatistikler
        </button>
        <button 
            @click="switchTab('pending')" 
            class="pb-2 px-2 font-medium transition-colors border-b-2 relative"
            :class="activeTab === 'pending' ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400' : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-zinc-400 dark:hover:text-zinc-200'"
        >
            {{ t('admin.pending_approvals') }}
            <span v-if="stats && stats.pending_places > 0" class="absolute -top-1 -right-2 bg-red-500 text-white text-[10px] w-4 h-4 flex items-center justify-center rounded-full">{{ stats.pending_places }}</span>
        </button>
        <button 
            @click="switchTab('users')" 
            class="pb-2 px-2 font-medium transition-colors border-b-2"
            :class="activeTab === 'users' ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400' : 'border-transparent text-slate-500 hover:text-slate-700 dark:text-zinc-400 dark:hover:text-zinc-200'"
        >
            {{ t('admin.users') }}
        </button>
      </div>

      <div class="flex-grow overflow-y-auto pr-2">
        <!-- Stats Tab -->
        <div v-if="activeTab === 'stats'" class="animate-in fade-in">
            <div v-if="loading" class="text-center py-10 text-slate-400">{{ t('common.loading') }}</div>
            <div v-else-if="stats">
                <!-- KPI Cards -->
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                    <div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-xl border border-blue-100 dark:border-blue-800">
                        <h3 class="text-sm font-bold text-blue-600 dark:text-blue-400 m-0 uppercase tracking-wide">Toplam Yer</h3>
                        <p class="text-3xl font-bold text-slate-900 dark:text-white m-0 mt-2">{{ stats.total_places }}</p>
                    </div>
                    <div class="bg-amber-50 dark:bg-amber-900/20 p-4 rounded-xl border border-amber-100 dark:border-amber-800">
                        <h3 class="text-sm font-bold text-amber-600 dark:text-amber-400 m-0 uppercase tracking-wide">Bekleyen</h3>
                        <p class="text-3xl font-bold text-slate-900 dark:text-white m-0 mt-2">{{ stats.pending_places }}</p>
                    </div>
                    <div class="bg-emerald-50 dark:bg-emerald-900/20 p-4 rounded-xl border border-emerald-100 dark:border-emerald-800">
                        <h3 class="text-sm font-bold text-emerald-600 dark:text-emerald-400 m-0 uppercase tracking-wide">Kullanıcı</h3>
                        <p class="text-3xl font-bold text-slate-900 dark:text-white m-0 mt-2">{{ stats.total_users }}</p>
                    </div>
                    <div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded-xl border border-purple-100 dark:border-purple-800">
                        <h3 class="text-sm font-bold text-purple-600 dark:text-purple-400 m-0 uppercase tracking-wide">Yorum</h3>
                        <p class="text-3xl font-bold text-slate-900 dark:text-white m-0 mt-2">{{ stats.total_comments }}</p>
                    </div>
                </div>

                <!-- Category Chart (Simple Bars) -->
                <h3 class="text-lg font-bold mb-4">Kategori Dağılımı</h3>
                <div class="space-y-3">
                    <div v-for="(count, cat) in stats.categories" :key="cat" class="flex items-center gap-3">
                        <span class="w-24 text-sm font-medium text-slate-600 dark:text-zinc-400 truncate">{{ t(`categories.${cat}`) }}</span>
                        <div class="flex-grow bg-slate-100 dark:bg-zinc-800 rounded-full h-4 overflow-hidden">
                            <div class="bg-emerald-500 h-full rounded-full transition-all duration-1000" :style="{ width: `${(count / stats.total_places) * 100}%` }"></div>
                        </div>
                        <span class="w-8 text-sm font-bold text-right">{{ count }}</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Pending Approvals Tab -->
        <div v-if="activeTab === 'pending'" class="animate-in fade-in">
            <div v-if="loading" class="text-center py-10 text-slate-400">{{ t('common.loading') }}</div>
            <div v-else-if="pendingPlaces.length === 0" class="text-center py-10 text-slate-400">
                {{ t('admin.no_pending') }}
            </div>
            <div v-else class="grid gap-4">
                <div v-for="place in pendingPlaces" :key="place.id" class="bg-slate-50 dark:bg-zinc-800 p-4 rounded-xl border border-slate-200 dark:border-zinc-700 flex flex-col md:flex-row gap-4 items-start md:items-center">
                    <img v-if="place.imageUrl" :src="place.imageUrl" class="w-20 h-20 object-cover rounded-lg bg-slate-200" />
                    <div v-else class="w-20 h-20 bg-slate-200 dark:bg-zinc-700 rounded-lg flex items-center justify-center text-2xl">📍</div>
                    
                    <div class="flex-grow">
                        <h3 class="font-bold text-lg m-0">{{ place.name }}</h3>
                        <p class="text-sm text-slate-500 dark:text-zinc-400 m-0">{{ place.city }} • {{ t(`categories.${place.category}`) }}</p>
                        <p class="text-sm mt-1 line-clamp-2">{{ place.description }}</p>
                    </div>

                    <div class="flex gap-2 w-full md:w-auto">
                        <button @click="approvePlace(place.id)" class="flex-1 md:flex-none px-4 py-2 bg-emerald-500 hover:bg-emerald-600 text-white rounded-lg font-bold transition-colors">{{ t('admin.approve') }}</button>
                        <button @click="rejectPlace(place.id)" class="flex-1 md:flex-none px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg font-bold transition-colors">{{ t('admin.reject') }}</button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Users Tab -->
        <div v-if="activeTab === 'users'" class="animate-in fade-in">
            <div v-if="loading" class="text-center py-10 text-slate-400">{{ t('common.loading') }}</div>
            <div v-else class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr class="border-b border-slate-200 dark:border-zinc-700 text-slate-500 dark:text-zinc-400 text-sm">
                            <th class="p-3">ID</th>
                            <th class="p-3">{{ t('admin.username') }}</th>
                            <th class="p-3">{{ t('admin.role') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="user in users" :key="user.id" class="border-b border-slate-100 dark:border-zinc-800 hover:bg-slate-50 dark:hover:bg-zinc-800/50 transition-colors">
                            <td class="p-3 font-mono text-xs opacity-60">#{{ user.id }}</td>
                            <td class="p-3 font-medium">{{ user.username }}</td>
                            <td class="p-3">
                                <span class="px-2 py-1 rounded text-xs font-bold uppercase tracking-wide" 
                                    :class="user.role === 'admin' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' : 'bg-slate-100 text-slate-700 dark:bg-zinc-700 dark:text-zinc-300'">
                                    {{ user.role }}
                                </span>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

      </div>
    </div>
  </div>
</template>
