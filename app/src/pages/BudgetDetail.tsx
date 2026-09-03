import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import { formatDate, formatCurrency } from '../lib/format';
import type {
  Budget,
  Group,
  Category,
  ExpectedExpense,
  ActualExpense,
  CreateExpectedExpenseRequest,
  UpdateExpectedExpenseRequest,
  CreateActualExpenseRequest,
  UpdateActualExpenseRequest,
} from '../lib/types';
import CategoryCombobox from '../components/CategoryCombobox';
import CurrencySelect from '../components/CurrencySelect';

export default function BudgetDetail() {
  const { budgetId } = useParams<{ budgetId: string }>();
  const navigate = useNavigate();
  const { getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [eeModalOpen, setEeModalOpen] = useState(false);
  const [eeEditing, setEeEditing] = useState<ExpectedExpense | null>(null);
  const [eeDeleting, setEeDeleting] = useState<ExpectedExpense | null>(null);
  const [eeForm, setEeForm] = useState<CreateExpectedExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    category_id: '',
  });
  const [eeCategoryError, setEeCategoryError] = useState(false);

  const [aeModalOpen, setAeModalOpen] = useState(false);
  const [aeEditing, setAeEditing] = useState<ActualExpense | null>(null);
  const [aeDeleting, setAeDeleting] = useState<ActualExpense | null>(null);
  const [aeForm, setAeForm] = useState<CreateActualExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    expense_date: new Date().toISOString().split('T')[0],
    category_id: '',
  });
  const [aeCategoryError, setAeCategoryError] = useState(false);

  const { data: budgetInfo } = useQuery({
    queryKey: ['budget', budgetId],
    queryFn: async () => {
      if (!budgetId) return null;
      const api = await createApiClient(getAccessTokenSilently);
      const groupsRes = await api.get<Group[]>('/groups');
      for (const group of groupsRes.data) {
        const budgetsRes = await api.get<Budget[]>(`/groups/${group.id}/budgets`);
        const found = budgetsRes.data.find((b) => b.id === budgetId);
        if (found) return { budget: found, groupId: group.id, groupName: group.name };
      }
      return null;
    },
    enabled: !!budgetId,
  });

  const groupId = budgetInfo?.groupId ?? '';

  const { data: categories = [] } = useQuery({
    queryKey: ['categories', groupId],
    queryFn: async () => {
      if (!groupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Category[]>(`/groups/${groupId}/categories`);
      return response.data;
    },
    enabled: !!groupId,
  });

  const { data: expectedExpenses } = useQuery({
    queryKey: ['expected-expenses', budgetId],
    queryFn: async () => {
      if (!budgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ExpectedExpense[]>(`/budgets/${budgetId}/expected-expenses`);
      return response.data;
    },
    enabled: !!budgetId,
  });

  const { data: actualExpenses } = useQuery({
    queryKey: ['actual-expenses', budgetId],
    queryFn: async () => {
      if (!budgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ActualExpense[]>(`/budgets/${budgetId}/actual-expenses`);
      return response.data;
    },
    enabled: !!budgetId,
  });

  // ── Expected expense mutations ──
  const eeCreateMut = useMutation({
    mutationFn: async (data: CreateExpectedExpenseRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/budgets/${budgetId}/expected-expenses`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setEeModalOpen(false);
      setEeForm({ name: '', description: '', amount: { amount: '', currency: 'USD' }, category_id: '' });
    },
  });

  const eeUpdateMut = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateExpectedExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/expected-expenses/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setEeEditing(null);
    },
  });

  const eeDeleteMut = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/expected-expenses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setEeDeleting(null);
    },
  });

  // ── Actual expense mutations ──
  const aeCreateMut = useMutation({
    mutationFn: async (data: CreateActualExpenseRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/budgets/${budgetId}/actual-expenses`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setAeModalOpen(false);
      setAeForm({
        name: '',
        description: '',
        amount: { amount: '', currency: 'USD' },
        expense_date: new Date().toISOString().split('T')[0],
        category_id: '',
      });
    },
  });

  const aeUpdateMut = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateActualExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/actual-expenses/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setAeEditing(null);
    },
  });

  const aeDeleteMut = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/actual-expenses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setAeDeleting(null);
    },
  });

  // ── Handlers ──
  const handleEeSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!eeForm.category_id) { setEeCategoryError(true); return; }
    setEeCategoryError(false);
    eeCreateMut.mutate(eeForm);
  };

  const handleEeUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (eeEditing) {
      eeUpdateMut.mutate({
        id: eeEditing.id,
        data: {
          name: eeEditing.name,
          description: eeEditing.description,
          amount: eeEditing.amount,
          category_id: eeEditing.category_id,
        },
      });
    }
  };

  const handleAeSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!aeForm.category_id) { setAeCategoryError(true); return; }
    setAeCategoryError(false);
    aeCreateMut.mutate(aeForm);
  };

  const handleAeUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (aeEditing) {
      aeUpdateMut.mutate({
        id: aeEditing.id,
        data: {
          name: aeEditing.name,
          description: aeEditing.description,
          amount: aeEditing.amount,
          expense_date: aeEditing.expense_date,
          category_id: aeEditing.category_id,
        },
      });
    }
  };

  const expectedTotal = expectedExpenses?.reduce((s, e) => s + parseFloat(e.amount.amount), 0) ?? 0;
  const actualTotal = actualExpenses?.reduce((s, e) => s + parseFloat(e.amount.amount), 0) ?? 0;

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <button
            onClick={() => navigate('/budgets')}
            className="text-sm text-primary-600 hover:text-primary-900 mb-2"
          >
            ← {t('budgetDetail.backToBudgets')}
          </button>
          <h1 className="text-2xl font-semibold text-gray-900">{budgetInfo?.budget.name ?? '...'}</h1>
          <p className="mt-1 text-sm text-gray-700">
            {budgetInfo && `${formatDate(budgetInfo.budget.start_date)} - ${formatDate(budgetInfo.budget.end_date)}`}
            {budgetInfo?.budget.description && ` · ${budgetInfo.budget.description}`}
          </p>
        </div>
      </div>

      {/* Summary cards */}
      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-gray-200">
          <p className="text-sm font-medium text-gray-500">{t('budgetDetail.expectedTotal')}</p>
          <p className="mt-1 text-2xl font-semibold text-gray-900">{formatCurrency(String(expectedTotal), 'USD')}</p>
        </div>
        <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-gray-200">
          <p className="text-sm font-medium text-gray-500">{t('budgetDetail.actualTotal')}</p>
          <p className="mt-1 text-2xl font-semibold text-gray-900">{formatCurrency(String(actualTotal), 'USD')}</p>
        </div>
        <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-gray-200">
          <p className="text-sm font-medium text-gray-500">{t('budgetDetail.difference')}</p>
          <p className={`mt-1 text-2xl font-semibold ${expectedTotal - actualTotal >= 0 ? 'text-green-600' : 'text-red-600'}`}>
            {formatCurrency(String(expectedTotal - actualTotal), 'USD')}
          </p>
        </div>
      </div>

      {/* Expected Expenses section */}
      <div className="mt-8">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">{t('budgetDetail.expectedExpenses')}</h2>
          <button
            onClick={() => setEeModalOpen(true)}
            className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
          >
            {t('expectedExpenses.addExpectedExpense')}
          </button>
        </div>
        <div className="mt-4 overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
          <table className="min-w-full divide-y divide-gray-300">
            <thead className="bg-gray-50">
              <tr>
                <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{t('common.name')}</th>
                <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expectedExpenses.amount')}</th>
                <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expectedExpenses.category')}</th>
                <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('common.description')}</th>
                <th className="relative py-3.5 pl-3 pr-4 sm:pr-6"><span className="sr-only">{t('common.actions')}</span></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {expectedExpenses?.map((expense) => {
                const category = categories.find((c) => c.id === expense.category_id);
                return (
                  <tr key={expense.id}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{expense.name}</td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{formatCurrency(expense.amount.amount, expense.amount.currency)}</td>
                    <td className="px-3 py-4 text-sm text-gray-500">
                      {category ? (
                        <div className="flex items-center gap-2">
                          <div className="w-3 h-3 rounded-full" style={{ backgroundColor: category.color }}></div>
                          <span>{category.name}</span>
                        </div>
                      ) : '-'}
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500">{expense.description || '-'}</td>
                    <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <button onClick={() => setEeEditing(expense)} className="text-blue-600 hover:text-blue-900 mr-4">{t('common.edit')}</button>
                      <button onClick={() => setEeDeleting(expense)} className="text-red-600 hover:text-red-900">{t('common.delete')}</button>
                    </td>
                  </tr>
                );
              })}
              {expectedExpenses?.length === 0 && (
                <tr><td colSpan={5} className="py-6 text-center text-sm text-gray-500">{t('budgetDetail.noExpectedExpenses')}</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Actual Expenses section */}
      <div className="mt-8">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">{t('budgetDetail.actualExpenses')}</h2>
          <button
            onClick={() => setAeModalOpen(true)}
            className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
          >
            {t('expenses.addExpense')}
          </button>
        </div>
        <div className="mt-4 overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
          <table className="min-w-full divide-y divide-gray-300">
            <thead className="bg-gray-50">
              <tr>
                <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{t('common.name')}</th>
                <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.date')}</th>
                <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.amount')}</th>
                <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.category')}</th>
                <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('common.description')}</th>
                <th className="relative py-3.5 pl-3 pr-4 sm:pr-6"><span className="sr-only">{t('common.actions')}</span></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {actualExpenses?.map((expense) => {
                const category = categories.find((c) => c.id === expense.category_id);
                return (
                  <tr key={expense.id}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{expense.name}</td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{formatDate(expense.expense_date)}</td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{formatCurrency(expense.amount.amount, expense.amount.currency)}</td>
                    <td className="px-3 py-4 text-sm text-gray-500">
                      {category ? (
                        <div className="flex items-center gap-2">
                          <div className="w-3 h-3 rounded-full" style={{ backgroundColor: category.color }}></div>
                          <span>{category.name}</span>
                        </div>
                      ) : '-'}
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500">{expense.description || '-'}</td>
                    <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <button onClick={() => setAeEditing(expense)} className="text-blue-600 hover:text-blue-900 mr-4">{t('common.edit')}</button>
                      <button onClick={() => setAeDeleting(expense)} className="text-red-600 hover:text-red-900">{t('common.delete')}</button>
                    </td>
                  </tr>
                );
              })}
              {actualExpenses?.length === 0 && (
                <tr><td colSpan={6} className="py-6 text-center text-sm text-gray-500">{t('budgetDetail.noActualExpenses')}</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* ── Expected Expense Create Modal ── */}
      {eeModalOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expectedExpenses.addExpectedExpense')}</h2>
            <form onSubmit={handleEeSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input type="text" required value={eeForm.name} onChange={(e) => setEeForm((p) => ({ ...p, name: e.target.value }))}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expectedExpenses.amount')}</label>
                  <div className="mt-1 flex gap-2">
                    <input type="number" step="0.01" required value={eeForm.amount.amount}
                      onChange={(e) => setEeForm((p) => ({ ...p, amount: { ...p.amount, amount: e.target.value } }))}
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                    <div className="w-28">
                      <CurrencySelect value={eeForm.amount.currency}
                        onChange={(currency) => setEeForm((p) => ({ ...p, amount: { ...p.amount, currency } }))} />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.description')}</label>
                  <input type="text" value={eeForm.description}
                    onChange={(e) => setEeForm((p) => ({ ...p, description: e.target.value }))}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expectedExpenses.category')}</label>
                  <CategoryCombobox groupId={groupId} value={eeForm.category_id}
                    onChange={(categoryId) => { setEeForm((p) => ({ ...p, category_id: categoryId })); setEeCategoryError(false); }}
                    getAccessTokenSilently={getAccessTokenSilently} />
                  {eeCategoryError && <p className="mt-1 text-sm text-red-600">{t('categories.categoryRequired')}</p>}
                </div>
              </div>
              {eeCreateMut.isError && (
                <p className="mt-2 text-sm text-red-600">{t('expectedExpenses.createError')}
                  {getErrorMessage(eeCreateMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(eeCreateMut.error)}</span>}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button type="button" onClick={() => { setEeModalOpen(false); eeCreateMut.reset(); }}
                  className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">{t('common.cancel')}</button>
                <button type="submit" disabled={eeCreateMut.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50">
                  {eeCreateMut.isPending ? `${t('common.create')}...` : t('common.create')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Expected Expense Edit Modal ── */}
      {eeEditing && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expectedExpenses.editExpectedExpense')}</h2>
            <form onSubmit={handleEeUpdate}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input type="text" required value={eeEditing.name}
                    onChange={(e) => setEeEditing((p) => p ? { ...p, name: e.target.value } : p)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expectedExpenses.amount')}</label>
                  <div className="mt-1 flex gap-2">
                    <input type="number" step="0.01" required value={eeEditing.amount.amount}
                      onChange={(e) => setEeEditing((p) => p ? { ...p, amount: { ...p.amount, amount: e.target.value } } : p)}
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                    <div className="w-28">
                      <CurrencySelect value={eeEditing.amount.currency}
                        onChange={(currency) => setEeEditing((p) => p ? { ...p, amount: { ...p.amount, currency } } : p)} />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.description')}</label>
                  <input type="text" value={eeEditing.description}
                    onChange={(e) => setEeEditing((p) => p ? { ...p, description: e.target.value } : p)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expectedExpenses.category')}</label>
                  <CategoryCombobox groupId={groupId} value={eeEditing.category_id}
                    onChange={(categoryId) => setEeEditing((p) => p ? { ...p, category_id: categoryId } : p)}
                    getAccessTokenSilently={getAccessTokenSilently} />
                </div>
              </div>
              {eeUpdateMut.isError && (
                <p className="mt-2 text-sm text-red-600">{t('expectedExpenses.updateError')}
                  {getErrorMessage(eeUpdateMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(eeUpdateMut.error)}</span>}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button type="button" onClick={() => { setEeEditing(null); eeUpdateMut.reset(); }}
                  className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">{t('common.cancel')}</button>
                <button type="submit" disabled={eeUpdateMut.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50">
                  {eeUpdateMut.isPending ? `${t('common.update')}...` : t('common.update')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Expected Expense Delete Modal ── */}
      {eeDeleting && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expectedExpenses.deleteExpectedExpense')}</h2>
            <p className="text-sm text-gray-500 mb-4">{t('common.deleteConfirm', { name: eeDeleting.name }).replace(/\*\*/g, '')}</p>
            {eeDeleteMut.isError && (
              <p className="mb-4 text-sm text-red-600">{t('expectedExpenses.deleteError')}
                {getErrorMessage(eeDeleteMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(eeDeleteMut.error)}</span>}
              </p>
            )}
            <div className="flex justify-end space-x-3">
              <button type="button" onClick={() => { setEeDeleting(null); eeDeleteMut.reset(); }}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">{t('common.cancel')}</button>
              <button type="button" onClick={() => eeDeleteMut.mutate(eeDeleting.id)} disabled={eeDeleteMut.isPending}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 disabled:opacity-50">
                {eeDeleteMut.isPending ? `${t('common.delete')}...` : t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Actual Expense Create Modal ── */}
      {aeModalOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expenses.addExpense')}</h2>
            <form onSubmit={handleAeSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input type="text" required value={aeForm.name}
                    onChange={(e) => setAeForm((p) => ({ ...p, name: e.target.value }))}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.amount')}</label>
                  <div className="mt-1 flex gap-2">
                    <input type="number" step="0.01" required value={aeForm.amount.amount}
                      onChange={(e) => setAeForm((p) => ({ ...p, amount: { ...p.amount, amount: e.target.value } }))}
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                    <div className="w-28">
                      <CurrencySelect value={aeForm.amount.currency}
                        onChange={(currency) => setAeForm((p) => ({ ...p, amount: { ...p.amount, currency } }))} />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.date')}</label>
                  <input type="date" required value={aeForm.expense_date}
                    onChange={(e) => setAeForm((p) => ({ ...p, expense_date: e.target.value }))}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.description')}</label>
                  <input type="text" value={aeForm.description}
                    onChange={(e) => setAeForm((p) => ({ ...p, description: e.target.value }))}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.category')}</label>
                  <CategoryCombobox groupId={groupId} value={aeForm.category_id}
                    onChange={(categoryId) => { setAeForm((p) => ({ ...p, category_id: categoryId })); setAeCategoryError(false); }}
                    getAccessTokenSilently={getAccessTokenSilently} />
                  {aeCategoryError && <p className="mt-1 text-sm text-red-600">{t('categories.categoryRequired')}</p>}
                </div>
              </div>
              {aeCreateMut.isError && (
                <p className="mt-2 text-sm text-red-600">{t('expenses.createError')}
                  {getErrorMessage(aeCreateMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(aeCreateMut.error)}</span>}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button type="button" onClick={() => { setAeModalOpen(false); aeCreateMut.reset(); }}
                  className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">{t('common.cancel')}</button>
                <button type="submit" disabled={aeCreateMut.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50">
                  {aeCreateMut.isPending ? `${t('common.create')}...` : t('common.create')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Actual Expense Edit Modal ── */}
      {aeEditing && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expenses.editExpense')}</h2>
            <form onSubmit={handleAeUpdate}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input type="text" required value={aeEditing.name}
                    onChange={(e) => setAeEditing((p) => p ? { ...p, name: e.target.value } : p)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.amount')}</label>
                  <div className="mt-1 flex gap-2">
                    <input type="number" step="0.01" required value={aeEditing.amount.amount}
                      onChange={(e) => setAeEditing((p) => p ? { ...p, amount: { ...p.amount, amount: e.target.value } } : p)}
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                    <div className="w-28">
                      <CurrencySelect value={aeEditing.amount.currency}
                        onChange={(currency) => setAeEditing((p) => p ? { ...p, amount: { ...p.amount, currency } } : p)} />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.date')}</label>
                  <input type="date" required value={aeEditing.expense_date}
                    onChange={(e) => setAeEditing((p) => p ? { ...p, expense_date: e.target.value } : p)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.description')}</label>
                  <input type="text" value={aeEditing.description}
                    onChange={(e) => setAeEditing((p) => p ? { ...p, description: e.target.value } : p)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.category')}</label>
                  <CategoryCombobox groupId={groupId} value={aeEditing.category_id}
                    onChange={(categoryId) => setAeEditing((p) => p ? { ...p, category_id: categoryId } : p)}
                    getAccessTokenSilently={getAccessTokenSilently} />
                </div>
              </div>
              {aeUpdateMut.isError && (
                <p className="mt-2 text-sm text-red-600">{t('expenses.updateError')}
                  {getErrorMessage(aeUpdateMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(aeUpdateMut.error)}</span>}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button type="button" onClick={() => { setAeEditing(null); aeUpdateMut.reset(); }}
                  className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">{t('common.cancel')}</button>
                <button type="submit" disabled={aeUpdateMut.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50">
                  {aeUpdateMut.isPending ? `${t('common.update')}...` : t('common.update')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ── Actual Expense Delete Modal ── */}
      {aeDeleting && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expenses.deleteExpense')}</h2>
            <p className="text-sm text-gray-500 mb-4">{t('common.deleteConfirm', { name: aeDeleting.name }).replace(/\*\*/g, '')}</p>
            {aeDeleteMut.isError && (
              <p className="mb-4 text-sm text-red-600">{t('expenses.deleteError')}
                {getErrorMessage(aeDeleteMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(aeDeleteMut.error)}</span>}
              </p>
            )}
            <div className="flex justify-end space-x-3">
              <button type="button" onClick={() => { setAeDeleting(null); aeDeleteMut.reset(); }}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">{t('common.cancel')}</button>
              <button type="button" onClick={() => aeDeleteMut.mutate(aeDeleting.id)} disabled={aeDeleteMut.isPending}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 disabled:opacity-50">
                {aeDeleteMut.isPending ? `${t('common.delete')}...` : t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
