<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

defineEmits<{
  (e: 'close'): void;
}>();
</script>

<template>
  <div class="fixed inset-0 w-full h-full bg-black/60 flex justify-center items-center z-[2000] backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-white dark:bg-zinc-800 p-6 rounded-2xl w-[90%] max-w-[450px] shadow-2xl border border-slate-200 dark:border-zinc-700 text-slate-800 dark:text-white transition-colors duration-300 relative overflow-hidden">
      
      <!-- Close Button -->
      <button class="absolute top-4 right-4 bg-transparent border-none text-slate-400 hover:text-slate-600 dark:text-zinc-400 dark:hover:text-zinc-200 cursor-pointer text-xl z-10" @click="$emit('close')">✕</button>

      <div class="flex flex-col items-center text-center gap-4">
        <img src="/Maplas.png" alt="Maplas Logo" class="w-20 h-20 object-contain rounded-xl shadow-md" />
        
        <div>
          <h2 class="text-2xl font-bold m-0 text-slate-900 dark:text-white tracking-tight">Maplas</h2>
          <p class="text-sm opacity-60 mt-1">{{ t('about.tagline') }}</p>
        </div>

        <p class="text-sm text-slate-600 dark:text-zinc-300 leading-relaxed max-w-sm">
          {{ t('about.description') }}
        </p>

        <div class="w-full h-[1px] bg-slate-100 dark:bg-zinc-700 my-2"></div>

        <div class="flex flex-col items-center gap-1 w-full">
          <h3 class="text-sm font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider m-0">{{ t('about.developer') }}</h3>
          
          <a href="https://tr.linkedin.com/in/ugur-andac" 
             target="_blank" 
             rel="noopener noreferrer" 
             class="text-base font-medium text-slate-900 dark:text-white no-underline hover:text-emerald-500 dark:hover:text-emerald-400 transition-colors">
             Uğur Andaç
          </a>

          <!-- LinkedIn Badge Container (Hidden or kept for script) -->
          <div class="badge-base LI-profile-badge opacity-0 absolute pointer-events-none" 
               data-locale="en_US" 
               data-size="medium" 
               data-theme="light" 
               data-type="VERTICAL" 
               data-vanity="ugur-andac" 
               data-version="v1">
          </div>
        </div>

        <div class="mt-4 text-xs text-slate-400 font-medium">
          v1.0.0 • {{ t('about.in_development') }}
        </div>
      </div>
    </div>
  </div>
</template>
