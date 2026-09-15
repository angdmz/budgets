import { useTranslation } from 'react-i18next';
import { formatCurrency } from '../lib/format';

interface BudgetSummaryProps {
  expectedTotal: number;
  actualTotal: number;
  difference: number;
  currency?: string;
}

export default function BudgetSummary({ expectedTotal, actualTotal, difference, currency = 'USD' }: BudgetSummaryProps) {
  const { t } = useTranslation();

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
      <div className="rounded-lg bg-white dark:bg-gray-800 p-4 shadow-sm ring-1 ring-gray-200 dark:ring-gray-700">
        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">{t('budgetDetail.expectedTotal')}</p>
        <p className="mt-1 text-xl font-semibold text-gray-900 dark:text-white sm:text-2xl">
          {formatCurrency(String(expectedTotal), currency)}
        </p>
      </div>
      <div className="rounded-lg bg-white dark:bg-gray-800 p-4 shadow-sm ring-1 ring-gray-200 dark:ring-gray-700">
        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">{t('budgetDetail.actualTotal')}</p>
        <p className="mt-1 text-xl font-semibold text-gray-900 dark:text-white sm:text-2xl">
          {formatCurrency(String(actualTotal), currency)}
        </p>
      </div>
      <div className="rounded-lg bg-white dark:bg-gray-800 p-4 shadow-sm ring-1 ring-gray-200 dark:ring-gray-700">
        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">{t('budgetDetail.difference')}</p>
        <p className={`mt-1 text-xl font-semibold sm:text-2xl ${difference >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
          {formatCurrency(String(difference), currency)}
        </p>
      </div>
    </div>
  );
}
