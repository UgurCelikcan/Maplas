<script setup lang="ts">
import { ref, onMounted } from 'vue';

interface Place {
  id?: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
}

const props = defineProps<{
  initialData?: Place;
}>();

const emit = defineEmits<{
  (e: 'save-place', place: Place): void;
  (e: 'close'): void;
}>();

const form = ref<Place>({
  name: '',
  description: '',
  lat: 0,
  lng: 0,
  category: '',
  city: ''
});

onMounted(() => {
    if (props.initialData) {
        // Clone to avoid mutating prop directly before save
        form.value = { ...props.initialData };
    }
});

function handleSubmit() {
  emit('save-place', form.value);
}
</script>

<template>
  <div class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ initialData ? 'Yeri Düzenle' : 'Yeni Yer Ekle' }}</h3>
        <button class="close-btn" @click="$emit('close')">✕</button>
      </div>
      
      <form @submit.prevent="handleSubmit" class="add-place-form">
        <div class="form-group">
          <label>Yer Adı</label>
          <input v-model="form.name" required placeholder="Örn: Galata Kulesi" />
        </div>

        <div class="form-row">
            <div class="form-group">
                <label>Şehir</label>
                <input v-model="form.city" required placeholder="Örn: İstanbul" />
            </div>
            <div class="form-group">
                <label>Kategori</label>
                <select v-model="form.category" required>
                    <option value="" disabled>Seçiniz</option>
                    <option value="Tarihi">Tarihi</option>
                    <option value="Doğa">Doğa</option>
                    <option value="Plaj">Plaj</option>
                    <option value="Müze">Müze</option>
                    <option value="Antik Kent">Antik Kent</option>
                    <option value="Manzara">Manzara</option>
                    <option value="Diğer">Diğer</option>
                </select>
            </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Enlem (Lat)</label>
            <input v-model.number="form.lat" type="number" step="any" required />
          </div>
          <div class="form-group">
            <label>Boylam (Lng)</label>
            <input v-model.number="form.lng" type="number" step="any" required />
          </div>
        </div>

        <div class="form-group">
          <label>Açıklama</label>
          <textarea v-model="form.description" rows="3" required placeholder="Kısa bir açıklama..."></textarea>
        </div>

        <div class="form-actions">
          <button type="button" class="btn-cancel" @click="$emit('close')">İptal</button>
          <button type="submit" class="btn-submit">{{ initialData ? 'Güncelle' : 'Kaydet' }}</button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: #27272a;
  padding: 24px;
  border-radius: 16px;
  width: 100%;
  max-width: 450px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  color: #fff;
  border: 1px solid #3f3f46;
}

.light-mode .modal-content { /* Parent class needs to be checked or passed down */
    background: #ffffff;
    color: #1e293b;
    border-color: #e2e8f0;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: #a1a1aa;
  cursor: pointer;
  font-size: 1.2rem;
}

.add-place-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-row {
    display: flex;
    gap: 12px;
}
.form-row .form-group {
    flex: 1;
}

label {
  font-size: 0.85rem;
  font-weight: 500;
  color: #a1a1aa;
}

input, textarea, select {
  padding: 10px;
  border-radius: 8px;
  border: 1px solid #3f3f46;
  background: #18181b;
  color: #fff;
  font-family: inherit;
  font-size: 0.95rem;
}
input:focus, textarea:focus, select:focus {
    outline: none;
    border-color: #42b883;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}

.btn-cancel {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid #3f3f46;
  background: transparent;
  color: #e4e4e7;
  cursor: pointer;
}

.btn-submit {
  padding: 8px 20px;
  border-radius: 8px;
  border: none;
  background: #42b883;
  color: #0f172a;
  font-weight: 600;
  cursor: pointer;
}
.btn-submit:hover {
    background: #34d399;
}
</style>