import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import type { Category, Group, CreateCategoryRequest, UpdateCategoryRequest } from '../lib/types';
import Dialog from '../components/Dialog';

export default function Categories() {
  const { getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [formData, setFormData] = useState<CreateCategoryRequest>({ name: '', description: '', color: '#0ea5e9' });
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);
  const [deletingCategory, setDeletingCategory] = useState<Category | null>(null);

  const { data: groups } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Group[]>('/groups');
      return response.data;
    },
  });

  const { data: categories } = useQuery({
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
    mutationFn: async (data: CreateCategoryRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/groups/${selectedGroupId}/categories`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setIsModalOpen(false);
      setFormData({ name: '', description: '', color: '#0ea5e9' });
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateCategoryRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/categories/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setEditingCategory(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/categories/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setDeletingCategory(null);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  const handleEdit = (category: Category) => {
    setEditingCategory(category);
  };

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingCategory) {
      const updateData: UpdateCategoryRequest = {
        name: editingCategory.name,
        description: editingCategory.description,
        color: editingCategory.color,
      };
      updateMutation.mutate({ id: editingCategory.id, data: updateData });
    }
  };

  const handleDelete = () => {
    if (deletingCategory) {
      deleteMutation.mutate(deletingCategory.id);
    }
  };

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">{t('categories.title')}</h1>
          <p className="mt-2 text-sm text-gray-700">{t('categories.subtitle')}</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedGroupId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            {t('categories.addCategory')}
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
            <option key={group.id} value={group.id}>{group.name}</option>
          ))}
        </select>
      </div>

      {selectedGroupId && (
        <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {categories?.map((category) => (
            <div key={category.id} className="bg-white shadow rounded-lg p-4 border-l-4" style={{ borderColor: category.color || '#0ea5e9' }}>
              <div className="flex justify-between items-start">
                <div className="flex-1">
                  <h3 className="text-lg font-medium text-gray-900">{category.name}</h3>
                  <p className="mt-1 text-sm text-gray-500">{category.description || t('categories.noDescription')}</p>
                </div>
                <div className="flex space-x-2">
                  <button
                    onClick={() => handleEdit(category)}
                    className="text-blue-600 hover:text-blue-800 min-h-[44px] px-2"
                  >
                    {t('common.edit')}
                  </button>
                  <button
                    onClick={() => setDeletingCategory(category)}
                    className="text-red-600 hover:text-red-800 min-h-[44px] px-2"
                  >
                    {t('common.delete')}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {isModalOpen && (
        <Dialog title={t('categories.createCategory')} onClose={() => { setIsModalOpen(false); createMutation.reset(); }}>
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
                  <label className="block text-sm font-medium text-gray-700">{t('categories.color')}</label>
                  <input
                    type="color"
                    value={formData.color}
                    onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                    className="mt-1 block w-full h-10 rounded-md border-gray-300 shadow-sm"
                  />
                </div>
              </div>
              {createMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('categories.createError')}
                  {getErrorMessage(createMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createMutation.error)}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
                <button type="button" onClick={() => { setIsModalOpen(false); createMutation.reset(); }} className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 min-h-[44px]">
                  {t('common.cancel')}
                </button>
                <button type="submit" disabled={createMutation.isPending} className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 min-h-[44px]">
                  {createMutation.isPending ? `${t('common.create')}...` : t('common.create')}
                </button>
              </div>
            </form>
        </Dialog>
      )}

      {editingCategory && (
        <Dialog title={t('categories.editCategory')} onClose={() => { setEditingCategory(null); updateMutation.reset(); }}>
            <form onSubmit={handleUpdate}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={editingCategory.name}
                    onChange={(e) => setEditingCategory({ ...editingCategory, name: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">{t('categories.color')}</label>
                  <input
                    type="color"
                    value={editingCategory.color}
                    onChange={(e) => setEditingCategory({ ...editingCategory, color: e.target.value })}
                    className="mt-1 block w-full h-10 rounded-md border-gray-300 shadow-sm"
                  />
                </div>
              </div>
              {updateMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('categories.updateError')}
                  {getErrorMessage(updateMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(updateMutation.error)}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
                <button type="button" onClick={() => { setEditingCategory(null); updateMutation.reset(); }} className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 min-h-[44px]">
                  {t('common.cancel')}
                </button>
                <button type="submit" disabled={updateMutation.isPending} className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 min-h-[44px]">
                  {updateMutation.isPending ? `${t('common.update')}...` : t('common.update')}
                </button>
              </div>
            </form>
        </Dialog>
      )}

      {deletingCategory && (
        <Dialog title={t('categories.deleteCategory')} onClose={() => { setDeletingCategory(null); deleteMutation.reset(); }}>
            <p className="text-sm text-gray-500 mb-4">
              {t('common.deleteConfirm', { name: deletingCategory.name }).replace(/\*\*/g, '')}
            </p>
            {deleteMutation.isError && (
              <p className="mb-4 text-sm text-red-600">
                {t('categories.deleteError')}
                {getErrorMessage(deleteMutation.error) && (
                  <span className="block text-xs mt-1 opacity-75">{getErrorMessage(deleteMutation.error)}</span>
                )}
              </p>
            )}
            <div className="flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
              <button
                type="button"
                onClick={() => { setDeletingCategory(null); deleteMutation.reset(); }}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 min-h-[44px]"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 disabled:opacity-50 min-h-[44px]"
              >
                {deleteMutation.isPending ? `${t('common.delete')}...` : t('common.delete')}
              </button>
            </div>
        </Dialog>
      )}
    </div>
  );
}
