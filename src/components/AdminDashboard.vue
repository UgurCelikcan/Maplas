<script setup lang="ts">
import { ref, onMounted } from 'vue';
import api from '../api';

interface Place {
  id: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
  imageUrl?: string;
  status: string;
}

interface User {
  id: number;
  username: string;
  role: string;
}

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'place-approved'): void;
}>();

const activeTab = ref<'pending' | 'users'>('pending');
const pendingPlaces = ref<Place[]>([]);
const users = ref<User[]>([]);
const loading = ref(true);

async function fetchPendingPlaces() {
  loading.value = true;
  try {
    const response = await api.get<Place[]>('/admin?action=pending');
    if (response.data) {
        pendingPlaces.value = response.data;
    }
  } catch (error) {
    console.error('Error fetching pending places:', error);
  } finally {
    loading.value = false;
  }
}

async function fetchUsers() {
    loading.value = true;
    try {
        const response = await api.get<User[]>('/admin?action=users');
        if (response.data) {
            users.value = response.data;
        }
    } catch (error) {
        console.error('Error fetching users:', error);
    } finally {
        loading.value = false;
    }
}

async function approvePlace(id: number) {
  try {
    await api.post('/admin?action=approve', { id });
    pendingPlaces.value = pendingPlaces.value.filter(p => p.id !== id);
    emit('place-approved'); // Trigger refresh on main map
  } catch (error) {
    alert('Onaylama sırasında hata oluştu.');
  }
}

async function rejectPlace(id: number) {
  if (!confirm('Bu kaydı reddetmek ve silmek istediğinize emin misiniz?')) return;
  try {
    await api.post('/admin?action=reject', { id });
    pendingPlaces.value = pendingPlaces.value.filter(p => p.id !== id);
  } catch (error) {
    alert('Silme sırasında hata oluştu.');
  }
}

function switchTab(tab: 'pending' | 'users') {
    activeTab.value = tab;
    if (tab === 'pending') fetchPendingPlaces();
    else fetchUsers();
}

onMounted(() => {
  fetchPendingPlaces();
});
</script>

<template>
  <div class="fixed inset-0 w-full h-full bg-black/60 flex justify-center items-center z-[2050] backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-800 rounded-2xl w-[95%] max-w-4xl h-[80vh] flex flex-col shadow-2xl border border-slate-200 dark:border-zinc-700 text-slate-800 dark:text-white transition-colors duration-300 overflow-hidden">
      
      <!-- Header -->
      <div class="p-6 border-b border-slate-200 dark:border-zinc-700 flex justify-between items-center bg-slate-50 dark:bg-zinc-900/50">
        <div>
            <h2 class="m-0 text-2xl font-bold flex items-center gap-2">🛡️ Yönetici Paneli</h2>
            <div class="flex gap-4 mt-4">
                <button 
                    class="pb-2 text-sm font-semibold border-b-2 transition-colors cursor-pointer"
                    :class="activeTab === 'pending' ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400' : 'border-transparent text-slate-500 dark:text-zinc-400 hover:text-slate-700 dark:hover:text-zinc-200'"
                    @click="switchTab('pending')">
                    Onay Bekleyenler ({{ pendingPlaces.length }})
                </button>
                <button 
                    class="pb-2 text-sm font-semibold border-b-2 transition-colors cursor-pointer"
                    :class="activeTab === 'users' ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400' : 'border-transparent text-slate-500 dark:text-zinc-400 hover:text-slate-700 dark:hover:text-zinc-200'"
                    @click="switchTab('users')">
                    Kullanıcılar
                </button>
            </div>
        </div>
        <button class="bg-transparent border-none text-slate-400 hover:text-slate-600 dark:text-zinc-400 dark:hover:text-zinc-200 cursor-pointer text-2xl w-10 h-10 flex items-center justify-center rounded-full hover:bg-slate-200 dark:hover:bg-zinc-700 transition-colors" @click="$emit('close')">✕</button>
      </div>
      
      <!-- Content -->
      <div class="flex-grow overflow-y-auto p-6 bg-slate-50 dark:bg-zinc-900">
        <div v-if="loading" class="flex justify-center items-center h-full text-slate-500">
            Yükleniyor...
        </div>
        
        <!-- Pending Places Tab -->
        <div v-else-if="activeTab === 'pending'">
            <div v-if="pendingPlaces.length === 0" class="flex flex-col items-center justify-center h-full pt-20 text-slate-400 dark:text-zinc-500 gap-4">
                <span class="text-6xl">✅</span>
                <p class="text-lg font-medium">Bekleyen onay bulunmamaktadır.</p>
            </div>

            <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div v-for="place in pendingPlaces" :key="place.id" class="bg-white dark:bg-zinc-800 p-4 rounded-xl border border-slate-200 dark:border-zinc-700 shadow-sm hover:shadow-md transition-shadow">
                    <div class="flex justify-between items-start mb-3">
                        <h3 class="font-bold text-lg m-0">{{ place.name }}</h3>
                        <span class="text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400 px-2 py-1 rounded-md font-medium border border-yellow-200 dark:border-yellow-800">Bekliyor</span>
                    </div>
                    
                    <div class="text-sm text-slate-600 dark:text-zinc-300 mb-4 space-y-1">
                        <p><span class="font-semibold">Kategori:</span> {{ place.category }}</p>
                        <p><span class="font-semibold">Şehir:</span> {{ place.city }}</p>
                        <p class="line-clamp-2" :title="place.description"><span class="font-semibold">Açıklama:</span> {{ place.description }}</p>
                        <p class="text-xs text-slate-400 font-mono mt-2">Konum: {{ place.lat }}, {{ place.lng }}</p>
                    </div>

                    <div v-if="place.imageUrl" class="h-32 mb-4 rounded-lg overflow-hidden bg-slate-100 dark:bg-zinc-700">
                        <img :src="place.imageUrl" class="w-full h-full object-cover" />
                    </div>

                    <div class="flex gap-2 mt-auto">
                        <button @click="rejectPlace(place.id)" class="flex-1 py-2 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 border border-red-200 dark:border-red-800 rounded-lg font-medium hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors cursor-pointer">
                            Reddet
                        </button>
                        <button @click="approvePlace(place.id)" class="flex-1 py-2 bg-emerald-500 text-white rounded-lg font-medium hover:bg-emerald-600 transition-colors shadow-lg shadow-emerald-500/20 cursor-pointer">
                            Onayla
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Users Tab -->
        <div v-else-if="activeTab === 'users'">
            <div class="bg-white dark:bg-zinc-800 rounded-xl border border-slate-200 dark:border-zinc-700 overflow-hidden">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr class="bg-slate-50 dark:bg-zinc-900/50 border-b border-slate-200 dark:border-zinc-700 text-slate-500 dark:text-zinc-400 text-xs uppercase font-semibold">
                            <th class="p-4 w-20">ID</th>
                            <th class="p-4">Kullanıcı Adı</th>
                            <th class="p-4 w-32">Rol</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100 dark:divide-zinc-700">
                        <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50 dark:hover:bg-zinc-700/50 transition-colors">
                            <td class="p-4 text-slate-500 dark:text-zinc-500 font-mono text-sm">#{{ user.id }}</td>
                            <td class="p-4 font-medium text-slate-900 dark:text-white">{{ user.username }}</td>
                            <td class="p-4">
                                <span class="px-2 py-1 rounded text-xs font-bold uppercase"
                                      :class="user.role === 'admin' ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400 border border-purple-200 dark:border-purple-800' : 'bg-slate-100 dark:bg-zinc-700 text-slate-600 dark:text-zinc-300'">
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