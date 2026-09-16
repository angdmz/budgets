import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import type { ActualExpense, Category, CreateActualExpenseRequest, UpdateActualExpenseRequest } from '../lib/types';
import Dialog from '../components/Dialog';
import ExpenseFormDialog from '../components/ExpenseFormDialog';
import ExpenseList from '../components/ExpenseList';
import { useBudgetSelection } from '../hooks/useBudgetSelection';

export default function Expenses() {
  const { getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const {
    selectedGroupId,
    selectedBudgetId,
    setSelectedGroupId,
    setSelectedBudgetId,
    groups,
    budgets,
    actualExpenses: expenses,
  } = useBudgetSelection();
  const [formData, setFormData] = useState<CreateActualExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    expense_date: new Date().toISOString().split('T')[0],
    category_id: '',
  });
  const [editingExpense, setEditingExpense] = useState<ActualExpense | null>(null);
  const [deletingExpense, setDeletingExpense] = useState<ActualExpense | null>(null);
  const [categoryError, setCategoryError] = useState(false);

  const { data: categories = [] } = useQuery({
    queryKey: ['categories', selectedGroupId],
    queryFn: async () => {
      if (!selectedGroupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Category[]>(`/groups/${selectedGroupId}/categories`);
      return response.data;
    },
    enabled: !!selectedGroupId,
  });

  const createMutation = useMutation({
    mutationFn: async (data: CreateActualExpenseRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/budgets/${selectedBudgetId}/actual-expenses`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setIsModalOpen(false);
      setFormData({
        name: '',
        description: '',
        amount: { amount: '', currency: 'USD' },
        expense_date: new Date().toISOString().split('T')[0],
        category_id: '',
      });
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateActualExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/actual-expenses/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setEditingExpense(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/actual-expenses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setDeletingExpense(null);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.category_id) {
      setCategoryError(true);
      return;
    }
    setCategoryError(false);
    createMutation.mutate(formData);
  };

  const handleEdit = (expense: ActualExpense) => {
    setEditingExpense(expense);
  };

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingExpense) {
      const updateData: UpdateActualExpenseRequest = {
        name: editingExpense.name,
        description: editingExpense.description,
        amount: editingExpense.amount,
        expense_date: editingExpense.expense_date,
        category_id: editingExpense.category_id,
      };
      updateMutation.mutate({ id: editingExpense.id, data: updateData });
    }
  };

  const handleDelete = () => {
    if (deletingExpense) {
      deleteMutation.mutate(deletingExpense.id);
    }
  };

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">{t('expenses.title')}</h1>
          <p className="mt-2 text-sm text-gray-700 dark:text-gray-300">{t('expenses.subtitle')}</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedBudgetId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            {t('expenses.addExpense')}
          </button>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label className="form-label">{t('common.selectGroup')}</label>
          <select
            value={selectedGroupId}
            onChange={(e) => { setSelectedGroupId(e.target.value); setSelectedBudgetId(''); }}
            className="form-select"
          >
            <option value="">{t('common.selectGroupPlaceholder')}</option>
            {groups?.map((group) => (
              <option key={group.id} value={group.id}>{group.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="form-label">{t('common.selectBudget')}</label>
          <select
            value={selectedBudgetId}
            onChange={(e) => setSelectedBudgetId(e.target.value)}
            disabled={!selectedGroupId}
            className="form-select disabled:opacity-50"
          >
            <option value="">{t('common.selectBudgetPlaceholder')}</option>
            {budgets?.map((budget) => (
              <option key={budget.id} value={budget.id}>{budget.name}</option>
            ))}
          </select>
        </div>
      </div>

      {selectedBudgetId && (
        <div className="mt-8">
          <ExpenseList
            expenses={expenses ?? []}
            categories={categories}
            showDate
            showCategory
            showDescription
            showActions
            onEdit={handleEdit}
            onDelete={(expense) => setDeletingExpense(expense)}
          />
        </div>
      )}

      {isModalOpen && (
        <ExpenseFormDialog
          title={t('expenses.addExpense')}
          onClose={() => { setIsModalOpen(false); createMutation.reset(); }}
          onSubmit={handleSubmit}
          name={formData.name}
          amount={formData.amount.amount}
          currency={formData.amount.currency}
          description={formData.description}
          categoryId={formData.category_id}
          groupId={selectedGroupId}
          expenseDate={formData.expense_date}
          showDate
          categoryError={categoryError}
          isPending={createMutation.isPending}
          error={createMutation.isError ? t('expenses.createError') : undefined}
          errorMessageDetail={createMutation.isError ? getErrorMessage(createMutation.error) : undefined}
          submitLabel={t('common.create')}
          submitPendingLabel={`${t('common.create')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setFormData(prev => ({ ...prev, name }))}
          onAmountChange={(amount) => setFormData(prev => ({ ...prev, amount: { ...prev.amount, amount } }))}
          onCurrencyChange={(currency) => setFormData(prev => ({ ...prev, amount: { ...prev.amount, currency } }))}
          onDescriptionChange={(description) => setFormData(prev => ({ ...prev, description }))}
          onCategoryIdChange={(categoryId) => setFormData(prev => ({ ...prev, category_id: categoryId }))}
          onExpenseDateChange={(expense_date) => setFormData(prev => ({ ...prev, expense_date }))}
          onCategoryErrorClear={() => setCategoryError(false)}
        />
      )}

      {editingExpense && (
        <ExpenseFormDialog
          title={t('expenses.editExpense')}
          onClose={() => { setEditingExpense(null); updateMutation.reset(); }}
          onSubmit={handleUpdate}
          name={editingExpense.name}
          amount={editingExpense.amount.amount}
          currency={editingExpense.amount.currency}
          description={editingExpense.description}
          categoryId={editingExpense.category_id}
          groupId={selectedGroupId}
          expenseDate={editingExpense.expense_date}
          showDate
          isPending={updateMutation.isPending}
          error={updateMutation.isError ? t('expenses.updateError') : undefined}
          errorMessageDetail={updateMutation.isError ? getErrorMessage(updateMutation.error) : undefined}
          submitLabel={t('common.update')}
          submitPendingLabel={`${t('common.update')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setEditingExpense(prev => prev ? { ...prev, name } : null)}
          onAmountChange={(amount) => setEditingExpense(prev => prev ? { ...prev, amount: { ...prev.amount, amount } } : null)}
          onCurrencyChange={(currency) => setEditingExpense(prev => prev ? { ...prev, amount: { ...prev.amount, currency } } : null)}
          onDescriptionChange={(description) => setEditingExpense(prev => prev ? { ...prev, description } : null)}
          onCategoryIdChange={(categoryId) => setEditingExpense(prev => prev ? { ...prev, category_id: categoryId } : null)}
          onExpenseDateChange={(expense_date) => setEditingExpense(prev => prev ? { ...prev, expense_date } : null)}
        />
      )}

      {deletingExpense && (
        <Dialog title={t('expenses.deleteExpense')} onClose={() => { setDeletingExpense(null); deleteMutation.reset(); }}>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            {t('common.deleteConfirm', { name: deletingExpense.name }).replace(/\*\*/g, '')}
          </p>
          {deleteMutation.isError && (
            <p className="mb-4 text-sm text-red-600">
              {t('expenses.deleteError')}
              {getErrorMessage(deleteMutation.error) && (
                <span className="block text-xs mt-1 opacity-75">{getErrorMessage(deleteMutation.error)}</span>
              )}
            </p>
          )}
          <div className="flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
            <button
              type="button"
              onClick={() => { setDeletingExpense(null); deleteMutation.reset(); }}
              className="rounded-md bg-white dark:bg-gray-700 px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 min-h-[44px]"
            >
              {t('common.cancel')}
            </button>
            <button
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
              className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 disabled:opacity-50 min-h-[44px]"
            >
              {deleteMutation.isPending ? `${t('common.delete')}...` : t('common.delete')}
            </button>
          </div>
        </Dialog>
      )}
    </div>
  );
}
