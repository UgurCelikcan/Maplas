import { ref, shallowRef, type Ref } from 'vue';
import L from 'leaflet';
import { useI18n } from 'vue-i18n';

export function useMapRouting(map: Ref<L.Map | null>, transportMode: Ref<string>) {
    const { t } = useI18n();
    const routingControl = shallowRef<any>(null);
    const routeInfo = ref<{ roads: string[], totalDistance: string, totalTime: string } | null>(null);
    const routeWaypoints = ref<Array<{lat: number, lng: number, name: string}>>([]);

    function clearRoute() {
        routeWaypoints.value = [];
        if (routingControl.value) {
            map.value?.removeControl(routingControl.value);
            routingControl.value = null;
        }
        routeInfo.value = null;
    }

    function removeWaypoint(index: number) {
        routeWaypoints.value.splice(index, 1);
        updateRouteDisplay();
    }

    function moveWaypoint(index: number, direction: 'up' | 'down') {
        if (direction === 'up' && index > 0) {
            const temp = routeWaypoints.value[index]!;
            routeWaypoints.value[index] = routeWaypoints.value[index - 1]!;
            routeWaypoints.value[index - 1] = temp;
        } else if (direction === 'down' && index < routeWaypoints.value.length - 1) {
            const temp = routeWaypoints.value[index]!;
            routeWaypoints.value[index] = routeWaypoints.value[index + 1]!;
            routeWaypoints.value[index + 1] = temp;
        }
        updateRouteDisplay();
    }

    function updateRouteDisplay() {
        if (!map.value || routeWaypoints.value.length < 2) {
            if (routingControl.value) {
                map.value?.removeControl(routingControl.value);
                routingControl.value = null;
                routeInfo.value = null;
            }
            return;
        }

        if (routingControl.value) {
            map.value?.removeControl(routingControl.value);
        }

        // @ts-ignore
        if (L.Routing) {
            let profile = 'car';
            if (transportMode.value === 'bicycle') profile = 'bike';
            if (transportMode.value === 'foot' || transportMode.value === 'foot-walking') profile = 'foot';
            if (transportMode.value === 'driving-car') profile = 'car';
            if (transportMode.value === 'cycling-regular') profile = 'bike';

            // @ts-ignore
            const control = L.Routing.control({
                waypoints: routeWaypoints.value.map(wp => L.latLng(wp.lat, wp.lng)),
                router: L.Routing.osrmv1({
                    serviceUrl: 'https://router.project-osrm.org/route/v1',
                    profile: profile
                }),
                routeWhileDragging: false,
                addWaypoints: false,
                fitSelectedRoutes: true,
                showAlternatives: false,
                show: false,
                lineOptions: {
                    styles: [{ 
                        color: (profile === 'bike') ? '#3b82f6' : (profile === 'foot' ? '#f97316' : '#42b883'), 
                        opacity: 0.8, 
                        weight: 6 
                    }],
                    extendToWaypoints: true,
                    missingRouteTolerance: 0
                },
                // @ts-ignore
                createMarker: () => null
            });

            control.addTo(map.value!);
            routingControl.value = control;

            control.on('routesfound', function(e: any) {
                const routes = e.routes;
                const summary = routes[0].summary;
                const distanceMeters = summary.totalDistance;
                
                let speedMetersPerMin = 833; 
                if (profile === 'bike') speedMetersPerMin = 250;
                if (profile === 'foot') speedMetersPerMin = 83;

                let totalMinutes = Math.round(summary.totalTime / 60);
                if (profile !== 'car') {
                    totalMinutes = Math.round(distanceMeters / speedMetersPerMin);
                }

                const hours = Math.floor(totalMinutes / 60);
                const mins = totalMinutes % 60;
                const timeStr = hours > 0 ? `${hours} ${t('map.hours')} ${mins} ${t('map.minutes')}` : `${mins} ${t('map.minutes')}`;
                
                const distKm = (distanceMeters / 1000).toFixed(1);
                const instructions = routes[0].instructions;
                const roads: string[] = instructions
                    .map((i: any) => i.road)
                    .filter((r: string) => {
                        if (!r || r.trim().length === 0 || r === '{road}') return false;
                        const name = r.toLocaleUpperCase('tr-TR');
                        return name.match(/^(O|D|E)\s?-?\s?\d+/) || name.includes('OTOYOL') || name.includes('ÇEVRE YOLU');
                    })
                    .reduce((acc: string[], curr: string) => {
                        if (acc.length === 0 || acc[acc.length - 1] !== curr) acc.push(curr);
                        return acc;
                    }, []);

                routeInfo.value = {
                    roads: roads.length > 0 ? roads : [t('map.main_roads')],
                    totalDistance: `${distKm} km`,
                    totalTime: timeStr
                };
            });
        }
    }

    function shareRoute() {
        if (routeWaypoints.value.length < 2) return;
        const data = routeWaypoints.value.map(wp => ({ la: wp.lat, lo: wp.lng, n: wp.name }));
        const encodedData = btoa(encodeURIComponent(JSON.stringify(data)));
        const url = new URL(window.location.href);
        url.searchParams.set('route', encodedData);
        navigator.clipboard.writeText(url.toString()).then(() => {
            alert(t('map.route_copied') || "Rota bağlantısı kopyalandı!");
        });
    }

    return {
        routeWaypoints,
        routeInfo,
        routingControl,
        clearRoute,
        removeWaypoint,
        moveWaypoint,
        updateRouteDisplay,
        shareRoute
    };
}
