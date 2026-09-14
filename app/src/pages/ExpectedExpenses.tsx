import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import type { ExpectedExpense, Category, CreateExpectedExpenseRequest, UpdateExpectedExpenseRequest } from '../lib/types';
import Dialog from '../components/Dialog';
import ExpenseFormDialog from '../components/ExpenseFormDialog';
import ExpenseList from '../components/ExpenseList';
import { useBudgetSelection } from '../hooks/useBudgetSelection';

export default function ExpectedExpenses() {
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
    expectedExpenses: expenses,
  } = useBudgetSelection();
  const [formData, setFormData] = useState<CreateExpectedExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    category_id: '',
  });
  const [editingExpense, setEditingExpense] = useState<ExpectedExpense | null>(null);
  const [deletingExpense, setDeletingExpense] = useState<ExpectedExpense | null>(null);
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
    mutationFn: async (data: CreateExpectedExpenseRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/budgets/${selectedBudgetId}/expected-expenses`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setIsModalOpen(false);
      setFormData({
        name: '',
        description: '',
        category_id: '',
        amount: { amount: '', currency: 'USD' },
      });
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateExpectedExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/expected-expenses/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setEditingExpense(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/expected-expenses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
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

  const handleEdit = (expense: ExpectedExpense) => {
    setEditingExpense(expense);
  };

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingExpense) {
      const updateData: UpdateExpectedExpenseRequest = {
        name: editingExpense.name,
        description: editingExpense.description,
        amount: editingExpense.amount,
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
          <h1 className="text-2xl font-semibold text-gray-900">{t('expectedExpenses.title')}</h1>
          <p className="mt-2 text-sm text-gray-700">{t('expectedExpenses.subtitle')}</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedBudgetId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            {t('expectedExpenses.addExpectedExpense')}
          </button>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label className="block text-sm font-medium text-gray-700">{t('common.selectGroup')}</label>
          <select
            value={selectedGroupId}
            onChange={(e) => { setSelectedGroupId(e.target.value); setSelectedBudgetId(''); }}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          >
            <option value="">{t('common.selectGroupPlaceholder')}</option>
            {groups?.map((group) => (
              <option key={group.id} value={group.id}>{group.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700">{t('common.selectBudget')}</label>
          <select
            value={selectedBudgetId}
            onChange={(e) => setSelectedBudgetId(e.target.value)}
            disabled={!selectedGroupId}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm disabled:opacity-50"
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
            showCategory
            showDescription
            showActions
            onEdit={handleEdit}
            onDelete={setDeletingExpense}
            emptyMessage={t('expectedExpenses.noExpenses')}
          />
        </div>
      )}

      {isModalOpen && (
        <ExpenseFormDialog
          title={t('expectedExpenses.addExpectedExpense')}
          onClose={() => { setIsModalOpen(false); createMutation.reset(); }}
          onSubmit={handleSubmit}
          name={formData.name}
          amount={formData.amount.amount}
          currency={formData.amount.currency}
          description={formData.description}
          categoryId={formData.category_id}
          groupId={selectedGroupId}
          showDate={false}
          categoryError={categoryError}
          isPending={createMutation.isPending}
          error={createMutation.isError ? t('expectedExpenses.createError') : undefined}
          errorMessageDetail={createMutation.isError ? getErrorMessage(createMutation.error) : undefined}
          submitLabel={t('common.create')}
          submitPendingLabel={`${t('common.create')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setFormData(prev => ({ ...prev, name }))}
          onAmountChange={(amount) => setFormData(prev => ({ ...prev, amount: { ...prev.amount, amount } }))}
          onCurrencyChange={(currency) => setFormData(prev => ({ ...prev, amount: { ...prev.amount, currency } }))}
          onDescriptionChange={(description) => setFormData(prev => ({ ...prev, description }))}
          onCategoryIdChange={(categoryId) => setFormData(prev => ({ ...prev, category_id: categoryId }))}
          onCategoryErrorClear={() => setCategoryError(false)}
        />
      )}

      {editingExpense && (
        <ExpenseFormDialog
          title={t('expectedExpenses.editExpectedExpense')}
          onClose={() => { setEditingExpense(null); updateMutation.reset(); }}
          onSubmit={handleUpdate}
          name={editingExpense.name}
          amount={editingExpense.amount.amount}
          currency={editingExpense.amount.currency}
          description={editingExpense.description}
          categoryId={editingExpense.category_id}
          groupId={selectedGroupId}
          showDate={false}
          isPending={updateMutation.isPending}
          error={updateMutation.isError ? t('expectedExpenses.updateError') : undefined}
          errorMessageDetail={updateMutation.isError ? getErrorMessage(updateMutation.error) : undefined}
          submitLabel={t('common.update')}
          submitPendingLabel={`${t('common.update')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setEditingExpense(prev => prev ? { ...prev, name } : prev)}
          onAmountChange={(amount) => setEditingExpense(prev => prev ? { ...prev, amount: { ...prev.amount, amount } } : prev)}
          onCurrencyChange={(currency) => setEditingExpense(prev => prev ? { ...prev, amount: { ...prev.amount, currency } } : prev)}
          onDescriptionChange={(description) => setEditingExpense(prev => prev ? { ...prev, description } : prev)}
          onCategoryIdChange={(categoryId) => setEditingExpense(prev => prev ? { ...prev, category_id: categoryId } : prev)}
        />
      )}

      {deletingExpense && (
        <Dialog title={t('expectedExpenses.deleteExpectedExpense')} onClose={() => { setDeletingExpense(null); deleteMutation.reset(); }}>
            <p className="text-sm text-gray-500 mb-4">
              {t('common.deleteConfirm', { name: deletingExpense.name }).replace(/\*\*/g, '')}
            </p>
            {deleteMutation.isError && (
              <p className="mb-4 text-sm text-red-600">
                {t('expectedExpenses.deleteError')}
                {getErrorMessage(deleteMutation.error) && (
                  <span className="block text-xs mt-1 opacity-75">{getErrorMessage(deleteMutation.error)}</span>
                )}
              </p>
            )}
            <div className="flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
              <button
                type="button"
                onClick={() => { setDeletingExpense(null); deleteMutation.reset(); }}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 min-h-[44px]"
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
