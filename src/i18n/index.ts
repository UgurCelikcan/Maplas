import { createI18n } from 'vue-i18n';
import tr from './locales/tr.json';
import en from './locales/en.json';
import el from './locales/el.json';
import de from './locales/de.json';
import fr from './locales/fr.json';
import es from './locales/es.json';
import it from './locales/it.json';
import pt from './locales/pt.json';
import ru from './locales/ru.json';
import ja from './locales/ja.json';
import zhCN from './locales/zh-CN.json';
import ko from './locales/ko.json';
import ar from './locales/ar.json';

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('lang') || 'tr',
  fallbackLocale: 'en',
  messages: {
    tr,
    en,
    el,
    de,
    fr,
    es,
    it,
    pt,
    ru,
    ja,
    'zh-CN': zhCN,
    ko,
    ar
  }
});

export default i18n;