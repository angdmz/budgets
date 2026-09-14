import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import { formatDate } from '../lib/format';
import type {
  Budget,
  Group,
  Category,
  ExpectedExpense,
  ActualExpense,
  CreateExpectedExpenseRequest,
  UpdateExpectedExpenseRequest,
  CreateActualExpenseRequest,
  UpdateActualExpenseRequest,
} from '../lib/types';
import Dialog from '../components/Dialog';
import ExpenseFormDialog from '../components/ExpenseFormDialog';
import ExpenseList from '../components/ExpenseList';
import BudgetSummary from '../components/BudgetSummary';

export default function BudgetDetail() {
  const { budgetId } = useParams<{ budgetId: string }>();
  const navigate = useNavigate();
  const { getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [eeModalOpen, setEeModalOpen] = useState(false);
  const [eeEditing, setEeEditing] = useState<ExpectedExpense | null>(null);
  const [eeDeleting, setEeDeleting] = useState<ExpectedExpense | null>(null);
  const [eeForm, setEeForm] = useState<CreateExpectedExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    category_id: '',
  });
  const [eeCategoryError, setEeCategoryError] = useState(false);

  const [aeModalOpen, setAeModalOpen] = useState(false);
  const [aeEditing, setAeEditing] = useState<ActualExpense | null>(null);
  const [aeDeleting, setAeDeleting] = useState<ActualExpense | null>(null);
  const [aeForm, setAeForm] = useState<CreateActualExpenseRequest>({
    name: '',
    description: '',
    amount: { amount: '', currency: 'USD' },
    expense_date: new Date().toISOString().split('T')[0],
    category_id: '',
  });
  const [aeCategoryError, setAeCategoryError] = useState(false);

  const { data: budgetInfo } = useQuery({
    queryKey: ['budget', budgetId],
    queryFn: async () => {
      if (!budgetId) return null;
      const api = await createApiClient(getAccessTokenSilently);
      const groupsRes = await api.get<Group[]>('/groups');
      for (const group of groupsRes.data) {
        const budgetsRes = await api.get<Budget[]>(`/groups/${group.id}/budgets`);
        const found = budgetsRes.data.find((b) => b.id === budgetId);
        if (found) return { budget: found, groupId: group.id, groupName: group.name };
      }
      return null;
    },
    enabled: !!budgetId,
  });

  const groupId = budgetInfo?.groupId ?? '';

  const { data: categories = [] } = useQuery({
    queryKey: ['categories', groupId],
    queryFn: async () => {
      if (!groupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Category[]>(`/groups/${groupId}/categories`);
      return response.data;
    },
    enabled: !!groupId,
  });

  const { data: expectedExpenses } = useQuery({
    queryKey: ['expected-expenses', budgetId],
    queryFn: async () => {
      if (!budgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ExpectedExpense[]>(`/budgets/${budgetId}/expected-expenses`);
      return response.data;
    },
    enabled: !!budgetId,
  });

  const { data: actualExpenses } = useQuery({
    queryKey: ['actual-expenses', budgetId],
    queryFn: async () => {
      if (!budgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ActualExpense[]>(`/budgets/${budgetId}/actual-expenses`);
      return response.data;
    },
    enabled: !!budgetId,
  });

  // ── Expected expense mutations ──
  const eeCreateMut = useMutation({
    mutationFn: async (data: CreateExpectedExpenseRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/budgets/${budgetId}/expected-expenses`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setEeModalOpen(false);
      setEeForm({ name: '', description: '', amount: { amount: '', currency: 'USD' }, category_id: '' });
    },
  });

  const eeUpdateMut = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateExpectedExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/expected-expenses/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setEeEditing(null);
    },
  });

  const eeDeleteMut = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/expected-expenses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses'] });
      setEeDeleting(null);
    },
  });

  // ── Actual expense mutations ──
  const aeCreateMut = useMutation({
    mutationFn: async (data: CreateActualExpenseRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/budgets/${budgetId}/actual-expenses`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setAeModalOpen(false);
      setAeForm({
        name: '',
        description: '',
        amount: { amount: '', currency: 'USD' },
        expense_date: new Date().toISOString().split('T')[0],
        category_id: '',
      });
    },
  });

  const aeUpdateMut = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateActualExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/actual-expenses/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setAeEditing(null);
    },
  });

  const aeDeleteMut = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/actual-expenses/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses'] });
      setAeDeleting(null);
    },
  });

  // ── Handlers ──
  const handleEeSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!eeForm.category_id) { setEeCategoryError(true); return; }
    setEeCategoryError(false);
    eeCreateMut.mutate(eeForm);
  };

  const handleEeUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (eeEditing) {
      eeUpdateMut.mutate({
        id: eeEditing.id,
        data: {
          name: eeEditing.name,
          description: eeEditing.description,
          amount: eeEditing.amount,
          category_id: eeEditing.category_id,
        },
      });
    }
  };

  const handleAeSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!aeForm.category_id) { setAeCategoryError(true); return; }
    setAeCategoryError(false);
    aeCreateMut.mutate(aeForm);
  };

  const handleAeUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (aeEditing) {
      aeUpdateMut.mutate({
        id: aeEditing.id,
        data: {
          name: aeEditing.name,
          description: aeEditing.description,
          amount: aeEditing.amount,
          expense_date: aeEditing.expense_date,
          category_id: aeEditing.category_id,
        },
      });
    }
  };

  const expectedTotal = expectedExpenses?.reduce((s, e) => s + parseFloat(e.amount.amount), 0) ?? 0;
  const actualTotal = actualExpenses?.reduce((s, e) => s + parseFloat(e.amount.amount), 0) ?? 0;
  const currency = expectedExpenses?.[0]?.amount.currency ?? actualExpenses?.[0]?.amount.currency ?? 'USD';

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      {/* Header */}
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <button
            onClick={() => navigate('/budgets')}
            className="text-sm text-primary-600 hover:text-primary-900 mb-2"
          >
            ← {t('budgetDetail.backToBudgets')}
          </button>
          <h1 className="text-2xl font-semibold text-gray-900">{budgetInfo?.budget.name ?? '...'}</h1>
          <p className="mt-1 text-sm text-gray-700">
            {budgetInfo && `${formatDate(budgetInfo.budget.start_date)} - ${formatDate(budgetInfo.budget.end_date)}`}
            {budgetInfo?.budget.description && ` · ${budgetInfo.budget.description}`}
          </p>
        </div>
      </div>

      {/* Summary cards */}
      <div className="mt-6">
        <BudgetSummary expectedTotal={expectedTotal} actualTotal={actualTotal} difference={expectedTotal - actualTotal} currency={currency} />
      </div>

      {/* Expected Expenses section */}
      <div className="mt-8">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">{t('budgetDetail.expectedExpenses')}</h2>
          <button
            onClick={() => setEeModalOpen(true)}
            className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
          >
            {t('expectedExpenses.addExpectedExpense')}
          </button>
        </div>
        <div className="mt-4">
          <ExpenseList
            expenses={expectedExpenses ?? []}
            categories={categories}
            showCategory
            showDescription
            showActions
            onEdit={setEeEditing}
            onDelete={setEeDeleting}
            emptyMessage={t('budgetDetail.noExpectedExpenses')}
          />
        </div>
      </div>

      {/* Actual Expenses section */}
      <div className="mt-8">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">{t('budgetDetail.actualExpenses')}</h2>
          <button
            onClick={() => setAeModalOpen(true)}
            className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
          >
            {t('expenses.addExpense')}
          </button>
        </div>
        <div className="mt-4">
          <ExpenseList
            expenses={actualExpenses ?? []}
            categories={categories}
            showDate
            showCategory
            showDescription
            showActions
            onEdit={setAeEditing}
            onDelete={setAeDeleting}
            emptyMessage={t('budgetDetail.noActualExpenses')}
          />
        </div>
      </div>

      {/* ── Expected Expense Create Modal ── */}
      {eeModalOpen && (
        <ExpenseFormDialog
          title={t('expectedExpenses.addExpectedExpense')}
          onClose={() => { setEeModalOpen(false); eeCreateMut.reset(); }}
          onSubmit={handleEeSubmit}
          name={eeForm.name}
          amount={eeForm.amount.amount}
          currency={eeForm.amount.currency}
          description={eeForm.description}
          categoryId={eeForm.category_id}
          groupId={groupId}
          showDate={false}
          categoryError={eeCategoryError}
          isPending={eeCreateMut.isPending}
          error={eeCreateMut.isError ? t('expectedExpenses.createError') : undefined}
          errorMessageDetail={eeCreateMut.isError ? getErrorMessage(eeCreateMut.error) : undefined}
          submitLabel={t('common.create')}
          submitPendingLabel={`${t('common.create')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setEeForm((p) => ({ ...p, name }))}
          onAmountChange={(amount) => setEeForm((p) => ({ ...p, amount: { ...p.amount, amount } }))}
          onCurrencyChange={(currency) => setEeForm((p) => ({ ...p, amount: { ...p.amount, currency } }))}
          onDescriptionChange={(description) => setEeForm((p) => ({ ...p, description }))}
          onCategoryIdChange={(categoryId) => setEeForm((p) => ({ ...p, category_id: categoryId }))}
          onCategoryErrorClear={() => setEeCategoryError(false)}
        />
      )}

      {/* ── Expected Expense Edit Modal ── */}
      {eeEditing && (
        <ExpenseFormDialog
          title={t('expectedExpenses.editExpectedExpense')}
          onClose={() => { setEeEditing(null); eeUpdateMut.reset(); }}
          onSubmit={handleEeUpdate}
          name={eeEditing.name}
          amount={eeEditing.amount.amount}
          currency={eeEditing.amount.currency}
          description={eeEditing.description}
          categoryId={eeEditing.category_id}
          groupId={groupId}
          showDate={false}
          isPending={eeUpdateMut.isPending}
          error={eeUpdateMut.isError ? t('expectedExpenses.updateError') : undefined}
          errorMessageDetail={eeUpdateMut.isError ? getErrorMessage(eeUpdateMut.error) : undefined}
          submitLabel={t('common.update')}
          submitPendingLabel={`${t('common.update')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setEeEditing((p) => p ? { ...p, name } : p)}
          onAmountChange={(amount) => setEeEditing((p) => p ? { ...p, amount: { ...p.amount, amount } } : p)}
          onCurrencyChange={(currency) => setEeEditing((p) => p ? { ...p, amount: { ...p.amount, currency } } : p)}
          onDescriptionChange={(description) => setEeEditing((p) => p ? { ...p, description } : p)}
          onCategoryIdChange={(categoryId) => setEeEditing((p) => p ? { ...p, category_id: categoryId } : p)}
        />
      )}

      {/* ── Expected Expense Delete Modal ── */}
      {eeDeleting && (
        <Dialog title={t('expectedExpenses.deleteExpectedExpense')} onClose={() => { setEeDeleting(null); eeDeleteMut.reset(); }}>
            <p className="text-sm text-gray-500 mb-4">{t('common.deleteConfirm', { name: eeDeleting.name }).replace(/\*\*/g, '')}</p>
            {eeDeleteMut.isError && (
              <p className="mb-4 text-sm text-red-600">{t('expectedExpenses.deleteError')}
                {getErrorMessage(eeDeleteMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(eeDeleteMut.error)}</span>}
              </p>
            )}
            <div className="flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
              <button type="button" onClick={() => { setEeDeleting(null); eeDeleteMut.reset(); }}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 min-h-[44px]">{t('common.cancel')}</button>
              <button type="button" onClick={() => eeDeleteMut.mutate(eeDeleting.id)} disabled={eeDeleteMut.isPending}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 disabled:opacity-50 min-h-[44px]">
                {eeDeleteMut.isPending ? `${t('common.delete')}...` : t('common.delete')}
              </button>
            </div>
        </Dialog>
      )}

      {/* ── Actual Expense Create Modal ── */}
      {aeModalOpen && (
        <ExpenseFormDialog
          title={t('expenses.addExpense')}
          onClose={() => { setAeModalOpen(false); aeCreateMut.reset(); }}
          onSubmit={handleAeSubmit}
          name={aeForm.name}
          amount={aeForm.amount.amount}
          currency={aeForm.amount.currency}
          description={aeForm.description}
          categoryId={aeForm.category_id}
          groupId={groupId}
          expenseDate={aeForm.expense_date}
          showDate
          categoryError={aeCategoryError}
          isPending={aeCreateMut.isPending}
          error={aeCreateMut.isError ? t('expenses.createError') : undefined}
          errorMessageDetail={aeCreateMut.isError ? getErrorMessage(aeCreateMut.error) : undefined}
          submitLabel={t('common.create')}
          submitPendingLabel={`${t('common.create')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setAeForm((p) => ({ ...p, name }))}
          onAmountChange={(amount) => setAeForm((p) => ({ ...p, amount: { ...p.amount, amount } }))}
          onCurrencyChange={(currency) => setAeForm((p) => ({ ...p, amount: { ...p.amount, currency } }))}
          onDescriptionChange={(description) => setAeForm((p) => ({ ...p, description }))}
          onCategoryIdChange={(categoryId) => setAeForm((p) => ({ ...p, category_id: categoryId }))}
          onExpenseDateChange={(expense_date) => setAeForm((p) => ({ ...p, expense_date }))}
          onCategoryErrorClear={() => setAeCategoryError(false)}
        />
      )}

      {/* ── Actual Expense Edit Modal ── */}
      {aeEditing && (
        <ExpenseFormDialog
          title={t('expenses.editExpense')}
          onClose={() => { setAeEditing(null); aeUpdateMut.reset(); }}
          onSubmit={handleAeUpdate}
          name={aeEditing.name}
          amount={aeEditing.amount.amount}
          currency={aeEditing.amount.currency}
          description={aeEditing.description}
          categoryId={aeEditing.category_id}
          groupId={groupId}
          expenseDate={aeEditing.expense_date}
          showDate
          isPending={aeUpdateMut.isPending}
          error={aeUpdateMut.isError ? t('expenses.updateError') : undefined}
          errorMessageDetail={aeUpdateMut.isError ? getErrorMessage(aeUpdateMut.error) : undefined}
          submitLabel={t('common.update')}
          submitPendingLabel={`${t('common.update')}...`}
          getAccessTokenSilently={getAccessTokenSilently}
          onNameChange={(name) => setAeEditing((p) => p ? { ...p, name } : p)}
          onAmountChange={(amount) => setAeEditing((p) => p ? { ...p, amount: { ...p.amount, amount } } : p)}
          onCurrencyChange={(currency) => setAeEditing((p) => p ? { ...p, amount: { ...p.amount, currency } } : p)}
          onDescriptionChange={(description) => setAeEditing((p) => p ? { ...p, description } : p)}
          onCategoryIdChange={(categoryId) => setAeEditing((p) => p ? { ...p, category_id: categoryId } : p)}
          onExpenseDateChange={(expense_date) => setAeEditing((p) => p ? { ...p, expense_date } : p)}
        />
      )}

      {/* ── Actual Expense Delete Modal ── */}
      {aeDeleting && (
        <Dialog title={t('expenses.deleteExpense')} onClose={() => { setAeDeleting(null); aeDeleteMut.reset(); }}>
            <p className="text-sm text-gray-500 mb-4">{t('common.deleteConfirm', { name: aeDeleting.name }).replace(/\*\*/g, '')}</p>
            {aeDeleteMut.isError && (
              <p className="mb-4 text-sm text-red-600">{t('expenses.deleteError')}
                {getErrorMessage(aeDeleteMut.error) && <span className="block text-xs mt-1 opacity-75">{getErrorMessage(aeDeleteMut.error)}</span>}
              </p>
            )}
            <div className="flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
              <button type="button" onClick={() => { setAeDeleting(null); aeDeleteMut.reset(); }}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 min-h-[44px]">{t('common.cancel')}</button>
              <button type="button" onClick={() => aeDeleteMut.mutate(aeDeleting.id)} disabled={aeDeleteMut.isPending}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 disabled:opacity-50 min-h-[44px]">
                {aeDeleteMut.isPending ? `${t('common.delete')}...` : t('common.delete')}
              </button>
            </div>
        </Dialog>
      )}
    </div>
  );
}
