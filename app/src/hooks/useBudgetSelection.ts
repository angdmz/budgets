import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { createApiClient } from '../lib/api';
import type { Budget, Group, ExpectedExpense, ActualExpense, BudgetSummary } from '../lib/types';

export function useBudgetSelection() {
  const { getAccessTokenSilently } = useAuth0();
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [selectedBudgetId, setSelectedBudgetId] = useState('');

  const { data: groups } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Group[]>('/groups');
      return response.data;
    },
  });

  const { data: budgets } = useQuery({
    queryKey: ['budgets', selectedGroupId],
    queryFn: async () => {
      if (!selectedGroupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Budget[]>(`/groups/${selectedGroupId}/budgets`);
      return response.data;
    },
    enabled: !!selectedGroupId,
  });

  const { data: expectedExpenses } = useQuery({
    queryKey: ['expected-expenses', selectedBudgetId],
    queryFn: async () => {
      if (!selectedBudgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ExpectedExpense[]>(`/budgets/${selectedBudgetId}/expected-expenses`);
      return response.data;
    },
    enabled: !!selectedBudgetId,
  });

  const { data: actualExpenses } = useQuery({
    queryKey: ['actual-expenses', selectedBudgetId],
    queryFn: async () => {
      if (!selectedBudgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ActualExpense[]>(`/budgets/${selectedBudgetId}/actual-expenses`);
      return response.data;
    },
    enabled: !!selectedBudgetId,
  });

  const { data: summary } = useQuery({
    queryKey: ['budget-summary', selectedBudgetId],
    queryFn: async () => {
      if (!selectedBudgetId) return null;
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<BudgetSummary>(`/budgets/${selectedBudgetId}/summary`);
      return response.data;
    },
    enabled: !!selectedBudgetId,
  });

  const expectedTotal = summary ? parseFloat(summary.expected_total.amount) : 0;
  const actualTotal = summary ? parseFloat(summary.actual_total.amount) : 0;
  const difference = summary ? parseFloat(summary.difference.amount) : 0;
  const currency = summary?.expected_total.currency ?? 'USD';

  return {
    selectedGroupId,
    selectedBudgetId,
    setSelectedGroupId,
    setSelectedBudgetId,
    groups,
    budgets,
    expectedExpenses,
    actualExpenses,
    expectedTotal,
    actualTotal,
    difference,
    currency,
  };
}
