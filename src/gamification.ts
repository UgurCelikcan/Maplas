
export interface Rank {
    title: string;
    icon: string;
    minPoints: number;
    color: string;
}

export const RANKS: Rank[] = [
    { title: 'Yeni Kaşif', icon: '🌱', minPoints: 0, color: 'text-emerald-500' },
    { title: 'Yerel Rehber', icon: '🧭', minPoints: 100, color: 'text-blue-500' },
    { title: 'Rota Ustası', icon: '🗺️', minPoints: 500, color: 'text-purple-500' },
    { title: 'Efsane', icon: '👑', minPoints: 1000, color: 'text-yellow-500' }
];

export function getUserRank(points: number): Rank {
    const rank = RANKS.slice().reverse().find(r => points >= r.minPoints);
    return rank || RANKS[0] as Rank;
}

export function getNextRank(points: number): Rank | null {
    return RANKS.find(r => points < r.minPoints) || null;
}

export function getProgress(points: number): number {
    const currentRank = getUserRank(points);
    const nextRank = getNextRank(points);

    if (!nextRank) return 100; // Max level

    const totalGap = nextRank.minPoints - currentRank.minPoints;
    const currentProgress = points - currentRank.minPoints;
    
    return Math.min(Math.round((currentProgress / totalGap) * 100), 100);
}

export function getUserPoints(user: any): number {
    return user?.points ?? 0;
}
