export function getLocalizedContent(content: any, locale: string, fallback: string = 'tr'): string {
    if (!content) return '';
    if (typeof content === 'string') return content; // Backward compatibility
    
    return content[locale] || content[fallback] || content['en'] || Object.values(content)[0] || '';
}
