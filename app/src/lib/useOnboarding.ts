import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { createApiClient } from './api';
import type {
  OnboardingResponse,
  OnboardingStep,
  OnboardingStepData,
} from './types';

export const ONBOARDING_STEPS: OnboardingStep[] = [
  'welcome',
  'choose_group',
  'choose_cadence',
  'add_expected_expenses',
  'budget_summary',
  'register_actual_expense',
  'compare_expenses',
  'dashboard_tour',
  'complete',
];

export function useOnboarding() {
  const { getAccessTokenSilently } = useAuth0();
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['onboarding'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<OnboardingResponse>('/onboarding');
      return response.data;
    },
  });

  const completeStepMutation = useMutation({
    mutationFn: async ({ step, data }: { step: OnboardingStep; data?: OnboardingStepData }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<OnboardingResponse>(
        `/onboarding/steps/${step}/complete`,
        { data: data ?? {} },
      );
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['onboarding'], data);
    },
  });

  const skipStepMutation = useMutation({
    mutationFn: async (step: OnboardingStep) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<OnboardingResponse>(`/onboarding/steps/${step}/skip`, {});
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['onboarding'], data);
    },
  });

  const goBackMutation = useMutation({
    mutationFn: async (step: OnboardingStep) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<OnboardingResponse>(`/onboarding/steps/${step}/back`, {});
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['onboarding'], data);
    },
  });

  const resetMutation = useMutation({
    mutationFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<OnboardingResponse>('/onboarding/reset', {});
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['onboarding'], data);
    },
  });

  return {
    onboarding: data,
    isLoading,
    error,
    currentStep: data?.current_step ?? 'welcome',
    steps: data?.steps ?? [],
    isCompleted: data?.status === 'completed' || data?.status === 'skipped',
    completeStep: completeStepMutation.mutateAsync,
    skipStep: skipStepMutation.mutateAsync,
    goBack: goBackMutation.mutateAsync,
    reset: resetMutation.mutateAsync,
    isPending: completeStepMutation.isPending || skipStepMutation.isPending || goBackMutation.isPending || resetMutation.isPending,
  };
}
