<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import api from '../api';

const { t } = useI18n();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

interface UserProfile {
  id: number;
  username: string;
  role: string;
  email: string;
  bio: string;
  avatar_url: string;
}

const user = ref<UserProfile | null>(null);
const loading = ref(true);
const isEditing = ref(false);
const editForm = ref({
  email: '',
  bio: '',
  avatar_url: ''
});

async function fetchProfile() {
  loading.value = true;
  try {
    const res = await api.get<UserProfile>('/user');
    user.value = res.data;
    resetForm();
  } catch (err) {
    console.error("Failed to fetch profile", err);
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  if (user.value) {
    editForm.value = {
      email: user.value.email,
      bio: user.value.bio,
      avatar_url: user.value.avatar_url
    };
  }
}

async function saveProfile() {
  if (!user.value) return;
  try {
    const res = await api.put<UserProfile>('/user', editForm.value);
    user.value = res.data;
    isEditing.value = false;
  } catch (err) {
    console.error("Failed to update profile", err);
    alert("Profil güncellenemedi.");
  }
}

onMounted(() => {
  fetchProfile();
});
</script>

<template>
  <div class="fixed inset-0 z-[2000] flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-800 rounded-2xl shadow-2xl w-full max-w-md overflow-hidden flex flex-col max-h-[90vh]">
      
      <!-- Header -->
      <div class="p-4 border-b border-slate-100 dark:border-zinc-700 flex justify-between items-center bg-slate-50 dark:bg-zinc-900/50">
        <h2 class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
          👤 {{ t('profile.title', 'Profilim') }}
        </h2>
        <button @click="$emit('close')" class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-slate-200 dark:hover:bg-zinc-700 text-slate-500 dark:text-zinc-400 transition-colors">
          ✕
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 overflow-y-auto">
        <div v-if="loading" class="flex justify-center py-8">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-emerald-500"></div>
        </div>

        <div v-else-if="user" class="flex flex-col gap-6">
          
          <!-- Avatar & Identity -->
          <div class="flex flex-col items-center gap-3">
            <div class="w-24 h-24 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center text-4xl overflow-hidden border-4 border-white dark:border-zinc-700 shadow-lg">
              <img v-if="user.avatar_url" :src="user.avatar_url" alt="Avatar" class="w-full h-full object-cover" />
              <span v-else>👤</span>
            </div>
            <div class="text-center">
              <h3 class="text-xl font-bold text-slate-900 dark:text-white">{{ user.username }}</h3>
              <span class="inline-block px-2 py-0.5 rounded text-xs font-bold uppercase tracking-wide bg-emerald-100 dark:bg-emerald-900 text-emerald-700 dark:text-emerald-400 mt-1">
                {{ user.role }}
              </span>
            </div>
          </div>

          <!-- View Mode -->
          <div v-if="!isEditing" class="flex flex-col gap-4">
            <div class="bg-slate-50 dark:bg-zinc-700/30 p-4 rounded-xl">
              <label class="block text-xs font-bold text-slate-400 dark:text-zinc-500 uppercase mb-1">Email</label>
              <p class="text-slate-800 dark:text-slate-200">{{ user.email || '-' }}</p>
            </div>
            
            <div class="bg-slate-50 dark:bg-zinc-700/30 p-4 rounded-xl">
              <label class="block text-xs font-bold text-slate-400 dark:text-zinc-500 uppercase mb-1">Bio</label>
              <p class="text-slate-800 dark:text-slate-200 whitespace-pre-wrap">{{ user.bio || 'Henüz bir biyografi yok.' }}</p>
            </div>

            <button @click="isEditing = true" class="w-full py-3 bg-emerald-500 hover:bg-emerald-600 text-white font-bold rounded-xl transition-colors shadow-lg shadow-emerald-500/20">
              ✏️ Profili Düzenle
            </button>
          </div>

          <!-- Edit Mode -->
          <div v-else class="flex flex-col gap-4 animate-fade-in">
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1">Avatar URL</label>
              <input v-model="editForm.avatar_url" type="text" placeholder="https://..." class="w-full p-3 rounded-xl border border-slate-300 dark:border-zinc-600 bg-white dark:bg-zinc-900 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all" />
            </div>

            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1">Email</label>
              <input v-model="editForm.email" type="email" placeholder="ornek@email.com" class="w-full p-3 rounded-xl border border-slate-300 dark:border-zinc-600 bg-white dark:bg-zinc-900 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all" />
            </div>

            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1">Hakkımda (Bio)</label>
              <textarea v-model="editForm.bio" rows="3" placeholder="Kendinizden bahsedin..." class="w-full p-3 rounded-xl border border-slate-300 dark:border-zinc-600 bg-white dark:bg-zinc-900 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all resize-none"></textarea>
            </div>

            <div class="flex gap-3 mt-2">
              <button @click="isEditing = false; resetForm()" class="flex-1 py-3 bg-slate-200 dark:bg-zinc-700 hover:bg-slate-300 dark:hover:bg-zinc-600 text-slate-700 dark:text-white font-bold rounded-xl transition-colors">
                İptal
              </button>
              <button @click="saveProfile" class="flex-1 py-3 bg-emerald-500 hover:bg-emerald-600 text-white font-bold rounded-xl transition-colors shadow-lg shadow-emerald-500/20">
                Kaydet
              </button>
            </div>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.2s ease-out;
}
@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}
</style>