<script setup lang="ts">
import { ref, provide, watch, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import api, { getNearbyPlaces, setFavoriteStatus } from './api';
import MapDisplay from './components/MapDisplay.vue';
import PlaceList from './components/PlaceList.vue';
import AddPlaceModal from './components/AddPlaceModal.vue';
import CommentsModal from './components/CommentsModal.vue';
import AuthModal from './components/AuthModal.vue';
import AdminDashboard from './components/AdminDashboard.vue';
import AboutModal from './components/AboutModal.vue';
import UserProfile from './components/UserProfile.vue';
import LeaderboardModal from './components/LeaderboardModal.vue';
import SmartPlannerModal from './components/SmartPlannerModal.vue';
import { getLocalizedContent } from './utils';

const { t, locale } = useI18n();

interface Place {
  id?: number;
  name: Record<string, string>; // Changed to Record to match updated PlaceList/utils
  description: Record<string, string>;
  lat: number;
  lng: number;
  category: string;
  city: string;
  imageUrl?: string;
  is_favorite?: boolean;
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
const showLeaderboardModal = ref(false);
const showSmartPlannerModal = ref(false);
const currentUser = ref<User | null>(null);
const commentPlace = ref<{id: number, name: string} | null>(null);
const editingPlace = ref<Place | undefined>(undefined);
const isSidebarOpen = ref(false);
const mapDisplayRef = ref<any>(null);
const currentUserLocation = ref<{lat: number, lng: number} | null>(null);

// Filter States
const searchQuery = ref('');
const selectedCategories = ref<string[]>([]);
const selectedCity = ref<string>('');
const showFavoritesOnly = ref(false);

const filteredPlaces = computed(() => {
  return places.value.filter(place => {
    const name = getLocalizedContent(place.name, locale.value);
    const matchesSearch = name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                          place.city.toLowerCase().includes(searchQuery.value.toLowerCase());
    
    const matchesCategory = selectedCategories.value.length === 0 || selectedCategories.value.includes(place.category);
    
    const matchesCity = selectedCity.value === '' || place.city === selectedCity.value;

    const matchesFavorites = !showFavoritesOnly.value || place.is_favorite;

    return matchesSearch && matchesCategory && matchesCity && matchesFavorites;
  });
});

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
  
  // Auto-theme based on time
  const hour = new Date().getHours();
  if (hour >= 19 || hour < 7) {
      isDarkMode.value = true;
  } else {
      isDarkMode.value = false;
  }

  // Listen for global auth errors (token expired)
  window.addEventListener('auth-error', () => {
      if (currentUser.value) {
          handleLogout();
          showAuthModal.value = true;
          alert("Oturum süreniz doldu. Lütfen tekrar giriş yapın.");
      }
  });
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
             alert(`🎉 ${t('place.pending_approval')}\n\n🌟 +50 XP Kazandın! (Onaylanınca hesabına işlenecek)`);
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

function handleSmartRoute(places: Place[]) {
    if (mapDisplayRef.value) {
        mapDisplayRef.value.setRoute(places);
    }
    showSmartPlannerModal.value = false;
    isSidebarOpen.value = false; // Close sidebar on mobile to show map
}

async function handleToggleFavorite(id: number) {
  if (!currentUser.value) {
    showAuthModal.value = true;
    return;
  }

  const index = places.value.findIndex(p => p.id === id);
  if (index !== -1) {
    const place = places.value[index]!;
    const newStatus = !place.is_favorite;
    
    // Optimistic update
    // We create a new object reference to trigger reactivity
    places.value[index] = { ...place, is_favorite: newStatus } as Place;
    
    // Force reactivity for deep watchers if any (Vue sometimes misses array item updates if not done carefully)
    places.value = [...places.value];

    try {
      await setFavoriteStatus(id, newStatus);
    } catch (error: any) {
      console.error('Error toggling favorite:', error);
      // Revert on error
      places.value[index] = { ...place, is_favorite: !newStatus } as Place;
      places.value = [...places.value];
      
      const errorMessage = error.response?.data || error.message || "Bilinmeyen hata";
      alert(`Favori işlemi başarısız oldu:\n${errorMessage}`);
    }
  }
}

function handleViewComments(id: number) {
  const place = places.value.find(p => p.id === id);
  if (place) {
    // Correctly handle the name if it's a Record<string, string> or string
    const placeName = typeof place.name === 'string' ? place.name : getLocalizedContent(place.name, locale.value);
    commentPlace.value = { id: place.id as number, name: placeName };
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
            
            if (nearbyPlaces && nearbyPlaces.length > 0) {
                places.value = nearbyPlaces;
            } else {
                alert(t('ui.no_results'));
                // If no results, reload all places to avoid empty screen
                fetchPlaces();
            }
        } catch (error) {
            console.error('Error searching nearby:', error);
            alert(t('common.error'));
            fetchPlaces(); // Fallback
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
      :visible-places="filteredPlaces"
      v-model:searchQuery="searchQuery"
      v-model:selectedCategories="selectedCategories"
      v-model:selectedCity="selectedCity"
      v-model:showFavoritesOnly="showFavoritesOnly"
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
      @open-leaderboard="showLeaderboardModal = true"
      @open-smart-planner="showSmartPlannerModal = true"
      @close-sidebar="isSidebarOpen = false"
      @open-about="showAboutModal = true"
      @search-nearby="handleNearbySearch"
      @toggle-favorite="handleToggleFavorite"
    />
    <MapDisplay 
      ref="mapDisplayRef"
      :places="filteredPlaces" 
      :selected-place-id="selectedPlaceId" 
      @select-place="handleSelectPlace"
      @view-comments="handleViewComments"
      @toggle-favorite="handleToggleFavorite"
      @location-updated="(loc) => currentUserLocation = loc"
    />


    <!-- Modals -->
    <UserProfile
      v-if="showProfileModal"
      @close="showProfileModal = false"
    />
    
    <LeaderboardModal
      v-if="showLeaderboardModal"
      @close="showLeaderboardModal = false"
    />

    <SmartPlannerModal
      v-if="showSmartPlannerModal"
      :places="places"
      :user-location="currentUserLocation"
      @close="showSmartPlannerModal = false"
      @create-route="handleSmartRoute"
    />

    <AddPlaceModal 
      v-if="showModal" 
      :initial-data="editingPlace as any"
      @close="showModal = false" 
      @save-place="(p: any) => handleSavePlace(p)" 
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