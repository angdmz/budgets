import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import { formatDate } from '../lib/format';
import type { Budget, Group, CreateBudgetRequest, UpdateBudgetRequest } from '../lib/types';

export default function Budgets() {
  const { getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [formData, setFormData] = useState<CreateBudgetRequest>({
    name: '',
    description: '',
    start_date: '',
    end_date: '',
  });
  const [editingBudget, setEditingBudget] = useState<Budget | null>(null);
  const [deletingBudget, setDeletingBudget] = useState<Budget | null>(null);

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

  const createMutation = useMutation({
    mutationFn: async (data: CreateBudgetRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/groups/${selectedGroupId}/budgets`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      setIsModalOpen(false);
      setFormData({ name: '', description: '', start_date: '', end_date: '' });
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateBudgetRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/budgets/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      setEditingBudget(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/budgets/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      setDeletingBudget(null);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  const handleEdit = (budget: Budget) => {
    setEditingBudget(budget);
  };

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingBudget) {
      const updateData: UpdateBudgetRequest = {
        name: editingBudget.name,
        description: editingBudget.description,
        start_date: editingBudget.start_date,
        end_date: editingBudget.end_date,
      };
      updateMutation.mutate({ id: editingBudget.id, data: updateData });
    }
  };

  const handleDelete = () => {
    if (deletingBudget) {
      deleteMutation.mutate(deletingBudget.id);
    }
  };

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">{t('budgets.title')}</h1>
          <p className="mt-2 text-sm text-gray-700">{t('budgets.subtitle')}</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedGroupId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            {t('budgets.addBudget')}
          </button>
        </div>
      </div>

      <div className="mt-6">
        <label className="block text-sm font-medium text-gray-700">{t('common.selectGroup')}</label>
        <select
          value={selectedGroupId}
          onChange={(e) => setSelectedGroupId(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
        >
          <option value="">{t('common.selectGroupPlaceholder')}</option>
          {groups?.map((group) => (
            <option key={group.id} value={group.id}>
              {group.name}
            </option>
          ))}
        </select>
      </div>

      {selectedGroupId && (
        <div className="mt-8 flow-root">
          <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300">
              <thead className="bg-gray-50">
                <tr>
                  <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{t('common.name')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('budgets.period')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('common.description')}</th>
                  <th className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">{t('common.actions')}</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {budgets?.map((budget) => (
                  <tr key={budget.id}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                      {budget.name}
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {formatDate(budget.start_date)} - {formatDate(budget.end_date)}
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500">{budget.description || '-'}</td>
                    <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <button
                        onClick={() => handleEdit(budget)}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        {t('common.edit')}
                      </button>
                      <button
                        onClick={() => setDeletingBudget(budget)}
                        className="text-red-600 hover:text-red-900"
                      >
                        {t('common.delete')}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('budgets.createBudget')}</h2>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('budgets.startDate')}</label>
                  <input
                    type="date"
                    required
                    value={formData.start_date}
                    onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('budgets.endDate')}</label>
                  <input
                    type="date"
                    required
                    value={formData.end_date}
                    onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
              </div>
              {createMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.createError')}
                  {getErrorMessage(createMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createMutation.error)}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => { setIsModalOpen(false); createMutation.reset(); }}
                  className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
                >
                  {createMutation.isPending ? `${t('common.create')}...` : t('common.create')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editingBudget && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('budgets.editBudget')}</h2>
            <form onSubmit={handleUpdate}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={editingBudget.name}
                    onChange={(e) => setEditingBudget({ ...editingBudget, name: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('budgets.startDate')}</label>
                  <input
                    type="date"
                    required
                    value={editingBudget.start_date}
                    onChange={(e) => setEditingBudget({ ...editingBudget, start_date: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('budgets.endDate')}</label>
                  <input
                    type="date"
                    required
                    value={editingBudget.end_date}
                    onChange={(e) => setEditingBudget({ ...editingBudget, end_date: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
              </div>
              {updateMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.updateError')}
                  {getErrorMessage(updateMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(updateMutation.error)}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => { setEditingBudget(null); updateMutation.reset(); }}
                  className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={updateMutation.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
                >
                  {updateMutation.isPending ? `${t('common.update')}...` : t('common.update')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {deletingBudget && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">{t('budgets.deleteBudget')}</h2>
            <p className="text-sm text-gray-500 mb-4">
              {t('common.deleteConfirm', { name: deletingBudget.name }).replace(/\*\*/g, '')}
            </p>
            {deleteMutation.isError && (
              <p className="mb-4 text-sm text-red-600">
                {t('budgets.deleteError')}
                {getErrorMessage(deleteMutation.error) && (
                  <span className="block text-xs mt-1 opacity-75">{getErrorMessage(deleteMutation.error)}</span>
                )}
              </p>
            )}
            <div className="flex justify-end space-x-3">
              <button
                type="button"
                onClick={() => { setDeletingBudget(null); deleteMutation.reset(); }}
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
