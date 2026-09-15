import { useTranslation } from 'react-i18next';
import Dialog from './Dialog';
import ExpenseFormFields from './ExpenseFormFields';

interface ExpenseFormDialogProps {
  title: string;
  onClose: () => void;
  onSubmit: (e: React.FormEvent) => void;
  name: string;
  amount: string;
  currency: string;
  description: string;
  categoryId: string;
  groupId: string;
  expenseDate?: string;
  showDate?: boolean;
  categoryError?: boolean;
  isPending?: boolean;
  error?: string;
  errorMessageDetail?: string;
  submitLabel: string;
  submitPendingLabel: string;
  getAccessTokenSilently: () => Promise<string>;
  onNameChange: (name: string) => void;
  onAmountChange: (amount: string) => void;
  onCurrencyChange: (currency: string) => void;
  onDescriptionChange: (description: string) => void;
  onCategoryIdChange: (categoryId: string) => void;
  onExpenseDateChange?: (date: string) => void;
  onCategoryErrorClear?: () => void;
}

export default function ExpenseFormDialog({
  title,
  onClose,
  onSubmit,
  name,
  amount,
  currency,
  description,
  categoryId,
  groupId,
  expenseDate,
  showDate = false,
  categoryError = false,
  isPending = false,
  error,
  errorMessageDetail,
  submitLabel,
  submitPendingLabel,
  getAccessTokenSilently,
  onNameChange,
  onAmountChange,
  onCurrencyChange,
  onDescriptionChange,
  onCategoryIdChange,
  onExpenseDateChange,
  onCategoryErrorClear,
}: ExpenseFormDialogProps) {
  const { t } = useTranslation();

  return (
    <Dialog title={title} onClose={onClose}>
      <form onSubmit={onSubmit}>
        <ExpenseFormFields
          name={name}
          amount={amount}
          currency={currency}
          description={description}
          categoryId={categoryId}
          groupId={groupId}
          expenseDate={expenseDate}
          showDate={showDate}
          categoryError={categoryError}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={onNameChange}
          onAmountChange={onAmountChange}
          onCurrencyChange={onCurrencyChange}
          onDescriptionChange={onDescriptionChange}
          onCategoryIdChange={onCategoryIdChange}
          onExpenseDateChange={onExpenseDateChange}
          onCategoryErrorClear={onCategoryErrorClear}
        />
        {error && (
          <p role="alert" className="mt-2 text-sm text-red-600">
            {error}
            {errorMessageDetail && <span className="block text-xs mt-1 opacity-75">{errorMessageDetail}</span>}
          </p>
        )}
        <div className="mt-6 flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md bg-white dark:bg-gray-700 px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 min-h-[44px]"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 min-h-[44px]"
          >
            {isPending ? submitPendingLabel : submitLabel}
          </button>
        </div>
      </form>
    </Dialog>
  );
}
