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
      <label className="form-label">{t('budgets.periodType')}</label>
      <div className="mt-1 grid grid-cols-2 gap-2 sm:grid-cols-4">
        {PERIOD_OPTIONS.map((option) => (
          <button
            key={option}
            type="button"
            onClick={() => handlePeriodTypeChange(option)}
            className={`rounded-md px-3 py-2 text-sm font-medium border ${
              periodType === option
                ? 'bg-primary-600 text-white border-primary-600'
                : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600'
            }`}
          >
            {t(`budgets.periodTypes.${option}`)}
          </button>
        ))}
      </div>

      {periodType === 'custom' ? (
        <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label className="form-label">{t('budgets.startDate')}</label>
            <input
              type="date"
              required
              value={startDate}
              onChange={(e) => onDatesChange(e.target.value, endDate)}
              className="form-input"
            />
          </div>
          <div>
            <label className="form-label">{t('budgets.endDate')}</label>
            <input
              type="date"
              required
              value={endDate}
              onChange={(e) => onDatesChange(startDate, e.target.value)}
              className="form-input"
            />
          </div>
        </div>
      ) : (
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          {startDate && endDate ? `${startDate} → ${endDate}` : ''}
        </p>
      )}
    </div>
  );
}
