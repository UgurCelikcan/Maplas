import { createI18n } from 'vue-i18n';
import tr from './locales/tr.json';
import en from './locales/en.json';

// Turkish (default) and English (fallback) are bundled; the other languages
// are downloaded only when someone picks them.
const localeLoaders = import.meta.glob<{ default: Record<string, unknown> }>('./locales/*.json');

const i18n = createI18n({
  legacy: false,
  locale: 'tr',
  fallbackLocale: 'en',
  messages: { tr, en } as Record<string, any>
});

export async function setLanguage(code: string): Promise<void> {
  const global = i18n.global;
  if (!global.availableLocales.includes(code)) {
    const loader = localeLoaders[`./locales/${code}.json`];
    if (!loader) return;
    const messages = await loader();
    global.setLocaleMessage(code, messages.default as any);
  }
  global.locale.value = code;
  localStorage.setItem('lang', code);
  document.documentElement.lang = code;
}

export default i18n;
