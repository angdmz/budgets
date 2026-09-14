import { useTranslation } from 'react-i18next';
import { formatDate, formatCurrency } from '../lib/format';
import type { Category, Money } from '../lib/types';

export interface ExpenseListItem {
  id: string;
  name: string;
  description: string;
  amount: Money;
  category_id: string;
  expense_date?: string;
}

interface ExpenseListProps<T extends ExpenseListItem> {
  expenses: T[];
  categories?: Category[];
  showDate?: boolean;
  showCategory?: boolean;
  showDescription?: boolean;
  showActions?: boolean;
  onEdit?: (expense: T) => void;
  onDelete?: (expense: T) => void;
  emptyMessage?: string;
  maxRows?: number;
}

export default function ExpenseList<T extends ExpenseListItem>({
  expenses,
  categories = [],
  showDate = false,
  showCategory = false,
  showDescription = false,
  showActions = false,
  onEdit,
  onDelete,
  emptyMessage,
  maxRows,
}: ExpenseListProps<T>) {
  const { t } = useTranslation();
  const rows = maxRows ? expenses.slice(0, maxRows) : expenses;
  const colCount = 1 + (showDate ? 1 : 0) + 1 + (showCategory ? 1 : 0) + (showDescription ? 1 : 0) + (showActions ? 1 : 0);

  return (
    <div className="overflow-x-auto shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
      <table className="min-w-full divide-y divide-gray-300">
        <thead className="bg-gray-50">
          <tr>
            <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{t('common.name')}</th>
            {showDate && (
              <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.date')}</th>
            )}
            <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.amount')}</th>
            {showCategory && (
              <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('expenses.category')}</th>
            )}
            {showDescription && (
              <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{t('common.description')}</th>
            )}
            {showActions && (
              <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                <span className="sr-only">{t('common.actions')}</span>
              </th>
            )}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200 bg-white">
          {rows.map((expense) => {
            const category = categories.find((c) => c.id === expense.category_id);
            return (
              <tr key={expense.id}>
                <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                  {expense.name}
                </td>
                {showDate && (
                  <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                    {expense.expense_date ? formatDate(expense.expense_date) : '-'}
                  </td>
                )}
                <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                  {formatCurrency(expense.amount.amount, expense.amount.currency)}
                </td>
                {showCategory && (
                  <td className="px-3 py-4 text-sm text-gray-500">
                    {category ? (
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: category.color }}></div>
                        <span>{category.name}</span>
                      </div>
                    ) : '-'}
                  </td>
                )}
                {showDescription && (
                  <td className="px-3 py-4 text-sm text-gray-500">{expense.description || '-'}</td>
                )}
                {showActions && (
                  <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                    <button
                      onClick={() => onEdit?.(expense)}
                      aria-label={`${t('common.edit')} ${expense.name}`}
                      className="text-blue-600 hover:text-blue-900 mr-4 min-h-[44px]"
                    >
                      {t('common.edit')}
                    </button>
                    <button
                      onClick={() => onDelete?.(expense)}
                      aria-label={`${t('common.delete')} ${expense.name}`}
                      className="text-red-600 hover:text-red-900 min-h-[44px]"
                    >
                      {t('common.delete')}
                    </button>
                  </td>
                )}
              </tr>
            );
          })}
          {rows.length === 0 && emptyMessage && (
            <tr>
              <td colSpan={colCount} className="py-6 text-center text-sm text-gray-500">{emptyMessage}</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
