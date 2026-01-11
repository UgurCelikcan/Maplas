import { createI18n } from 'vue-i18n';
import tr from './locales/tr.json';
import en from './locales/en.json';
import el from './locales/el.json';

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('lang') || 'tr',
  fallbackLocale: 'en',
  messages: {
    tr,
    en,
    el
  }
});

export default i18n;
