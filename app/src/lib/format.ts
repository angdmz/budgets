import i18n from './i18n';

const LOCALE_MAP: Record<string, string> = {
  en: 'en-US',
  es: 'es-AR',
};

export function formatDate(date: string | Date): string {
  const locale = LOCALE_MAP[i18n.language] || 'en-US';
  return new Date(date).toLocaleDateString(locale);
}

export function formatCurrency(amount: string, currency: string): string {
  const locale = LOCALE_MAP[i18n.language] || 'en-US';
  return new Intl.NumberFormat(locale, { style: 'currency', currency }).format(parseFloat(amount));
}
