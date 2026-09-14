import { useTranslation } from 'react-i18next';
import CategoryCombobox from './CategoryCombobox';
import CurrencySelect from './CurrencySelect';

interface ExpenseFormFieldsProps {
  name: string;
  amount: string;
  currency: string;
  description: string;
  categoryId: string;
  groupId: string;
  expenseDate?: string;
  showDate?: boolean;
  categoryError?: boolean;
  getAccessTokenSilently: () => Promise<string>;
  onNameChange: (name: string) => void;
  onAmountChange: (amount: string) => void;
  onCurrencyChange: (currency: string) => void;
  onDescriptionChange: (description: string) => void;
  onCategoryIdChange: (categoryId: string) => void;
  onExpenseDateChange?: (date: string) => void;
  onCategoryErrorClear?: () => void;
}

export default function ExpenseFormFields({
  name,
  amount,
  currency,
  description,
  categoryId,
  groupId,
  expenseDate,
  showDate = false,
  categoryError = false,
  getAccessTokenSilently,
  onNameChange,
  onAmountChange,
  onCurrencyChange,
  onDescriptionChange,
  onCategoryIdChange,
  onExpenseDateChange,
  onCategoryErrorClear,
}: ExpenseFormFieldsProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-4">
      <div>
        <label htmlFor="expense-name" className="block text-sm font-medium text-gray-700">{t('common.name')}</label>
        <input
          id="expense-name"
          type="text"
          required
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
        />
      </div>
      <div>
        <label htmlFor="expense-amount" className="block text-sm font-medium text-gray-700">{t('expenses.amount')}</label>
        <div className="mt-1 flex gap-2">
          <input
            id="expense-amount"
            type="number"
            step="0.01"
            required
            value={amount}
            onChange={(e) => onAmountChange(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          />
          <div className="w-28">
            <CurrencySelect value={currency} onChange={onCurrencyChange} />
          </div>
        </div>
      </div>
      {showDate && expenseDate !== undefined && onExpenseDateChange && (
        <div>
          <label htmlFor="expense-date" className="block text-sm font-medium text-gray-700">{t('expenses.date')}</label>
          <input
            id="expense-date"
            type="date"
            required
            value={expenseDate}
            onChange={(e) => onExpenseDateChange(e.target.value)}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          />
        </div>
      )}
      <div>
        <label htmlFor="expense-description" className="block text-sm font-medium text-gray-700">{t('common.description')}</label>
        <input
          id="expense-description"
          type="text"
          value={description}
          onChange={(e) => onDescriptionChange(e.target.value)}
          className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700">{t('expenses.category')}</label>
        <CategoryCombobox
          groupId={groupId}
          value={categoryId}
          onChange={(id) => {
            onCategoryIdChange(id);
            onCategoryErrorClear?.();
          }}
          getAccessTokenSilently={getAccessTokenSilently}
        />
        {categoryError && <p role="alert" className="mt-1 text-sm text-red-600">{t('categories.categoryRequired')}</p>}
      </div>
    </div>
  );
}
