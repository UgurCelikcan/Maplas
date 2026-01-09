<script setup lang="ts">
import { ref, provide, watch, onMounted } from 'vue';
import axios from 'axios';
import MapDisplay from './components/MapDisplay.vue';
import PlaceList from './components/PlaceList.vue';

interface Place {
  id: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
}

const places = ref<Place[]>([]);
const selectedPlaceId = ref<number | null>(null);
const isDarkMode = ref(true);

onMounted(async () => {
  try {
    const response = await axios.get<Place[]>('http://localhost:8080/api/places');
    places.value = response.data;
  } catch (error) {
    console.error('Error fetching places:', error);
  }
});

function handleSelectPlace(id: number) {
  selectedPlaceId.value = id;
}

function toggleTheme() {
  isDarkMode.value = !isDarkMode.value;
}

// Provide theme to children
provide('isDarkMode', isDarkMode);

// Update body class for global styles if needed
watch(isDarkMode, (newVal) => {
  if (newVal) {
    document.body.classList.add('dark-theme');
    document.body.classList.remove('light-theme');
  } else {
    document.body.classList.add('light-theme');
    document.body.classList.remove('dark-theme');
  }
}, { immediate: true });
</script>

<template>
  <div class="app-container" :class="{ 'light-mode': !isDarkMode }">
    <PlaceList 
      :places="places" 
      :selected-place-id="selectedPlaceId" 
      @select-place="handleSelectPlace"
      @toggle-theme="toggleTheme"
    />
    <MapDisplay 
      :places="places" 
      :selected-place-id="selectedPlaceId" 
      @select-place="handleSelectPlace"
    />
  </div>
</template>

<style scoped>
.app-container {
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: #0f172a; /* Default dark bg */
  color: #fff;
  transition: background-color 0.3s, color 0.3s;
}

.app-container.light-mode {
  background-color: #f8fafc;
  color: #1e293b;
}
</style>