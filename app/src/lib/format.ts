import i18n from './i18n';

const LOCALE_MAP: Record<string, string> = {
  en: 'en-US',
  es: 'es-AR',
};

const CURRENCY_LOCALE: Record<string, string> = {
  USD: 'en-US',
  EUR: 'de-DE',
  GBP: 'en-GB',
  ARS: 'es-AR',
  BRL: 'pt-BR',
  MXN: 'es-MX',
  CLP: 'es-CL',
  COP: 'es-CO',
  PEN: 'es-PE',
  UYU: 'es-UY',
};

export function formatDate(date: string | Date): string {
  const locale = LOCALE_MAP[i18n.language] || 'en-US';
  return new Date(date).toLocaleDateString(locale);
}

export function formatCurrency(amount: string | number, currency: string): string {
  const numeric = typeof amount === 'string' ? parseFloat(amount) : amount;
  if (isNaN(numeric)) {
    return `${amount} ${currency}`;
  }
  const locale = CURRENCY_LOCALE[currency] || LOCALE_MAP[i18n.language] || 'en-US';
  return new Intl.NumberFormat(locale, { style: 'currency', currency, currencyDisplay: 'code' }).format(numeric);
}
