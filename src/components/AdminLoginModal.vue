<script setup lang="ts">
import { ref } from 'vue';
import axios from 'axios';

const emit = defineEmits<{
  (e: 'login-success'): void;
  (e: 'close'): void;
}>();

const password = ref('');
const error = ref('');
const loading = ref(false);

async function handleLogin() {
  error.value = '';
  loading.value = true;
  try {
    const response = await axios.post('http://localhost:8080/api/admin?action=login', {
      password: password.value
    });
    
    if (response.data && response.data.success) {
      emit('login-success');
    }
  } catch (err) {
    error.value = 'Şifre hatalı!';
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="fixed inset-0 w-full h-full bg-black/70 flex justify-center items-center z-[2100] backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-800 p-8 rounded-2xl w-full max-w-[350px] shadow-2xl border border-slate-200 dark:border-zinc-700 text-slate-800 dark:text-white transition-colors duration-300">
      <div class="flex flex-col items-center mb-6">
        <div class="w-16 h-16 bg-slate-100 dark:bg-zinc-700 rounded-full flex items-center justify-center text-3xl mb-4">
            🔒
        </div>
        <h3 class="m-0 text-xl font-bold">Yönetici Girişi</h3>
        <p class="text-sm text-slate-500 dark:text-zinc-400 mt-1">Lütfen yönetici şifresini girin</p>
      </div>
      
      <form @submit.prevent="handleLogin" class="flex flex-col gap-4">
        <div class="flex flex-col gap-1.5">
          <input v-model="password" type="password" required placeholder="Şifre" autofocus
                 class="p-3 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
          <span v-if="error" class="text-red-500 text-xs font-medium">{{ error }}</span>
        </div>

        <button type="submit" :disabled="loading" 
                class="w-full py-3 rounded-lg border-none bg-emerald-500 text-white font-semibold cursor-pointer hover:bg-emerald-600 transition-colors shadow-lg shadow-emerald-500/20 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ loading ? 'Kontrol ediliyor...' : 'Giriş Yap' }}
        </button>
        <button type="button" @click="$emit('close')" class="w-full py-2 bg-transparent border-none text-slate-500 dark:text-zinc-400 text-sm hover:text-slate-700 dark:hover:text-zinc-200 cursor-pointer">
            İptal
        </button>
      </form>
    </div>
  </div>
</template>
