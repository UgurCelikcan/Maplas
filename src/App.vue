<script setup lang="ts">
import { ref, provide, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import api, { getNearbyPlaces } from './api';
import MapDisplay from './components/MapDisplay.vue';
import PlaceList from './components/PlaceList.vue';
import AddPlaceModal from './components/AddPlaceModal.vue';
import CommentsModal from './components/CommentsModal.vue';
import AuthModal from './components/AuthModal.vue';
import AdminDashboard from './components/AdminDashboard.vue';
import AboutModal from './components/AboutModal.vue';
import UserProfile from './components/UserProfile.vue';

const { t } = useI18n();

interface Place {
  id?: number;
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
const showAboutModal = ref(false);
const showProfileModal = ref(false);
const currentUser = ref<User | null>(null);
const commentPlace = ref<{id: number, name: string} | null>(null);
const editingPlace = ref<Place | undefined>(undefined);
const isSidebarOpen = ref(false);

const fetchPlaces = async () => {
  try {
    const response = await api.get<Place[]>('/places');
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
        const response = await api.put<Place>('/places', placeData);
        if (response.status === 200) {
            const index = places.value.findIndex(p => p.id === placeData.id);
            if (index !== -1) {
                places.value[index] = response.data;
            }
        }
    } else {
        const response = await api.post<Place>('/places', placeData);
        if (response.status === 201) {
             alert(t('place.pending_approval'));
        }
    }
    showModal.value = false;
  } catch (error) {
    console.error('Error saving place:', error);
    alert(t('place.save_error'));
  }
}

async function handleDeletePlace(id: number) {
  if (!confirm(t('place.delete_confirm'))) return;

  try {
    await api.delete(`/places?id=${id}`);
    places.value = places.value.filter(p => p.id !== id);
    if (selectedPlaceId.value === id) {
        selectedPlaceId.value = null;
    }
  } catch (error: any) {
    console.error('Error deleting place:', error);
    if (error.response && error.response.status === 403) {
        alert(t('place.no_permission'));
    } else {
        alert(t('place.delete_error'));
    }
  }
}

function handleSelectPlace(id: number) {
  selectedPlaceId.value = id;
}

function handleViewComments(id: number) {
  const place = places.value.find(p => p.id === id);
  if (place) {
    commentPlace.value = { id: place.id as number, name: place.name };
    showCommentsModal.value = true;
  }
}

async function handleNearbySearch() {
    if (!navigator.geolocation) {
        alert(t('map.location_not_found'));
        return;
    }

    navigator.geolocation.getCurrentPosition(async (pos) => {
        try {
            const { latitude, longitude } = pos.coords;
            const nearbyPlaces = await getNearbyPlaces(latitude, longitude, 10); // 10km radius
            places.value = nearbyPlaces || [];
            
            if (places.value.length === 0) {
                alert(t('ui.no_results'));
            } else {
                // Optionally fly to user location on map if we could access map instance here, 
                // but MapDisplay handles its own user location logic.
                // We could emit an event to MapDisplay via a prop change or ref, but let's keep it simple.
            }
        } catch (error) {
            console.error('Error searching nearby:', error);
            alert(t('common.error'));
        }
    }, (err) => {
        console.error('Geolocation error:', err);
        alert(t('map.location_not_found'));
    });
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
  <div class="flex w-screen h-screen overflow-hidden bg-slate-50 text-slate-800 dark:bg-slate-900 dark:text-white transition-colors duration-300 relative">
    
    <!-- Mobile Menu Toggle Button -->
    <button 
        class="absolute top-4 left-4 z-[999] md:hidden w-11 h-11 flex items-center justify-center bg-white dark:bg-zinc-800 border border-slate-300 dark:border-zinc-700 rounded-lg text-2xl shadow-lg cursor-pointer transition-colors hover:bg-slate-50 dark:hover:bg-zinc-700" 
        @click="isSidebarOpen = true"
    >
        ☰
    </button>

    <PlaceList 
      :places="places" 
      :selected-place-id="selectedPlaceId"
      :current-user="currentUser"
      :is-open="isSidebarOpen"
      @select-place="handleSelectPlace"
      @toggle-theme="toggleTheme"
      @add-click="openAddModal"
      @delete-place="handleDeletePlace"
      @edit-place="openEditModal"
      @open-auth="showAuthModal = true"
      @logout="handleLogout"
      @open-admin="showAdminDashboard = true"
      @open-profile="showProfileModal = true"
      @close-sidebar="isSidebarOpen = false"
      @open-about="showAboutModal = true"
      @search-nearby="handleNearbySearch"
    />
    <MapDisplay 
      :places="places" 
      :selected-place-id="selectedPlaceId" 
      @select-place="handleSelectPlace"
      @view-comments="handleViewComments"
    />


    <!-- Modals -->
    <UserProfile
      v-if="showProfileModal"
      @close="showProfileModal = false"
    />

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

    <AboutModal
      v-if="showAboutModal"
      @close="showAboutModal = false"
    />
  </div>
</template>