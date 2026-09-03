import { useTranslation } from 'react-i18next';
import { quarterMonths, remainingMonthsOfYear, type MonthOption } from '../lib/budgetPeriod';

interface MonthPickerProps {
  year: number;
  selectedMonths: MonthOption[];
  onChange: (months: MonthOption[]) => void;
  referenceDate?: Date;
}

const MONTH_NAMES = [
  'january', 'february', 'march', 'april', 'may', 'june',
  'july', 'august', 'september', 'october', 'november', 'december',
];

function isSelected(months: MonthOption[], year: number, month: number): boolean {
  return months.some((m) => m.year === year && m.month === month);
}

function sortMonths(months: MonthOption[]): MonthOption[] {
  return [...months].sort((a, b) => a.year - b.year || a.month - b.month);
}

export default function MonthPicker({ year, selectedMonths, onChange, referenceDate }: MonthPickerProps) {
  const { t } = useTranslation();

  const toggleMonth = (month: number) => {
    if (isSelected(selectedMonths, year, month)) {
      onChange(selectedMonths.filter((m) => !(m.year === year && m.month === month)));
    } else {
      onChange(sortMonths([...selectedMonths, { year, month }]));
    }
  };

  return (
    <div>
      <label className="block text-sm font-medium text-gray-700">{t('budgets.selectMonths')}</label>
      <div className="mt-1 flex flex-wrap gap-2">
        <button
          type="button"
          onClick={() => onChange(sortMonths(remainingMonthsOfYear(referenceDate)))}
          className="rounded-md px-2 py-1 text-xs font-medium border bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
        >
          {t('budgets.restOfYear')}
        </button>
        {([1, 2, 3, 4] as const).map((q) => (
          <button
            key={q}
            type="button"
            onClick={() => onChange(sortMonths(quarterMonths(q, year)))}
            className="rounded-md px-2 py-1 text-xs font-medium border bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
          >
            {t('budgets.quarter', { quarter: q })}
          </button>
        ))}
        <button
          type="button"
          onClick={() => onChange([])}
          className="rounded-md px-2 py-1 text-xs font-medium border bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
        >
          {t('budgets.clearMonths')}
        </button>
      </div>
      <div className="mt-2 grid grid-cols-3 gap-2 sm:grid-cols-4">
        {MONTH_NAMES.map((name, idx) => {
          const month = idx + 1;
          const selected = isSelected(selectedMonths, year, month);
          return (
            <button
              key={month}
              type="button"
              onClick={() => toggleMonth(month)}
              aria-pressed={selected}
              className={`rounded-md px-3 py-2 text-sm font-medium border ${
                selected
                  ? 'bg-primary-600 text-white border-primary-600'
                  : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
              }`}
            >
              {t(`budgets.months.${name}`)}
            </button>
          );
        })}
      </div>
      <p className="mt-2 text-sm text-gray-500">
        {t('budgets.monthsSelectedCount', { count: selectedMonths.length })}
      </p>
    </div>
  );
}
