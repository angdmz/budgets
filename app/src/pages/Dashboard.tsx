import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { createApiClient } from '../lib/api';
import type { Budget, ExpectedExpense, ActualExpense } from '../lib/types';

export default function Dashboard() {
  const { getAccessTokenSilently } = useAuth0();
  const [selectedBudgetId, setSelectedBudgetId] = useState<string>('');

  const { data: budgets } = useQuery({
    queryKey: ['budgets'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Budget[]>('/budgets');
      return response.data;
    },
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

  const calculateTotals = () => {
    const expectedTotal = expectedExpenses?.reduce((sum, exp) => sum + parseFloat(exp.amount.amount), 0) || 0;
    const actualTotal = actualExpenses?.reduce((sum, exp) => sum + parseFloat(exp.amount.amount), 0) || 0;
    const difference = expectedTotal - actualTotal;

    return { expectedTotal, actualTotal, difference };
  };

  const { expectedTotal, actualTotal, difference } = calculateTotals();

  const chartData = [
    {
      name: 'Budget Overview',
      Expected: expectedTotal,
      Actual: actualTotal,
    },
  ];

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
          <p className="mt-2 text-sm text-gray-700">
            Overview of your budget performance
          </p>
        </div>
      </div>

      {/* Budget Selector */}
      <div className="mt-6">
        <label htmlFor="budget" className="block text-sm font-medium text-gray-700">
          Select Budget
        </label>
        <select
          id="budget"
          value={selectedBudgetId}
          onChange={(e) => setSelectedBudgetId(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
        >
          <option value="">Select a budget...</option>
          {budgets?.map((budget) => (
            <option key={budget.id} value={budget.id}>
              {budget.name} ({new Date(budget.start_date).toLocaleDateString()} - {new Date(budget.end_date).toLocaleDateString()})
            </option>
          ))}
        </select>
      </div>

      {selectedBudgetId && (
        <>
          {/* Summary Cards */}
          <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-3">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <span className="text-2xl">💰</span>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">Expected</dt>
                      <dd className="text-lg font-semibold text-gray-900">
                        ${expectedTotal.toFixed(2)}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <span className="text-2xl">💸</span>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">Actual</dt>
                      <dd className="text-lg font-semibold text-gray-900">
                        ${actualTotal.toFixed(2)}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <span className="text-2xl">{difference >= 0 ? '✅' : '⚠️'}</span>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">Difference</dt>
                      <dd className={`text-lg font-semibold ${difference >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        ${Math.abs(difference).toFixed(2)} {difference >= 0 ? 'under' : 'over'}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Chart */}
          <div className="mt-6 bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Budget vs Actual</h2>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="Expected" fill="#0ea5e9" />
                <Bar dataKey="Actual" fill="#f59e0b" />
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Recent Expenses */}
          <div className="mt-6 bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Recent Expenses</h2>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead>
                  <tr>
                    <th className="px-6 py-3 bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Name
                    </th>
                    <th className="px-6 py-3 bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Date
                    </th>
                    <th className="px-6 py-3 bg-gray-50 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Amount
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {actualExpenses?.slice(0, 5).map((expense) => (
                    <tr key={expense.id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {expense.name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(expense.expense_date).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${parseFloat(expense.amount.amount).toFixed(2)} {expense.amount.currency}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
