import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { createApiClient } from '../lib/api';
import type { ActualExpense, Budget, CreateActualExpenseRequest } from '../lib/types';

export default function Expenses() {
  const { getAccessTokenSilently } = useAuth0();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedBudgetId, setSelectedBudgetId] = useState('');
  const [formData, setFormData] = useState<CreateActualExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    expense_date: new Date().toISOString().split('T')[0],
  });

  const { data: budgets } = useQuery({
    queryKey: ['budgets'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Budget[]>('/budgets');
      return response.data;
    },
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
      });
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
          <h1 className="text-2xl font-semibold text-gray-900">Expenses</h1>
          <p className="mt-2 text-sm text-gray-700">Track your actual expenses</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedBudgetId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            Add Expense
          </button>
        </div>
      </div>

      <div className="mt-6">
        <label className="block text-sm font-medium text-gray-700">Select Budget</label>
        <select
          value={selectedBudgetId}
          onChange={(e) => setSelectedBudgetId(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
        >
          <option value="">Select a budget...</option>
          {budgets?.map((budget) => (
            <option key={budget.id} value={budget.id}>{budget.name}</option>
          ))}
        </select>
      </div>

      {selectedBudgetId && (
        <div className="mt-8 flow-root">
          <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300">
              <thead className="bg-gray-50">
                <tr>
                  <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">Name</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Date</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Amount</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Description</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {expenses?.map((expense) => (
                  <tr key={expense.id}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                      {expense.name}
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {new Date(expense.expense_date).toLocaleDateString()}
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      ${parseFloat(expense.amount.amount).toFixed(2)} {expense.amount.currency}
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500">{expense.description || '-'}</td>
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
            <h2 className="text-lg font-semibold mb-4">Add Expense</h2>
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
                  <label className="block text-sm font-medium text-gray-700">Amount</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={formData.amount.amount}
                    onChange={(e) => setFormData({ ...formData, amount: { ...formData.amount, amount: e.target.value } })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Date</label>
                  <input
                    type="date"
                    required
                    value={formData.expense_date}
                    onChange={(e) => setFormData({ ...formData, expense_date: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
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
