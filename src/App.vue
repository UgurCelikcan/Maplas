<script setup lang="ts">
import { ref, provide, watch, onMounted } from 'vue';
import axios from 'axios';
import MapDisplay from './components/MapDisplay.vue';
import PlaceList from './components/PlaceList.vue';
import AddPlaceModal from './components/AddPlaceModal.vue';
import CommentsModal from './components/CommentsModal.vue';
import AuthModal from './components/AuthModal.vue';
import AdminDashboard from './components/AdminDashboard.vue';

interface Place {
  id: number;
  name: string;
  description: string;
  lat: number;
  lng: number;
  category: string;
  city: string;
  imageUrl?: string;
}

interface User {
  username: string;
  role: string;
}

const places = ref<Place[]>([]);
const selectedPlaceId = ref<number | null>(null);
const isDarkMode = ref(true);
const showModal = ref(false);
const showCommentsModal = ref(false);
const showAuthModal = ref(false);
const showAdminDashboard = ref(false);
const currentUser = ref<User | null>(null);
const commentPlace = ref<{id: number, name: string} | null>(null);
const editingPlace = ref<Place | undefined>(undefined);

// Axios Interceptor for Authorization Token
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

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

function checkAuth() {
    const token = localStorage.getItem('token');
    const userStr = localStorage.getItem('user');
    if (token && userStr) {
        currentUser.value = JSON.parse(userStr);
    }
}

onMounted(() => {
  checkAuth();
  fetchPlaces();
});

function handleLoginSuccess(user: User, token: string) {
    currentUser.value = user;
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(user));
    showAuthModal.value = false;
}

function handleLogout() {
    currentUser.value = null;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    showAdminDashboard.value = false;
}

function openAddModal() {
    // Optional: Require login to add places? For now, allowing public add (pending approval)
    // If we wanted to restrict: if (!currentUser.value) return showAuthModal.value = true;
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
        const response = await axios.put<Place>('http://localhost:8080/api/places', placeData);
        if (response.status === 200) {
            const index = places.value.findIndex(p => p.id === placeData.id);
            if (index !== -1) {
                places.value[index] = response.data;
            }
        }
    } else {
        const response = await axios.post<Place>('http://localhost:8080/api/places', placeData);
        if (response.status === 201) {
             alert('Yer eklendi ve yönetici onayına gönderildi.');
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
  } catch (error: any) {
    console.error('Error deleting place:', error);
    if (error.response && error.response.status === 403) {
        alert('Bu işlem için yetkiniz yok.');
    } else {
        alert('Yer silinirken bir hata oluştu.');
    }
  }
}

function handleSelectPlace(id: number) {
  selectedPlaceId.value = id;
}

function handleViewComments(id: number) {
  const place = places.value.find(p => p.id === id);
  if (place) {
    commentPlace.value = { id: place.id, name: place.name };
    showCommentsModal.value = true;
  }
}

function toggleTheme() {
  isDarkMode.value = !isDarkMode.value;
}

provide('isDarkMode', isDarkMode);

watch(isDarkMode, (newVal) => {
  if (newVal) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
}, { immediate: true });
</script>

<template>
  <div class="flex w-screen h-screen overflow-hidden bg-slate-50 text-slate-800 dark:bg-slate-900 dark:text-white transition-colors duration-300">
    <PlaceList 
      :places="places" 
      :selected-place-id="selectedPlaceId"
      :current-user="currentUser"
      @select-place="handleSelectPlace"
      @toggle-theme="toggleTheme"
      @add-click="openAddModal"
      @delete-place="handleDeletePlace"
      @edit-place="openEditModal"
      @open-auth="showAuthModal = true"
      @logout="handleLogout"
      @open-admin="showAdminDashboard = true"
    />
    <MapDisplay 
      :places="places" 
      :selected-place-id="selectedPlaceId" 
      @select-place="handleSelectPlace"
      @view-comments="handleViewComments"
    />

    <!-- Modals -->
    <AddPlaceModal 
      v-if="showModal" 
      :initial-data="editingPlace"
      @close="showModal = false" 
      @save-place="handleSavePlace" 
    />

    <CommentsModal
      v-if="showCommentsModal && commentPlace"
      :place-id="commentPlace.id"
      :place-name="commentPlace.name"
      @close="showCommentsModal = false"
    />

    <AuthModal
      v-if="showAuthModal"
      @close="showAuthModal = false"
      @login-success="handleLoginSuccess"
    />

    <AdminDashboard
      v-if="showAdminDashboard"
      @close="showAdminDashboard = false"
      @place-approved="fetchPlaces"
    />
  </div>
</template>