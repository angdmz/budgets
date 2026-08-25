import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import HttpBackend from 'i18next-http-backend';
import { DEFAULT_LANGUAGE } from './languages';

i18next
  .use(HttpBackend)
  .use(initReactI18next)
  .init({
    lng: DEFAULT_LANGUAGE,
    fallbackLng: DEFAULT_LANGUAGE,
    interpolation: {
      escapeValue: false,
    },
    backend: {
      loadPath: '/app/locales/{{lng}}/{{ns}}.json',
    },
  });

export default i18next;
