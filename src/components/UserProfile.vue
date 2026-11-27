<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import api from '../api';
import { getUserPoints, getUserRank, getNextRank, getProgress } from '../gamification';

const { t } = useI18n();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'account-deleted'): void;
}>();

// --- KVKK: download my data / delete my account ---
const isExporting = ref(false);
const showDeleteConfirm = ref(false);
const deletePassword = ref('');
const deleteError = ref('');
const isDeleting = ref(false);

async function downloadMyData() {
  isExporting.value = true;
  try {
    const res = await api.get('/user?action=export', { responseType: 'blob' });
    const url = URL.createObjectURL(res.data as Blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'maplas-verilerim.json';
    a.click();
    URL.revokeObjectURL(url);
  } catch (err) {
    console.error('Export failed', err);
    alert('Verileriniz indirilemedi. Lütfen tekrar deneyin.');
  } finally {
    isExporting.value = false;
  }
}

async function deleteAccount() {
  deleteError.value = '';
  if (!deletePassword.value) {
    deleteError.value = 'Şifrenizi girin.';
    return;
  }
  isDeleting.value = true;
  try {
    await api.delete('/user', { data: { password: deletePassword.value } });
    alert('Hesabınız ve kişisel verileriniz silindi.');
    emit('account-deleted');
  } catch (err: any) {
    deleteError.value = err?.response?.data?.error || 'Hesap silinemedi. Lütfen tekrar deneyin.';
  } finally {
    isDeleting.value = false;
  }
}

interface UserProfile {
  id: number;
  username: string;
  role: string;
  email: string;
  bio: string;
  avatar_url: string;
  points?: number; // Backend might not send this yet, handled by util
}

const user = ref<UserProfile | null>(null);
const loading = ref(true);
const isEditing = ref(false);

// Gamification Stats
const points = computed(() => user.value ? getUserPoints(user.value) : 0);
const rank = computed(() => getUserRank(points.value));
const nextRank = computed(() => getNextRank(points.value));
const progress = computed(() => getProgress(points.value));

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
              <h3 class="text-xl font-bold text-slate-900 dark:text-white flex items-center justify-center gap-2">
                  {{ user.username }} 
                  <span class="text-xs px-2 py-0.5 rounded-full bg-slate-100 dark:bg-zinc-700 text-slate-500 font-normal" v-if="user.role === 'admin'">Admin</span>
              </h3>
              <div class="mt-2 flex flex-col items-center">
                  <span class="text-2xl" :title="rank.title">{{ rank.icon }}</span>
                  <span class="font-bold text-sm" :class="rank.color">{{ rank.title }}</span>
                  <span class="text-[10px] text-slate-400 font-mono">{{ points }} XP</span>
              </div>
            </div>
          </div>

          <!-- Gamification Progress -->
          <div class="bg-slate-50 dark:bg-zinc-700/30 p-4 rounded-xl relative overflow-hidden">
             <div class="flex justify-between text-xs font-bold text-slate-500 dark:text-zinc-400 mb-1">
                 <span>İlerleme</span>
                 <span v-if="nextRank">{{ points }} / {{ nextRank.minPoints }} XP</span>
                 <span v-else>MAX</span>
             </div>
             <div class="w-full h-2.5 bg-slate-200 dark:bg-zinc-700 rounded-full overflow-hidden">
                 <div class="h-full bg-gradient-to-r from-emerald-400 to-emerald-600 transition-all duration-1000 ease-out" :style="{ width: `${progress}%` }"></div>
             </div>
             <div class="mt-2 text-[10px] text-slate-400 text-center" v-if="nextRank">
                 Sonraki rütbe <strong>{{ nextRank.title }}</strong> için {{ nextRank.minPoints - points }} XP daha kazanmalısın!
             </div>
             <div class="mt-2 text-[10px] text-emerald-500 font-bold text-center" v-else>
                 🏆 Zirvedesin! Harita Efsanesi sensin.
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

            <!-- Privacy (KVKK) -->
            <div class="border-t border-slate-100 dark:border-zinc-700 pt-4 flex flex-col gap-2">
              <h4 class="text-xs font-bold text-slate-400 dark:text-zinc-500 uppercase m-0">Gizlilik ve Verilerim</h4>
              <button @click="downloadMyData" :disabled="isExporting" class="w-full py-2.5 bg-slate-100 dark:bg-zinc-700 hover:bg-slate-200 dark:hover:bg-zinc-600 text-slate-700 dark:text-white text-sm font-semibold rounded-xl transition-colors disabled:opacity-50">
                {{ isExporting ? 'Hazırlanıyor...' : '📥 Verilerimi İndir' }}
              </button>
              <button v-if="!showDeleteConfirm" @click="showDeleteConfirm = true" class="w-full py-2.5 bg-transparent border border-red-200 dark:border-red-900/50 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 text-sm font-semibold rounded-xl transition-colors">
                🗑️ Hesabımı Sil
              </button>
              <div v-else class="p-3 rounded-xl border border-red-200 dark:border-red-900/50 bg-red-50 dark:bg-red-900/10 flex flex-col gap-2">
                <p class="text-xs text-red-600 dark:text-red-400 m-0 leading-relaxed">
                  Bu işlem <strong>geri alınamaz.</strong> Profiliniz, yorumlarınız, favorileriniz ve puanlarınız kalıcı olarak silinir.
                  Eklediğiniz yerler isminiz olmadan haritada kalır. Onaylamak için şifrenizi girin.
                </p>
                <input v-model="deletePassword" type="password" placeholder="Şifreniz" autocomplete="current-password"
                       class="w-full p-2.5 rounded-lg border border-red-200 dark:border-red-900/50 bg-white dark:bg-zinc-900 text-slate-900 dark:text-white text-sm outline-none focus:border-red-400" />
                <span v-if="deleteError" class="text-xs text-red-500 font-medium">{{ deleteError }}</span>
                <div class="flex gap-2">
                  <button @click="showDeleteConfirm = false; deletePassword = ''; deleteError = ''" class="flex-1 py-2 bg-slate-200 dark:bg-zinc-700 text-slate-700 dark:text-white text-sm font-bold rounded-lg">Vazgeç</button>
                  <button @click="deleteAccount" :disabled="isDeleting" class="flex-1 py-2 bg-red-500 hover:bg-red-600 text-white text-sm font-bold rounded-lg disabled:opacity-50">{{ isDeleting ? 'Siliniyor...' : 'Kalıcı Olarak Sil' }}</button>
                </div>
              </div>
              <div class="flex justify-center gap-3 text-[11px] mt-1">
                <a href="/kvkk.html" target="_blank" rel="noopener" class="text-slate-400 hover:text-emerald-500">KVKK Aydınlatma Metni</a>
                <a href="/gizlilik.html" target="_blank" rel="noopener" class="text-slate-400 hover:text-emerald-500">Gizlilik Politikası</a>
              </div>
            </div>
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