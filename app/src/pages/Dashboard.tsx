import { useTranslation } from 'react-i18next';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { formatDate } from '../lib/format';
import { useBudgetSelection } from '../hooks/useBudgetSelection';
import BudgetSummary from '../components/BudgetSummary';
import ExpenseList from '../components/ExpenseList';

export default function Dashboard() {
  const { t } = useTranslation();
  const {
    selectedGroupId,
    selectedBudgetId,
    setSelectedGroupId,
    setSelectedBudgetId,
    groups,
    budgets,
    actualExpenses,
    expectedTotal,
    actualTotal,
    difference,
    currency,
  } = useBudgetSelection();

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
          <h1 className="text-2xl font-semibold text-gray-900">{t('dashboard.title')}</h1>
          <p className="mt-2 text-sm text-gray-700">
            {t('dashboard.subtitle')}
          </p>
        </div>
      </div>

      {/* Group Selector */}
      <div className="mt-6">
        <label htmlFor="group" className="block text-sm font-medium text-gray-700">
          {t('common.selectGroup')}
        </label>
        <select
          id="group"
          value={selectedGroupId}
          onChange={(e) => { setSelectedGroupId(e.target.value); setSelectedBudgetId(''); }}
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

      {/* Budget Selector */}
      <div className="mt-4">
        <label htmlFor="budget" className="block text-sm font-medium text-gray-700">
          {t('common.selectBudget')}
        </label>
        <select
          id="budget"
          value={selectedBudgetId}
          onChange={(e) => setSelectedBudgetId(e.target.value)}
          disabled={!selectedGroupId}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm disabled:opacity-50"
        >
          <option value="">{t('common.selectBudgetPlaceholder')}</option>
          {budgets?.map((budget) => (
            <option key={budget.id} value={budget.id}>
              {budget.name} ({formatDate(budget.start_date)} - {formatDate(budget.end_date)})
            </option>
          ))}
        </select>
      </div>

      {selectedBudgetId && (
        <>
          {/* Summary Cards */}
          <div className="mt-6">
            <BudgetSummary expectedTotal={expectedTotal} actualTotal={actualTotal} difference={difference} currency={currency} />
          </div>

          {/* Chart */}
          <div className="mt-6 bg-white shadow rounded-lg p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">{t('dashboard.budgetVsActual')}</h2>
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
            <h2 className="text-lg font-medium text-gray-900 mb-4">{t('dashboard.recentExpenses')}</h2>
            <ExpenseList
              expenses={actualExpenses ?? []}
              showDate
              maxRows={5}
            />
          </div>
        </>
      )}
    </div>
  );
}
