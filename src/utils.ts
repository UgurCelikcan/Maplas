export function getLocalizedContent(content: any, locale: string, fallback: string = 'tr'): string {
    if (!content) return '';
    if (typeof content === 'string') return content; // Backward compatibility
    
    return content[locale] || content[fallback] || content['en'] || Object.values(content)[0] || '';
}

export function getCategoryColor(category: string): string {
    switch (category) {
      case 'Tarihi': return '#f87171';
      case 'Doğa': return '#4ade80';
      case 'Plaj': return '#60a5fa';
      case 'Müze': return '#c084fc';
      case 'Antik Kent': return '#fbbf24';
      case 'Alışveriş': return '#e91e63';
      default: return '#94a3b8';
    }
}

export function getCategoryEmoji(category: string): string {
    switch (category) {
      case 'Tarihi': return '🏛️';
      case 'Doğa': return '🌲';
      case 'Plaj': return '🏖️';
      case 'Müze': return '🖼️';
      case 'Antik Kent': return '🗿';
      case 'Alışveriş': return '🛍️';
      case 'Eğlence': return '🎢';
      default: return '📍';
    }
}

export function getDistance(lat1: number, lng1: number, lat2: number, lng2: number): number {
    const R = 6371; // km
    const dLat = (lat2 - lat1) * Math.PI / 180;
    const dLng = (lng2 - lng1) * Math.PI / 180;
    const a = Math.sin(dLat/2) * Math.sin(dLat/2) +
              Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) * 
              Math.sin(dLng/2) * Math.sin(dLng/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
    return R * c;
}

