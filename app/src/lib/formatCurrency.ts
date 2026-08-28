import type { Currency } from './usePreferences';

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

export function formatCurrency(amount: string | number, currency: Currency): string {
  const numeric = typeof amount === 'string' ? parseFloat(amount) : amount;
  if (isNaN(numeric)) {
    return `${amount} ${currency}`;
  }
  const locale = CURRENCY_LOCALE[currency] || 'en-US';
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
  }).format(numeric);
}
