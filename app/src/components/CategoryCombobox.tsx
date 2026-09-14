import { useState, useRef, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { createApiClient } from '../lib/api';
import type { Category, CreateCategoryRequest } from '../lib/types';

interface CategoryComboboxProps {
  groupId: string;
  value: string;
  onChange: (categoryId: string) => void;
  getAccessTokenSilently: () => Promise<string>;
  allowCreate?: boolean;
  placeholder?: string;
}

export default function CategoryCombobox({ groupId, value, onChange, getAccessTokenSilently, allowCreate = false, placeholder }: CategoryComboboxProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);
  const { t } = useTranslation();
  const queryClient = useQueryClient();

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

  const createCategoryMutation = useMutation({
    mutationFn: async ({ groupId, req }: { groupId: string; req: CreateCategoryRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<Category>(`/groups/${groupId}/categories`, req);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories', groupId] });
    },
  });

  const selectedCategory = categories.find(c => c.id === value);

  const filteredCategories = categories.filter(c =>
    c.name.toLowerCase().includes(search.toLowerCase())
  );

  const exactMatch = categories.some(c => c.name.toLowerCase() === search.toLowerCase().trim());
  const canCreate = allowCreate && search.trim() && !exactMatch && !createCategoryMutation.isPending;

  const handleCreate = () => {
    const name = search.trim();
    if (!name || !groupId) return;
    createCategoryMutation.mutate(
      { groupId, req: { name, description: '', color: '#3b82f6' } },
      {
        onSuccess: (newCat) => {
          onChange(newCat.id);
          setIsOpen(false);
          setSearch('');
        },
      },
    );
  };

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [isOpen]);

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-label={t('expenses.category')}
        className="w-full px-3 py-2 min-h-[44px] border border-gray-300 rounded-md text-left bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:hover:bg-gray-600"
      >
        {selectedCategory ? (
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: selectedCategory.color }}></div>
            <span>{selectedCategory.name}</span>
          </div>
        ) : (
          <span className="text-gray-500">{placeholder ?? t('categories.selectCategory')}</span>
        )}
      </button>

      {isOpen && (
        <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg dark:bg-gray-800 dark:border-gray-600" role="listbox">
          <div className="p-2 border-b border-gray-200 dark:border-gray-700">
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter' && canCreate) { e.preventDefault(); handleCreate(); } }}
              placeholder={allowCreate ? t('categories.searchOrCreate') : t('categories.searchCategories')}
              aria-label={t('categories.searchCategories')}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
              onClick={(e) => e.stopPropagation()}
              autoFocus
            />
          </div>
          <div className="max-h-60 overflow-y-auto">
            {filteredCategories.map((category) => (
              <button
                key={category.id}
                type="button"
                role="option"
                aria-selected={category.id === value}
                onClick={() => {
                  onChange(category.id);
                  setIsOpen(false);
                  setSearch('');
                }}
                className="w-full px-3 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
              >
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: category.color }}></div>
                <span className="dark:text-white">{category.name}</span>
              </button>
            ))}
            {canCreate && (
              <button
                type="button"
                onClick={handleCreate}
                className="w-full px-3 py-2 text-left hover:bg-blue-50 dark:hover:bg-blue-900 flex items-center gap-2 border-t border-gray-200 dark:border-gray-700"
              >
                <span className="text-blue-600 dark:text-blue-400 font-medium">+ {t('categories.createCategory')}: "{search.trim()}"</span>
              </button>
            )}
            {filteredCategories.length === 0 && !canCreate && (
              <div className="px-3 py-2 text-gray-500 text-sm">
                {allowCreate ? t('categories.typeToCreate') : t('categories.noCategories')}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
