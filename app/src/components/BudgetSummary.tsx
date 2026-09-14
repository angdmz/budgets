import { useTranslation } from 'react-i18next';
import { formatCurrency } from '../lib/format';

interface BudgetSummaryProps {
  expectedTotal: number;
  actualTotal: number;
  difference: number;
}

export default function BudgetSummary({ expectedTotal, actualTotal, difference }: BudgetSummaryProps) {
  const { t } = useTranslation();

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
      <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-gray-200">
        <p className="text-sm font-medium text-gray-500">{t('budgetDetail.expectedTotal')}</p>
        <p className="mt-1 text-xl font-semibold text-gray-900 sm:text-2xl">
          {formatCurrency(String(expectedTotal), 'USD')}
        </p>
      </div>
      <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-gray-200">
        <p className="text-sm font-medium text-gray-500">{t('budgetDetail.actualTotal')}</p>
        <p className="mt-1 text-xl font-semibold text-gray-900 sm:text-2xl">
          {formatCurrency(String(actualTotal), 'USD')}
        </p>
      </div>
      <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-gray-200">
        <p className="text-sm font-medium text-gray-500">{t('budgetDetail.difference')}</p>
        <p className={`mt-1 text-xl font-semibold sm:text-2xl ${difference >= 0 ? 'text-green-600' : 'text-red-600'}`}>
          {formatCurrency(String(difference), 'USD')}
        </p>
      </div>
    </div>
  );
}
