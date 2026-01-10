<script setup lang="ts">
import { ref } from 'vue';
import api from '../api';

const emit = defineEmits<{
  (e: 'login-success', user: any, token: string): void;
  (e: 'close'): void;
}>();

const mode = ref<'login' | 'register'>('login');
const username = ref('');
const password = ref('');
const secretCode = ref(''); // For admin registration
const error = ref('');
const loading = ref(false);

async function handleSubmit() {
  error.value = '';
  loading.value = true;
  
  try {
    const endpoint = mode.value === 'login' ? '/login' : '/register';
    const payload = {
        username: username.value,
        password: password.value,
        ...(mode.value === 'register' ? { secret_code: secretCode.value } : {})
    };

    const response = await api.post(endpoint, payload);
    
    if (mode.value === 'login') {
        // Login successful
        emit('login-success', { username: response.data.username, role: response.data.role }, response.data.token);
    } else {
        // Registration successful, switch to login or auto-login
        alert('Kayıt başarılı! Şimdi giriş yapabilirsiniz.');
        mode.value = 'login';
        password.value = '';
    }

  } catch (err: any) {
    if (err.response && err.response.status === 409) {
        error.value = 'Bu kullanıcı adı zaten alınmış.';
    } else if (err.response && err.response.status === 401) {
        error.value = 'Kullanıcı adı veya şifre hatalı.';
    } else {
        error.value = 'Bir hata oluştu. Lütfen tekrar deneyin.';
    }
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="fixed inset-0 w-full h-full bg-black/70 flex justify-center items-center z-[2100] backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-800 p-6 rounded-2xl w-[90%] max-w-[400px] max-h-[90vh] overflow-y-auto shadow-2xl border border-slate-200 dark:border-zinc-700 text-slate-800 dark:text-white transition-colors duration-300">
      
      <!-- Tabs -->
      <div class="flex mb-6 border-b border-slate-200 dark:border-zinc-700">
        <button class="flex-1 pb-2 text-sm font-semibold border-b-2 transition-colors"
           :class="mode === 'login' ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400' : 'border-transparent text-slate-500 dark:text-zinc-400 hover:text-slate-700 dark:hover:text-zinc-200'"
           @click="mode = 'login'; error = ''">Giriş Yap</button>
        <button class="flex-1 pb-2 text-sm font-semibold border-b-2 transition-colors"
           :class="mode === 'register' ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400' : 'border-transparent text-slate-500 dark:text-zinc-400 hover:text-slate-700 dark:hover:text-zinc-200'"
           @click="mode = 'register'; error = ''">Kayıt Ol</button>
      </div>

      <div class="flex flex-col items-center mb-6">
        <div class="w-16 h-16 bg-slate-100 dark:bg-zinc-700 rounded-full flex items-center justify-center text-3xl mb-4">
            {{ mode === 'login' ? '🔑' : '📝' }}
        </div>
        <h3 class="m-0 text-xl font-bold">{{ mode === 'login' ? 'Tekrar Hoşgeldiniz' : 'Hesap Oluştur' }}</h3>
      </div>
      
      <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
        <div class="flex flex-col gap-1.5">
          <input v-model="username" type="text" required placeholder="Kullanıcı Adı"
                 class="p-3 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
        </div>

        <div class="flex flex-col gap-1.5">
          <input v-model="password" type="password" required placeholder="Şifre"
                 class="p-3 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
        </div>

        <div v-if="mode === 'register'" class="flex flex-col gap-1.5">
            <label class="text-xs text-slate-400 ml-1">Davet kodunuz varsa giriniz:</label>
            <input v-model="secretCode" type="text" placeholder="Davet Kodu (Opsiyonel)"
                 class="p-3 rounded-lg border border-slate-300 dark:border-zinc-700 bg-slate-50 dark:bg-zinc-900 text-slate-900 dark:text-white focus:outline-none focus:border-emerald-500 dark:focus:border-emerald-500 transition-colors" />
        </div>

        <span v-if="error" class="text-red-500 text-xs font-medium text-center">{{ error }}</span>

        <button type="submit" :disabled="loading" 
                class="w-full py-3 rounded-lg border-none bg-emerald-500 text-white font-semibold cursor-pointer hover:bg-emerald-600 transition-colors shadow-lg shadow-emerald-500/20 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ loading ? 'İşleniyor...' : (mode === 'login' ? 'Giriş Yap' : 'Kayıt Ol') }}
        </button>
        <button type="button" @click="$emit('close')" class="w-full py-2 bg-transparent border-none text-slate-500 dark:text-zinc-400 text-sm hover:text-slate-700 dark:hover:text-zinc-200 cursor-pointer">
            Vazgeç
        </button>
      </form>
    </div>
  </div>
</template>
