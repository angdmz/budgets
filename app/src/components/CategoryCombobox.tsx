import { useState, useRef, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { createApiClient } from '../lib/api';
import type { Category } from '../lib/types';

interface CategoryComboboxProps {
  groupId: string;
  value?: string;
  onChange: (categoryId: string | undefined) => void;
  getAccessTokenSilently: () => Promise<string>;
}

export default function CategoryCombobox({ groupId, value, onChange, getAccessTokenSilently }: CategoryComboboxProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState('');
  const containerRef = useRef<HTMLDivElement>(null);

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

  const selectedCategory = categories.find(c => c.id === value);

  const filteredCategories = categories.filter(c =>
    c.name.toLowerCase().includes(search.toLowerCase())
  );

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
        className="w-full px-3 py-2 border border-gray-300 rounded-md text-left bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        {selectedCategory ? (
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: selectedCategory.color }}></div>
            <span>{selectedCategory.name}</span>
          </div>
        ) : (
          <span className="text-gray-400">Select category (optional)</span>
        )}
      </button>

      {isOpen && (
        <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg">
          <div className="p-2 border-b border-gray-200">
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search categories..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              onClick={(e) => e.stopPropagation()}
            />
          </div>
          <div className="max-h-60 overflow-y-auto">
            <button
              type="button"
              onClick={() => {
                onChange(undefined);
                setIsOpen(false);
                setSearch('');
              }}
              className="w-full px-3 py-2 text-left hover:bg-gray-100 text-gray-500"
            >
              None
            </button>
            {filteredCategories.map((category) => (
              <button
                key={category.id}
                type="button"
                onClick={() => {
                  onChange(category.id);
                  setIsOpen(false);
                  setSearch('');
                }}
                className="w-full px-3 py-2 text-left hover:bg-gray-100 flex items-center gap-2"
              >
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: category.color }}></div>
                <span>{category.name}</span>
              </button>
            ))}
            {filteredCategories.length === 0 && search && (
              <div className="px-3 py-2 text-gray-500 text-sm">No categories found</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
