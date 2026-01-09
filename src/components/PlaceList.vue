<script setup lang="ts">
import { ref, computed, inject } from 'vue';

interface Place {
  id: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
}

const props = defineProps<{
  places: Place[];
  selectedPlaceId: number | null;
}>();

const emit = defineEmits<{
  (e: 'select-place', id: number): void;
  (e: 'toggle-theme'): void;
  (e: 'add-click'): void;
  (e: 'delete-place', id: number): void;
  (e: 'edit-place', place: any): void;
}>();

const isDarkMode = inject('isDarkMode', ref(true));

const searchQuery = ref('');
const selectedCategory = ref('Tümü');

// Kategorileri dinamik olarak çıkar
const categories = computed(() => {
  const uniqueCategories = new Set(props.places.map(p => p.category));
  return ['Tümü', ...Array.from(uniqueCategories)];
});

// Filtreleme mantığı
const filteredPlaces = computed(() => {
  return props.places.filter(place => {
    const matchesSearch = place.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          place.city.toLowerCase().includes(searchQuery.value.toLowerCase());
    const matchesCategory = selectedCategory.value === 'Tümü' || place.category === selectedCategory.value;
    return matchesSearch && matchesCategory;
  });
});

function onSelect(id: number) {
  emit('select-place', id);
}

function getCategoryEmoji(category: string) {
  switch (category) {
    case 'Tarihi': return '🏛️';
    case 'Doğa': return '🌲';
    case 'Plaj': return '🏖️';
    case 'Müze': return '🎨';
    case 'Antik Kent': return '🏺';
    case 'Alışveriş': return '🛍️';
    case 'Manzara': return '🌄';
    default: return '📍';
  }
}
</script>

<template>
  <div class="sidebar" :class="{ 'light-sidebar': !isDarkMode }">
    <div class="header">
      <div class="brand-row">
        <div class="brand">
            <span class="logo-icon">🗺️</span>
            <div>
                <h2>Türkiye Rehberi</h2>
                <p>Keşfedilecek Rotalar</p>
            </div>
        </div>
        <div class="header-actions">
            <button class="action-btn" @click="$emit('add-click')" title="Yeni Yer Ekle">
                ➕
            </button>
            <button class="theme-toggle" @click="$emit('toggle-theme')" :title="isDarkMode ? 'Aydınlık Mod' : 'Karanlık Mod'">
                {{ isDarkMode ? '☀️' : '🌙' }}
            </button>
        </div>
      </div>
      
      <div class="filters">
        <div class="search-wrapper">
            <span class="search-icon">🔍</span>
            <input 
            v-model="searchQuery" 
            type="text" 
            placeholder="Şehir veya yer ara..." 
            class="search-input"
            />
        </div>
        
        <div class="category-pills">
          <button 
            v-for="cat in categories" 
            :key="cat"
            :class="{ active: selectedCategory === cat }"
            @click="selectedCategory = cat"
          >
            {{ cat }}
          </button>
        </div>
      </div>
    </div>

    <div class="place-list-container">
      <ul v-if="filteredPlaces.length > 0" class="place-list">
        <li 
          v-for="place in filteredPlaces" 
          :key="place.id" 
          :class="{ active: place.id === selectedPlaceId }"
          @click="onSelect(place.id)"
        >
          <div class="card-content">
            <div class="card-header">
                <div class="title-row">
                    <div class="title-left">
                        <h3>{{ place.name }}</h3>
                        <span class="category-icon" :title="place.category">{{ getCategoryEmoji(place.category) }}</span>
                    </div>
                    <div class="actions">
                        <button class="icon-btn edit-btn" @click.stop="$emit('edit-place', place)" title="Düzenle">
                            ✏️
                        </button>
                        <button class="icon-btn delete-btn" @click.stop="$emit('delete-place', place.id)" title="Sil">
                            🗑️
                        </button>
                    </div>
                </div>
                <div class="meta-row">
                    <span class="city-badge">📍 {{ place.city }}</span>
                    <span class="category-text">{{ place.category }}</span>
                </div>
            </div>
            <p class="desc">{{ place.description }}</p>
          </div>
        </li>
      </ul>
      <div v-else class="no-results">
        <p>Aradığınız kriterlere uygun yer bulunamadı.</p>
        <button class="clear-btn" @click="searchQuery = ''; selectedCategory = 'Tümü'">Filtreleri Temizle</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  width: 420px;
  background-color: #18181b; /* Zinc 900 */
  color: #f4f4f5; /* Zinc 100 */
  display: flex;
  flex-direction: column;
  height: 100%;
  border-right: 1px solid #27272a; /* Zinc 800 */
  box-shadow: 4px 0 24px rgba(0,0,0,0.4);
  z-index: 1000;
  font-family: 'Inter', sans-serif;
  transition: background-color 0.3s, border-color 0.3s, color 0.3s;
}

/* Light Mode Overrides */
.sidebar.light-sidebar {
    background-color: #ffffff;
    color: #1e293b;
    border-right-color: #e2e8f0;
    box-shadow: 4px 0 24px rgba(0,0,0,0.05);
}

.header {
  padding: 24px;
  background-color: #27272a;
  border-bottom: 1px solid #3f3f46;
  transition: background-color 0.3s, border-color 0.3s;
}

.light-sidebar .header {
    background-color: #f8fafc;
    border-bottom-color: #e2e8f0;
}

.brand-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
}

.brand {
    display: flex;
    align-items: center;
    gap: 12px;
}

