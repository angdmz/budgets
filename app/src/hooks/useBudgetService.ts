import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { createApiClient } from '../lib/api';
import {
  createBudgetService,
  type DuplicateBudgetOptions,
  type CreateRecurringOptions,
  type CreateRecurringProgress,
  type CreateBudgetsForMonthsOptions,
  type CreateBudgetsForMonthsProgress,
  type CreateBudgetPlanOptions,
  type CreateBudgetPlanProgress,
} from '../lib/services/budgetService';

export function useBudgetService() {
  const { getAccessTokenSilently } = useAuth0();
  const queryClient = useQueryClient();

  const duplicateBudget = useMutation({
    mutationFn: async (options: DuplicateBudgetOptions) => {
      const api = await createApiClient(getAccessTokenSilently);
      const service = createBudgetService(api);
      return service.duplicateBudget(options);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
    },
  });

  const createRecurringBudgets = useMutation({
    mutationFn: async (params: {
      options: CreateRecurringOptions;
      onProgress?: (progress: CreateRecurringProgress) => void;
    }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const service = createBudgetService(api);
      return service.createRecurringBudgets(params.options, params.onProgress);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
    },
  });

  const createBudgetsForMonths = useMutation({
    mutationFn: async (params: {
      options: CreateBudgetsForMonthsOptions;
      onProgress?: (progress: CreateBudgetsForMonthsProgress) => void;
    }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const service = createBudgetService(api);
      return service.createBudgetsForMonths(params.options, params.onProgress);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
    },
  });

  const createBudgetPlan = useMutation({
    mutationFn: async (params: {
      options: CreateBudgetPlanOptions;
      onProgress?: (progress: CreateBudgetPlanProgress) => void;
    }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const service = createBudgetService(api);
      return service.createBudgetPlan(params.options, params.onProgress);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
    },
  });

  return { duplicateBudget, createRecurringBudgets, createBudgetsForMonths, createBudgetPlan };
}
