import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import { formatDate, formatCurrency } from '../lib/format';
import type { ActualExpense, Budget, Group, Category, CreateActualExpenseRequest, UpdateActualExpenseRequest } from '../lib/types';
import CategoryCombobox from '../components/CategoryCombobox';
import CurrencySelect from '../components/CurrencySelect';

export default function Expenses() {
  const { getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [selectedBudgetId, setSelectedBudgetId] = useState('');
  const [formData, setFormData] = useState<CreateActualExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    expense_date: new Date().toISOString().split('T')[0],
    category_id: '',
  });
  const [editingExpense, setEditingExpense] = useState<ActualExpense | null>(null);
  const [deletingExpense, setDeletingExpense] = useState<ActualExpense | null>(null);
  const [categoryError, setCategoryError] = useState(false);

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

  const { data: expenses } = useQuery({
    queryKey: ['actual-expenses', selectedBudgetId],
    queryFn: async () => {
      if (!selectedBudgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ActualExpense[]>(`/budgets/${selectedBudgetId}/actual-expenses`);
      return response.data;
    },
    enabled: !!selectedBudgetId,
  });

  const { data: categories = [] } = useQuery({
    queryKey: ['categories', selectedGroupId],
    queryFn: async () => {
      if (!selectedGroupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Category[]>(`/groups/${selectedGroupId}/categories`);
      return response.data;
    },
    enabled: !!selectedGroupId,
  });

  const createMutation = useMutation({
    mutationFn: async (data: CreateActualExpenseRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/budgets/${selectedBudgetId}/actual-expenses`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setIsModalOpen(false);
      setFormData({
        name: '',
        description: '',
        amount: { amount: '', currency: 'USD' },
        expense_date: new Date().toISOString().split('T')[0],
        category_id: '',
      });
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateActualExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/actual-expenses/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setEditingExpense(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/actual-expenses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setDeletingExpense(null);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.category_id) {
      setCategoryError(true);
      return;
    }
    setCategoryError(false);
    createMutation.mutate(formData);
  };

  const handleEdit = (expense: ActualExpense) => {
    setEditingExpense(expense);
  };

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingExpense) {
      const updateData: UpdateActualExpenseRequest = {
        name: editingExpense.name,
        description: editingExpense.description,
        amount: editingExpense.amount,
        expense_date: editingExpense.expense_date,
        category_id: editingExpense.category_id,
      };
      updateMutation.mutate({ id: editingExpense.id, data: updateData });
    }
  };

  const handleDelete = () => {
    if (deletingExpense) {
      deleteMutation.mutate(deletingExpense.id);
    }
  };

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">{t('expenses.title')}</h1>
          <p className="mt-2 text-sm text-gray-700">{t('expenses.subtitle')}</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedBudgetId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            {t('expenses.addExpense')}
          </button>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label className="block text-sm font-medium text-gray-700">{t('common.selectGroup')}</label>
          <select
            value={selectedGroupId}
            onChange={(e) => { setSelectedGroupId(e.target.value); setSelectedBudgetId(''); }}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          >
            <option value="">{t('common.selectGroupPlaceholder')}</option>
            {groups?.map((group) => (
              <option key={group.id} value={group.id}>{group.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">{t('common.selectBudget')}</label>
          <select
            value={selectedBudgetId}
            onChange={(e) => setSelectedBudgetId(e.target.value)}
            disabled={!selectedGroupId}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm disabled:opacity-50"
          >
            <option value="">{t('common.selectBudgetPlaceholder')}</option>
            {budgets?.map((budget) => (
              <option key={budget.id} value={budget.id}>{budget.name}</option>
            ))}
          </select>
        </div>
      </div>

      {selectedBudgetId && (
        <div className="mt-8 flow-root">
          <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300">
              <thead className="bg-gray-50">
                <tr>
                  <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{t('common.name')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.date')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.amount')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.category')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('common.description')}</th>
                  <th className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">{t('common.actions')}</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {expenses?.map((expense) => {
                  const category = categories.find(c => c.id === expense.category_id);
                  return (
                  <tr key={expense.id}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                      {expense.name}
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {formatDate(expense.expense_date)}
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {formatCurrency(expense.amount.amount, expense.amount.currency)}
                    </td>
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
                      <button
                        onClick={() => handleEdit(expense)}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        {t('common.edit')}
                      </button>
                      <button
                        onClick={() => setDeletingExpense(expense)}
                        className="text-red-600 hover:text-red-900"
                      >
                        {t('common.delete')}
                      </button>
                    </td>
                  </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expenses.addExpense')}</h2>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.amount')}</label>
                  <div className="mt-1 flex gap-2">
                    <input
                      type="number"
                      step="0.01"
                      required
                      value={formData.amount.amount}
                      onChange={(e) => setFormData(prev => ({ ...prev, amount: { ...prev.amount, amount: e.target.value } }))}
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                    <div className="w-28">
                      <CurrencySelect
                        value={formData.amount.currency}
                        onChange={(currency) => setFormData(prev => ({ ...prev, amount: { ...prev.amount, currency } }))}
                      />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.date')}</label>
                  <input
                    type="date"
                    required
                    value={formData.expense_date}
                    onChange={(e) => setFormData(prev => ({ ...prev, expense_date: e.target.value }))}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.category')}</label>
                  <CategoryCombobox
                    groupId={selectedGroupId}
                    value={formData.category_id}
                    onChange={(categoryId) => { setFormData(prev => ({ ...prev, category_id: categoryId })); setCategoryError(false); }}
                    getAccessTokenSilently={getAccessTokenSilently}
                  />
                  {categoryError && (
                    <p className="mt-1 text-sm text-red-600">{t('categories.categoryRequired')}</p>
                  )}
                </div>
              </div>
              {createMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('expenses.createError')}
                  {getErrorMessage(createMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createMutation.error)}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button type="button" onClick={() => { setIsModalOpen(false); createMutation.reset(); }} className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">
                  {t('common.cancel')}
                </button>
                <button type="submit" disabled={createMutation.isPending} className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50">
                  {createMutation.isPending ? `${t('common.create')}...` : t('common.create')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editingExpense && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expenses.editExpense')}</h2>
            <form onSubmit={handleUpdate}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={editingExpense.name}
                    onChange={(e) => setEditingExpense(prev => prev ? { ...prev, name: e.target.value } : null)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.amount')}</label>
                  <div className="mt-1 flex gap-2">
                    <input
                      type="number"
                      step="0.01"
                      required
                      value={editingExpense.amount.amount}
                      onChange={(e) => setEditingExpense(prev => prev ? { ...prev, amount: { ...prev.amount, amount: e.target.value } } : null)}
                      className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    />
                    <div className="w-28">
                      <CurrencySelect
                        value={editingExpense.amount.currency}
                        onChange={(currency) => setEditingExpense(prev => prev ? { ...prev, amount: { ...prev.amount, currency } } : null)}
                      />
                    </div>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.date')}</label>
                  <input
                    type="date"
                    required
                    value={editingExpense.expense_date}
                    onChange={(e) => setEditingExpense(prev => prev ? { ...prev, expense_date: e.target.value } : null)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('expenses.category')}</label>
                  <CategoryCombobox
                    groupId={selectedGroupId}
                    value={editingExpense.category_id}
                    onChange={(categoryId) => setEditingExpense(prev => prev ? { ...prev, category_id: categoryId } : null)}
                    getAccessTokenSilently={getAccessTokenSilently}
                  />
                </div>
              </div>
              {updateMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('expenses.updateError')}
                  {getErrorMessage(updateMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(updateMutation.error)}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button type="button" onClick={() => { setEditingExpense(null); updateMutation.reset(); }} className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">
                  {t('common.cancel')}
                </button>
                <button type="submit" disabled={updateMutation.isPending} className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50">
                  {updateMutation.isPending ? `${t('common.update')}...` : t('common.update')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {deletingExpense && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('expenses.deleteExpense')}</h2>
            <p className="text-sm text-gray-500 mb-4">
              {t('common.deleteConfirm', { name: deletingExpense.name }).replace(/\*\*/g, '')}
            </p>
            {deleteMutation.isError && (
              <p className="mb-4 text-sm text-red-600">
                {t('expenses.deleteError')}
                {getErrorMessage(deleteMutation.error) && (
                  <span className="block text-xs mt-1 opacity-75">{getErrorMessage(deleteMutation.error)}</span>
                )}
              </p>
            )}
            <div className="flex justify-end space-x-3">
              <button
                type="button"
                onClick={() => { setDeletingExpense(null); deleteMutation.reset(); }}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 disabled:opacity-50"
              >
                {deleteMutation.isPending ? `${t('common.delete')}...` : t('common.delete')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
