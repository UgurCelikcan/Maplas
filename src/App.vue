<script setup lang="ts">
import { ref, provide, watch, onMounted } from 'vue';
import axios from 'axios';
import MapDisplay from './components/MapDisplay.vue';
import PlaceList from './components/PlaceList.vue';
import AddPlaceModal from './components/AddPlaceModal.vue';

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
const showModal = ref(false);
const editingPlace = ref<Place | undefined>(undefined);

const fetchPlaces = async () => {
  try {
    const response = await axios.get<Place[]>('http://localhost:8080/api/places');
    if (response.data) {
        places.value = response.data;
    }
  } catch (error) {
    console.error('Error fetching places:', error);
  }
};

onMounted(() => {
  fetchPlaces();
});

function openAddModal() {
    editingPlace.value = undefined;
    showModal.value = true;
}

function openEditModal(place: Place) {
    editingPlace.value = place;
    showModal.value = true;
}

async function handleSavePlace(placeData: Place) {
  try {
    if (placeData.id) {
        // Update existing
        const response = await axios.put<Place>('http://localhost:8080/api/places', placeData);
        if (response.status === 200) {
            const index = places.value.findIndex(p => p.id === placeData.id);
            if (index !== -1) {
                places.value[index] = response.data;
            }
        }
    } else {
        // Create new
        const response = await axios.post<Place>('http://localhost:8080/api/places', placeData);
        if (response.status === 201) {
            places.value.unshift(response.data);
        }
    }
    showModal.value = false;
  } catch (error) {
    console.error('Error saving place:', error);
    alert('İşlem sırasında bir hata oluştu.');
  }
}

async function handleDeletePlace(id: number) {
  if (!confirm('Bu yeri silmek istediğinize emin misiniz?')) return;

  try {
    await axios.delete(`http://localhost:8080/api/places?id=${id}`);
    places.value = places.value.filter(p => p.id !== id);
    if (selectedPlaceId.value === id) {
        selectedPlaceId.value = null;
    }
  } catch (error) {
    console.error('Error deleting place:', error);
    alert('Yer silinirken bir hata oluştu.');
  }
}

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
      @add-click="openAddModal"
      @delete-place="handleDeletePlace"
      @edit-place="openEditModal"
    />
    <MapDisplay 
      :places="places" 
      :selected-place-id="selectedPlaceId" 
      @select-place="handleSelectPlace"
    />

    <!-- Modal -->
    <AddPlaceModal 
      v-if="showModal" 
      :initial-data="editingPlace"
      @close="showModal = false" 
      @save-place="handleSavePlace" 
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