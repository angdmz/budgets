import { useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import i18n from './i18n';
import { createApiClient } from './api';
import type { LanguageCode } from './languages';

export type Theme = 'LIGHT' | 'DIM' | 'DARK';

export type Currency = 'USD' | 'EUR' | 'GBP' | 'ARS' | 'BRL' | 'MXN' | 'CLP' | 'COP' | 'PEN' | 'UYU';

export type QuoteType = 'OFFICIAL' | 'BLUE' | 'MEP' | 'CCL' | 'CRYPTO';

const THEME_STORAGE_KEY = 'budgets.theme';

export interface UserPreferences {
  theme: Theme;
  language: string;
  display_currency: Currency;
  preferred_quote_type: QuoteType;
}

export const SUPPORTED_THEMES: { value: Theme; label: string }[] = [
  { value: 'LIGHT', label: 'Light' },
  { value: 'DARK', label: 'Dark' },
];

export const SUPPORTED_CURRENCIES: { value: Currency; label: string }[] = [
  { value: 'USD', label: 'USD - US Dollar' },
  { value: 'EUR', label: 'EUR - Euro' },
  { value: 'GBP', label: 'GBP - British Pound' },
  { value: 'ARS', label: 'ARS - Argentine Peso' },
  { value: 'BRL', label: 'BRL - Brazilian Real' },
  { value: 'MXN', label: 'MXN - Mexican Peso' },
  { value: 'CLP', label: 'CLP - Chilean Peso' },
  { value: 'COP', label: 'COP - Colombian Peso' },
  { value: 'PEN', label: 'PEN - Peruvian Sol' },
  { value: 'UYU', label: 'UYU - Uruguayan Peso' },
];

export const SUPPORTED_QUOTE_TYPES: { value: QuoteType; label: string }[] = [
  { value: 'OFFICIAL', label: 'Official' },
  { value: 'BLUE', label: 'Blue' },
  { value: 'MEP', label: 'MEP' },
  { value: 'CCL', label: 'CCL' },
  { value: 'CRYPTO', label: 'Crypto' },
];

function readCachedTheme(): Theme | null {
  const value = localStorage.getItem(THEME_STORAGE_KEY);
  return value === 'LIGHT' || value === 'DIM' || value === 'DARK' ? value : null;
}

function applyThemeClass(theme: Theme) {
  const root = document.documentElement;
  root.classList.remove('dark', 'dim');
  if (theme === 'DARK') {
    root.classList.add('dark');
  } else if (theme === 'DIM') {
    root.classList.add('dim');
  }
}

// Apply cached theme synchronously to prevent flash on reload
const cachedTheme = readCachedTheme();
if (cachedTheme) {
  applyThemeClass(cachedTheme);
}

export function usePreferences() {
  const { getAccessTokenSilently } = useAuth0();
  const queryClient = useQueryClient();

  const { data: preferences } = useQuery({
    queryKey: ['preferences'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<UserPreferences>('/preferences');
      return response.data;
    },
  });

  useEffect(() => {
    if (preferences?.theme) {
      applyThemeClass(preferences.theme);
      localStorage.setItem(THEME_STORAGE_KEY, preferences.theme);
    }
  }, [preferences?.theme]);

  useEffect(() => {
    if (preferences?.language) {
      i18n.changeLanguage(preferences.language.toLowerCase());
    }
  }, [preferences?.language]);

  const patchMutation = useMutation({
    mutationFn: async (patch: Partial<Pick<UserPreferences, 'theme' | 'language' | 'display_currency' | 'preferred_quote_type'>>) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.patch<UserPreferences>('/preferences', patch);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['preferences'], data);
      queryClient.invalidateQueries({ queryKey: ['budget-summary'] });
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
    },
  });

  const updateTheme = useCallback((theme: Theme) => {
    const previousTheme = preferences?.theme ?? readCachedTheme() ?? 'LIGHT';
    applyThemeClass(theme);
    patchMutation.mutate({ theme }, {
      onError: () => {
        applyThemeClass(previousTheme);
      },
    });
  }, [patchMutation, preferences?.theme]);

  const updateLanguage = useCallback((languageCode: LanguageCode) => {
    i18n.changeLanguage(languageCode);
    patchMutation.mutate({ language: languageCode.toUpperCase() });
  }, [patchMutation]);

  const updateDisplayCurrency = useCallback((currency: Currency) => {
    patchMutation.mutate({ display_currency: currency });
  }, [patchMutation]);

  const updatePreferredQuoteType = useCallback((quoteType: QuoteType) => {
    patchMutation.mutate({ preferred_quote_type: quoteType });
  }, [patchMutation]);

  return {
    preferences,
    theme: preferences?.theme ?? readCachedTheme() ?? 'LIGHT',
    updateTheme,
    updateLanguage,
    updateDisplayCurrency,
    updatePreferredQuoteType,
    currentLanguage: i18n.language as LanguageCode,
    isUpdating: patchMutation.isPending,
  };
}
