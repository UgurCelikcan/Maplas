<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import axios from 'axios';

const { t, locale } = useI18n();

interface Comment {
  id: number;
  place_id: number;
  content: string;
  rating: number;
  created_at: string;
}

const props = defineProps<{
  placeId: number;
  placeName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const comments = ref<Comment[]>([]);
const newComment = ref('');
const newRating = ref(5);
const isLoading = ref(false);

const fetchComments = async () => {
  try {
    const response = await axios.get<Comment[]>(`http://localhost:8080/api/comments?place_id=${props.placeId}`);
    if (response.data) {
      comments.value = response.data;
    }
  } catch (error) {
    console.error('Error fetching comments:', error);
  }
};

const submitComment = async () => {
  if (!newComment.value.trim()) return;
  
  isLoading.value = true;
  try {
    const payload = {
      place_id: props.placeId,
      content: newComment.value,
      rating: newRating.value
    };
    
    const response = await axios.post<Comment>('http://localhost:8080/api/comments', payload);
    if (response.status === 201) {
      comments.value.unshift(response.data);
      newComment.value = '';
      newRating.value = 5;
    }
  } catch (error) {
    console.error('Error saving comment:', error);
    alert(t('comments.error_save'));
  } finally {
    isLoading.value = false;
  }
};

const formatDate = (dateStr: string) => {
  const loc = locale.value === 'el' ? 'el-GR' : (locale.value === 'en' ? 'en-US' : 'tr-TR');
  return new Date(dateStr).toLocaleDateString(loc, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

onMounted(() => {
  fetchComments();
});
</script>

<template>
  <div class="fixed inset-0 z-[2000] flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" @click.self="emit('close')">
    <div class="bg-white dark:bg-slate-800 rounded-2xl w-full max-w-lg max-h-[90vh] flex flex-col shadow-2xl overflow-hidden transition-all animate-fade-in-up">
      
      <!-- Header -->
      <div class="p-5 border-b border-slate-100 dark:border-slate-700 flex justify-between items-center bg-slate-50 dark:bg-slate-800/50">
        <div>
            <h2 class="text-xl font-bold text-slate-800 dark:text-white">{{ placeName }}</h2>
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('comments.title') }}</p>
        </div>
        <button @click="emit('close')" class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors text-slate-500">
          ✕
        </button>
      </div>

      <!-- Scrollable Content -->
      <div class="flex-1 overflow-y-auto p-5 space-y-6">
        
        <!-- Add Comment Form -->
        <div class="bg-slate-50 dark:bg-slate-700/50 p-4 rounded-xl border border-slate-100 dark:border-slate-700">
            <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-200 mb-3 flex items-center gap-2">
                ✍️ {{ t('comments.share_experience') }}
            </h3>
            
            <div class="flex items-center gap-2 mb-3">
                <span class="text-sm text-slate-600 dark:text-slate-400">{{ t('comments.your_rating') }}</span>
                <div class="flex gap-1">
                    <button v-for="star in 5" :key="star" 
                        @click="newRating = star"
                        class="text-2xl transition-transform hover:scale-110 focus:outline-none"
                        :class="star <= newRating ? 'text-yellow-400' : 'text-slate-300 dark:text-slate-600'"
                    >
                        ★
                    </button>
                </div>
            </div>

            <textarea 
                v-model="newComment"
                :placeholder="t('comments.placeholder')" 
                class="w-full p-3 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none text-sm resize-none h-24"
            ></textarea>
            
            <button 
                @click="submitComment" 
                :disabled="isLoading || !newComment.trim()"
                class="w-full mt-3 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-50 disabled:cursor-not-allowed text-white py-2 rounded-lg font-medium transition-colors"
            >
                {{ isLoading ? t('comments.saving') : t('comments.submit') }}
            </button>
        </div>

        <!-- Comments List -->
        <div class="space-y-4">
            <div v-if="comments.length === 0" class="text-center py-8 text-slate-400">
                {{ t('comments.no_comments') }}
            </div>

            <div v-for="comment in comments" :key="comment.id" class="border-b border-slate-100 dark:border-slate-700/50 pb-4 last:border-0">
                <div class="flex justify-between items-start mb-1">
                    <div class="flex text-yellow-400 text-sm tracking-tighter">
                        <span v-for="n in 5" :key="n">{{ n <= comment.rating ? '★' : '☆' }}</span>
                    </div>
                    <span class="text-xs text-slate-400">{{ formatDate(comment.created_at) }}</span>
                </div>
                <p class="text-slate-700 dark:text-slate-300 text-sm whitespace-pre-wrap leading-relaxed">{{ comment.content }}</p>
            </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in-up {
  animation: fadeInUp 0.3s ease-out;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