.header-actions {
    display: flex;
    gap: 8px;
}

.action-btn, .theme-toggle {
    background: transparent;
    border: 1px solid #3f3f46;
    color: #fff;
    cursor: pointer;
    font-size: 1.1rem;
    padding: 8px;
    border-radius: 8px;
    line-height: 1;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
}

.light-sidebar .action-btn, .light-sidebar .theme-toggle {
    border-color: #cbd5e1;
    color: #1e293b;
}

.action-btn:hover, .theme-toggle:hover {
    background-color: rgba(255,255,255,0.1);
    border-color: #42b883;
}

.light-sidebar .action-btn:hover, .light-sidebar .theme-toggle:hover {
    background-color: #f1f5f9;
    border-color: #42b883;
}

.logo-icon {
    font-size: 2rem;
}

.header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #ffffff;
  font-weight: 700;
  letter-spacing: -0.025em;
}

.light-sidebar .header h2 {
    color: #0f172a;
}

.header p {
  margin: 2px 0 0;
  opacity: 0.6;
  font-size: 0.85rem;
}

.filters {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.search-wrapper {
    position: relative;
}

.search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    opacity: 0.5;
    font-size: 0.9rem;
}

.search-input {
  width: 100%;
  padding: 12px 12px 12px 38px;
  border-radius: 12px;
  border: 1px solid #3f3f46;
  background-color: #18181b;
  color: #fff;
  font-size: 0.95rem;
  transition: all 0.2s;
  box-sizing: border-box;
}

.light-sidebar .search-input {
    background-color: #ffffff;
    border-color: #cbd5e1;
    color: #0f172a;
}

.search-input:focus {
  outline: none;
  border-color: #42b883;
  box-shadow: 0 0 0 2px rgba(66, 184, 131, 0.2);
}

.category-pills {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 8px;
  scrollbar-width: none; /* Firefox */
}
.category-pills::-webkit-scrollbar {
    display: none; /* Chrome/Safari */
}

.category-pills button {
  background: #3f3f46;
  border: 1px solid transparent;
  color: #d4d4d8;
  padding: 8px 16px;
  border-radius: 999px;
  cursor: pointer;
  white-space: nowrap;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.2s;
}

.light-sidebar .category-pills button {
    background: #f1f5f9;
    color: #64748b;
}

.category-pills button:hover {
  background: #52525b;
  color: #fff;
}

.light-sidebar .category-pills button:hover {
    background: #e2e8f0;
    color: #0f172a;
}

.category-pills button.active {
  background: #42b883;
  color: #0f172a;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(66, 184, 131, 0.3);
}

.place-list-container {
  flex-grow: 1;
  overflow-y: auto;
  background-color: #18181b;
  padding: 16px;
  transition: background-color 0.3s;
}

.light-sidebar .place-list-container {
    background-color: #f8fafc;
}

.place-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.place-list li {
  background-color: #27272a;
  border-radius: 16px;
  border: 1px solid #3f3f46;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.light-sidebar .place-list li {
    background-color: #ffffff;
    border-color: #e2e8f0;
    box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.place-list li:hover {
  transform: translateY(-2px);
  background-color: #3f3f46;
  border-color: #52525b;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3);
}

.light-sidebar .place-list li:hover {
    background-color: #f8fafc;
    border-color: #cbd5e1;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05);
}

.place-list li.active {
  background-color: #27272a;
  border-color: #42b883;
  box-shadow: 0 0 0 2px rgba(66, 184, 131, 0.4);
}

.light-sidebar .place-list li.active {
    background-color: #f0fdf4;
    border-color: #42b883;
}

.card-content {
    padding: 16px;
}

.card-header {
    margin-bottom: 12px;
}

.title-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 6px;
}

.title-left {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-grow: 1;
    overflow: hidden; /* Prevent text overflow if title is long */
}

.actions {
    display: flex;
    gap: 4px;
    opacity: 0;
    transition: opacity 0.2s;
}

.place-list li:hover .actions {
    opacity: 1;
}

.icon-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s;
}

.delete-btn:hover {
    background-color: rgba(239, 68, 68, 0.15);
}

.edit-btn:hover {
    background-color: rgba(66, 184, 131, 0.15);
}

.place-list h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #fff;
  font-weight: 600;
  line-height: 1.3;
}

.light-sidebar .place-list h3 {
    color: #0f172a;
}

.category-icon {
    font-size: 1.2rem;
    background: rgba(255,255,255,0.05);
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    margin-left: 8px;
    flex-shrink: 0;
}

.light-sidebar .category-icon {
    background: #f1f5f9;
}

.meta-row {
    display: flex;
    align-items: center;
    gap: 10px;
}

.city-badge {
  font-size: 0.75rem;
  color: #a1a1aa;
  display: flex;
  align-items: center;
  font-weight: 500;
}

.light-sidebar .city-badge {
    color: #64748b;
}

.category-text {
  font-size: 0.7rem;
  background-color: rgba(66, 184, 131, 0.1);
  color: #42b883;
  padding: 2px 8px;
  border-radius: 6px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.desc {
  margin: 0;
  font-size: 0.85rem;
  color: #a1a1aa;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.light-sidebar .desc {
    color: #64748b;
}

.no-results {
  padding: 60px 20px;
  text-align: center;
  color: #71717a;
}

.clear-btn {
    margin-top: 16px;
    background: transparent;
    border: 1px solid #52525b;
    color: #a1a1aa;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
}
.clear-btn:hover {
    border-color: #a1a1aa;
    color: #fff;
}
</style>