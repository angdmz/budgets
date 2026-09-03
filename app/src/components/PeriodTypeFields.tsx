import { useTranslation } from 'react-i18next';
import { computePeriodDates, type BudgetPeriodType } from '../lib/budgetPeriod';

interface PeriodTypeFieldsProps {
  periodType: BudgetPeriodType;
  onPeriodTypeChange: (periodType: BudgetPeriodType) => void;
  startDate: string;
  endDate: string;
  onDatesChange: (startDate: string, endDate: string) => void;
  referenceDate?: Date;
}

const PERIOD_OPTIONS: BudgetPeriodType[] = ['weekly', 'biweekly', 'monthly', 'custom'];

export default function PeriodTypeFields({
  periodType,
  onPeriodTypeChange,
  startDate,
  endDate,
  onDatesChange,
  referenceDate,
}: PeriodTypeFieldsProps) {
  const { t } = useTranslation();

  const handlePeriodTypeChange = (newType: BudgetPeriodType) => {
    onPeriodTypeChange(newType);
    if (newType !== 'custom') {
      const dates = computePeriodDates(newType, referenceDate);
      onDatesChange(dates.start_date, dates.end_date);
    }
  };

  return (
    <div>
      <label className="block text-sm font-medium text-gray-700">{t('budgets.periodType')}</label>
      <div className="mt-1 grid grid-cols-2 gap-2 sm:grid-cols-4">
        {PERIOD_OPTIONS.map((option) => (
          <button
            key={option}
            type="button"
            onClick={() => handlePeriodTypeChange(option)}
            className={`rounded-md px-3 py-2 text-sm font-medium border ${
              periodType === option
                ? 'bg-primary-600 text-white border-primary-600'
                : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
            }`}
          >
            {t(`budgets.periodTypes.${option}`)}
          </button>
        ))}
      </div>

      {periodType === 'custom' ? (
        <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label className="block text-sm font-medium text-gray-700">{t('budgets.startDate')}</label>
            <input
              type="date"
              required
              value={startDate}
              onChange={(e) => onDatesChange(e.target.value, endDate)}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700">{t('budgets.endDate')}</label>
            <input
              type="date"
              required
              value={endDate}
              onChange={(e) => onDatesChange(startDate, e.target.value)}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
            />
          </div>
        </div>
      ) : (
        <p className="mt-2 text-sm text-gray-500">
          {startDate && endDate ? `${startDate} → ${endDate}` : ''}
        </p>
      )}
    </div>
  );
}
