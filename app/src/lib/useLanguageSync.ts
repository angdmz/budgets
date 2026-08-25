import { useEffect } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import i18n from './i18n';
import { createApiClient } from './api';
import type { LanguageCode } from './languages';

interface UserPreferences {
  language: string;
}

export function useLanguageSync() {
  const { getAccessTokenSilently } = useAuth0();

  const { data: preferences } = useQuery({
    queryKey: ['preferences'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<UserPreferences>('/preferences');
      return response.data;
    },
  });

  useEffect(() => {
    if (preferences?.language) {
      i18n.changeLanguage(preferences.language.toLowerCase());
    }
  }, [preferences]);

  const changeLanguageMutation = useMutation({
    mutationFn: async (languageCode: LanguageCode) => {
      const api = await createApiClient(getAccessTokenSilently);
      await api.put('/preferences', { language: languageCode.toUpperCase() });
    },
  });

  const changeLanguage = async (languageCode: LanguageCode) => {
    await i18n.changeLanguage(languageCode);
    changeLanguageMutation.mutate(languageCode);
  };

  return { changeLanguage, currentLanguage: i18n.language as LanguageCode };
}
