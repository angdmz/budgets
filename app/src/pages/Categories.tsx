import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { createApiClient } from '../lib/api';
import type { Category, Group, CreateCategoryRequest } from '../lib/types';

export default function Categories() {
  const { getAccessTokenSilently } = useAuth0();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [formData, setFormData] = useState<CreateCategoryRequest>({ name: '', description: '', color: '#0ea5e9' });

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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Categories</h1>
          <p className="mt-2 text-sm text-gray-700">Manage expense categories</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedGroupId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            Add Category
          </button>
        </div>
      </div>

      <div className="mt-6">
        <label className="block text-sm font-medium text-gray-700">Select Group</label>
        <select
          value={selectedGroupId}
          onChange={(e) => setSelectedGroupId(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
        >
          <option value="">Select a group...</option>
          {groups?.map((group) => (
            <option key={group.id} value={group.id}>{group.name}</option>
          ))}
        </select>
      </div>

      {selectedGroupId && (
        <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {categories?.map((category) => (
            <div key={category.id} className="bg-white shadow rounded-lg p-4 border-l-4" style={{ borderColor: category.color || '#0ea5e9' }}>
              <h3 className="text-lg font-medium text-gray-900">{category.name}</h3>
              <p className="mt-1 text-sm text-gray-500">{category.description || 'No description'}</p>
            </div>
          ))}
        </div>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">Create Category</h2>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Name</label>
                  <input
                    type="text"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Color</label>
                  <input
                    type="color"
                    value={formData.color}
                    onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                    className="mt-1 block w-full h-10 rounded-md border-gray-300 shadow-sm"
                  />
                </div>
              </div>
              <div className="mt-6 flex justify-end space-x-3">
                <button type="button" onClick={() => setIsModalOpen(false)} className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">
                  Cancel
                </button>
                <button type="submit" className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500">
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
