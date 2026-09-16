import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import { formatDate } from '../lib/format';
import type { Budget, Group, CreateBudgetRequest, UpdateBudgetRequest } from '../lib/types';
import type { BudgetPeriodType, MonthOption } from '../lib/budgetPeriod';
import PeriodTypeFields from '../components/PeriodTypeFields';
import MonthPicker from '../components/MonthPicker';
import { useBudgetService } from '../hooks/useBudgetService';
import type { CreateRecurringProgress, CreateBudgetsForMonthsProgress } from '../lib/services/budgetService';
import Dialog from '../components/Dialog';

export default function Budgets() {
  const navigate = useNavigate();
  const { getAccessTokenSilently } = useAuth0();
  const { t, i18n } = useTranslation();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [formData, setFormData] = useState<CreateBudgetRequest>({
    name: '',
    description: '',
    start_date: '',
    end_date: '',
  });
  const [periodType, setPeriodType] = useState<BudgetPeriodType>('custom');
  const [isRecurring, setIsRecurring] = useState(false);
  const [numberOfPeriods, setNumberOfPeriods] = useState(3);
  const [recurringProgress, setRecurringProgress] = useState<CreateRecurringProgress | null>(null);
  const [editingBudget, setEditingBudget] = useState<Budget | null>(null);
  const [deletingBudget, setDeletingBudget] = useState<Budget | null>(null);
  const [duplicatingBudget, setDuplicatingBudget] = useState<Budget | null>(null);
  const [duplicateFormData, setDuplicateFormData] = useState<CreateBudgetRequest>({
    name: '',
    description: '',
    start_date: '',
    end_date: '',
  });
  const [duplicatePeriodType, setDuplicatePeriodType] = useState<BudgetPeriodType>('custom');
  const [duplicateCopyExpenses, setDuplicateCopyExpenses] = useState(true);
  const [duplicateIsRecurring, setDuplicateIsRecurring] = useState(false);
  const [duplicateNumberOfPeriods, setDuplicateNumberOfPeriods] = useState(3);
  const [duplicateSelectedMonths, setDuplicateSelectedMonths] = useState<MonthOption[]>([]);
  const [monthsProgress, setMonthsProgress] = useState<CreateBudgetsForMonthsProgress | null>(null);

  const { duplicateBudget, createRecurringBudgets, createBudgetsForMonths } = useBudgetService();

  const { data: groups } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Group[]>('/groups');
      return response.data;
    },
  });

  const { data: budgets } = useQuery({
    queryKey: ['budgets', selectedGroupId],
    queryFn: async () => {
      if (!selectedGroupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Budget[]>(`/groups/${selectedGroupId}/budgets`);
      return response.data;
    },
    enabled: !!selectedGroupId,
  });

  const createMutation = useMutation({
    mutationFn: async (data: CreateBudgetRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post(`/groups/${selectedGroupId}/budgets`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      closeCreateModal();
    },
  });

  const closeCreateModal = () => {
    setIsModalOpen(false);
    setFormData({ name: '', description: '', start_date: '', end_date: '' });
    setPeriodType('custom');
    setIsRecurring(false);
    setNumberOfPeriods(3);
    setRecurringProgress(null);
    createMutation.reset();
    createRecurringBudgets.reset();
  };

  const closeDuplicateModal = () => {
    setDuplicatingBudget(null);
    setDuplicateFormData({ name: '', description: '', start_date: '', end_date: '' });
    setDuplicatePeriodType('custom');
    setDuplicateCopyExpenses(true);
    setDuplicateIsRecurring(false);
    setDuplicateNumberOfPeriods(3);
    setDuplicateSelectedMonths([]);
    setRecurringProgress(null);
    setMonthsProgress(null);
    duplicateBudget.reset();
    createRecurringBudgets.reset();
    createBudgetsForMonths.reset();
  };

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateBudgetRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.put(`/budgets/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      setEditingBudget(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/budgets/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['budgets'] });
      setDeletingBudget(null);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (isRecurring && periodType !== 'custom') {
      createRecurringBudgets.mutate(
        {
          options: {
            groupId: selectedGroupId,
            baseName: formData.name,
            description: formData.description,
            periodType,
            firstStartDate: formData.start_date,
            firstEndDate: formData.end_date,
            numberOfPeriods,
            copyExpectedExpenses: false,
          },
          onProgress: setRecurringProgress,
        },
        { onSuccess: closeCreateModal }
      );
      return;
    }
    createMutation.mutate(formData);
  };

  const handleEdit = (budget: Budget) => {
    setEditingBudget(budget);
  };

  const handleDuplicateClick = (budget: Budget) => {
    setDuplicatingBudget(budget);
    setDuplicateFormData({
      name: `${budget.name} (Copy)`,
      description: budget.description,
      start_date: budget.start_date,
      end_date: budget.end_date,
    });
    setDuplicateSelectedMonths([]);
  };

  const handleDuplicateSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!duplicatingBudget) return;

    if (duplicateIsRecurring && duplicatePeriodType === 'monthly') {
      createBudgetsForMonths.mutate(
        {
          options: {
            groupId: selectedGroupId,
            baseName: duplicateFormData.name,
            description: duplicateFormData.description,
            months: duplicateSelectedMonths,
            sourceBudgetId: duplicatingBudget.id,
            copyExpectedExpenses: duplicateCopyExpenses,
            locale: i18n.language,
          },
          onProgress: setMonthsProgress,
        },
        { onSuccess: closeDuplicateModal }
      );
      return;
    }

    if (duplicateIsRecurring && duplicatePeriodType !== 'custom') {
      createRecurringBudgets.mutate(
        {
          options: {
            groupId: selectedGroupId,
            baseName: duplicateFormData.name,
            description: duplicateFormData.description,
            periodType: duplicatePeriodType,
            firstStartDate: duplicateFormData.start_date,
            firstEndDate: duplicateFormData.end_date,
            numberOfPeriods: duplicateNumberOfPeriods,
            sourceBudgetId: duplicatingBudget.id,
            copyExpectedExpenses: duplicateCopyExpenses,
          },
          onProgress: setRecurringProgress,
        },
        { onSuccess: closeDuplicateModal }
      );
      return;
    }

    duplicateBudget.mutate(
      {
        sourceBudgetId: duplicatingBudget.id,
        groupId: selectedGroupId,
        name: duplicateFormData.name,
        description: duplicateFormData.description,
        startDate: duplicateFormData.start_date,
        endDate: duplicateFormData.end_date,
        copyExpectedExpenses: duplicateCopyExpenses,
      },
      { onSuccess: closeDuplicateModal }
    );
  };

  const handleUpdate = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingBudget) {
      const updateData: UpdateBudgetRequest = {
        name: editingBudget.name,
        description: editingBudget.description,
        start_date: editingBudget.start_date,
        end_date: editingBudget.end_date,
      };
      updateMutation.mutate({ id: editingBudget.id, data: updateData });
    }
  };

  const handleDelete = () => {
    if (deletingBudget) {
      deleteMutation.mutate(deletingBudget.id);
    }
  };

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">{t('budgets.title')}</h1>
          <p className="mt-2 text-sm text-gray-700 dark:text-gray-300">{t('budgets.subtitle')}</p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none flex flex-wrap gap-2">
          <button
            onClick={() => navigate('/budgets/new')}
            disabled={!selectedGroupId}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
          >
            {t('budgetPlan.createPlan')}
          </button>
          <button
            onClick={() => setIsModalOpen(true)}
            disabled={!selectedGroupId}
            className="block rounded-md bg-white dark:bg-gray-700 px-3 py-2 text-center text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50"
          >
            {t('budgets.addBudget')}
          </button>
        </div>
      </div>

      <div className="mt-6">
        <label className="form-label">{t('common.selectGroup')}</label>
        <select
          value={selectedGroupId}
          onChange={(e) => setSelectedGroupId(e.target.value)}
          className="form-select"
        >
          <option value="">{t('common.selectGroupPlaceholder')}</option>
          {groups?.map((group) => (
            <option key={group.id} value={group.id}>
              {group.name}
            </option>
          ))}
        </select>
      </div>

      {selectedGroupId && (
        <div className="mt-8 flow-root">
          <div className="overflow-x-auto shadow ring-1 ring-black ring-opacity-5 dark:ring-white/10 sm:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-900/50">
                <tr>
                  <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-white sm:pl-6">{t('common.name')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">{t('budgets.period')}</th>
                  <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">{t('common.description')}</th>
                  <th className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span className="sr-only">{t('common.actions')}</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-800">
                {budgets?.map((budget) => (
                  <tr key={budget.id}>
                    <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 dark:text-white sm:pl-6">
                      <button onClick={() => navigate(`/budgets/${budget.id}`)} className="text-left hover:text-primary-600 dark:hover:text-primary-400">
                        {budget.name}
                      </button>
                    </td>
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {formatDate(budget.start_date)} - {formatDate(budget.end_date)}
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-500 dark:text-gray-400">{budget.description || '-'}</td>
                    <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                      <button
                        onClick={() => handleEdit(budget)}
                        className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-4 min-h-[44px]"
                      >
                        {t('common.edit')}
                      </button>
                      <button
                        onClick={() => handleDuplicateClick(budget)}
                        className="text-primary-600 hover:text-primary-900 dark:text-primary-400 dark:hover:text-primary-300 mr-4 min-h-[44px]"
                      >
                        {t('budgets.duplicate')}
                      </button>
                      <button
                        onClick={() => setDeletingBudget(budget)}
                        className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 min-h-[44px]"
                      >
                        {t('common.delete')}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {isModalOpen && (
        <Dialog title={t('budgets.createBudget')} onClose={closeCreateModal}>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="form-label">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="form-input"
                  />
                </div>
                <PeriodTypeFields
                  periodType={periodType}
                  onPeriodTypeChange={setPeriodType}
                  startDate={formData.start_date}
                  endDate={formData.end_date}
                  onDatesChange={(start_date, end_date) => setFormData({ ...formData, start_date, end_date })}
                />
                {periodType !== 'custom' && (
                  <div>
                    <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                      <input
                        type="checkbox"
                        checked={isRecurring}
                        onChange={(e) => setIsRecurring(e.target.checked)}
                        className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
                      />
                      {t('budgets.createRecurring')}
                    </label>
                    {isRecurring && (
                      <div className="mt-2">
                        <label className="form-label">{t('budgets.numberOfPeriods')}</label>
                        <input
                          type="number"
                          min={2}
                          max={24}
                          value={numberOfPeriods}
                          onChange={(e) => setNumberOfPeriods(Number(e.target.value))}
                          className="form-input w-24"
                        />
                      </div>
                    )}
                  </div>
                )}
              </div>
              {createMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.createError')}
                  {getErrorMessage(createMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createMutation.error)}</span>
                  )}
                </p>
              )}
              {createRecurringBudgets.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.createError')}
                  {getErrorMessage(createRecurringBudgets.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createRecurringBudgets.error)}</span>
                  )}
                </p>
              )}
              {recurringProgress && createRecurringBudgets.isPending && (
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  {t('budgets.recurringProgress', {
                    current: recurringProgress.currentPeriod,
                    total: recurringProgress.totalPeriods,
                  })}
                </p>
              )}
              <div className="mt-6 flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
                <button
                  type="button"
                  onClick={closeCreateModal}
                  className="rounded-md bg-white dark:bg-gray-700 px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 min-h-[44px]"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending || createRecurringBudgets.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 min-h-[44px]"
                >
                  {createMutation.isPending || createRecurringBudgets.isPending
                    ? `${t('common.create')}...`
                    : t('common.create')}
                </button>
              </div>
            </form>
        </Dialog>
      )}

      {editingBudget && (
        <Dialog title={t('budgets.editBudget')} onClose={() => { setEditingBudget(null); updateMutation.reset(); }}>
            <form onSubmit={handleUpdate}>
              <div className="space-y-4">
                <div>
                  <label className="form-label">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={editingBudget.name}
                    onChange={(e) => setEditingBudget({ ...editingBudget, name: e.target.value })}
                    className="form-input"
                  />
                </div>
                <div>
                  <label className="form-label">{t('budgets.startDate')}</label>
                  <input
                    type="date"
                    required
                    value={editingBudget.start_date}
                    onChange={(e) => setEditingBudget({ ...editingBudget, start_date: e.target.value })}
                    className="form-input"
                  />
                </div>
                <div>
                  <label className="form-label">{t('budgets.endDate')}</label>
                  <input
                    type="date"
                    required
                    value={editingBudget.end_date}
                    onChange={(e) => setEditingBudget({ ...editingBudget, end_date: e.target.value })}
                    className="form-input"
                  />
                </div>
              </div>
              {updateMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.updateError')}
                  {getErrorMessage(updateMutation.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(updateMutation.error)}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
                <button
                  type="button"
                  onClick={() => { setEditingBudget(null); updateMutation.reset(); }}
                  className="rounded-md bg-white dark:bg-gray-700 px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 min-h-[44px]"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={updateMutation.isPending}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 min-h-[44px]"
                >
                  {updateMutation.isPending ? `${t('common.update')}...` : t('common.update')}
                </button>
              </div>
            </form>
        </Dialog>
      )}

      {duplicatingBudget && (
        <Dialog title={t('budgets.duplicateBudget')} onClose={closeDuplicateModal}>
            <form onSubmit={handleDuplicateSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="form-label">{t('common.name')}</label>
                  <input
                    type="text"
                    required
                    value={duplicateFormData.name}
                    onChange={(e) => setDuplicateFormData({ ...duplicateFormData, name: e.target.value })}
                    className="form-input"
                  />
                </div>
                <PeriodTypeFields
                  periodType={duplicatePeriodType}
                  onPeriodTypeChange={setDuplicatePeriodType}
                  startDate={duplicateFormData.start_date}
                  endDate={duplicateFormData.end_date}
                  onDatesChange={(start_date, end_date) =>
                    setDuplicateFormData({ ...duplicateFormData, start_date, end_date })
                  }
                />
                <div>
                  <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <input
                      type="checkbox"
                      checked={duplicateCopyExpenses}
                      onChange={(e) => setDuplicateCopyExpenses(e.target.checked)}
                      className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
                    />
                    {t('budgets.copyExpectedExpenses')}
                  </label>
                </div>
                {duplicatePeriodType !== 'custom' && (
                  <div>
                    <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                      <input
                        type="checkbox"
                        checked={duplicateIsRecurring}
                        onChange={(e) => setDuplicateIsRecurring(e.target.checked)}
                        className="rounded border-gray-300 dark:border-gray-600 text-primary-600 focus:ring-primary-500"
                      />
                      {t('budgets.createRecurring')}
                    </label>
                    {duplicateIsRecurring && duplicatePeriodType === 'monthly' && (
                      <div className="mt-2">
                        <MonthPicker
                          year={
                            duplicateFormData.start_date
                              ? new Date(`${duplicateFormData.start_date}T00:00:00`).getFullYear()
                              : new Date().getFullYear()
                          }
                          selectedMonths={duplicateSelectedMonths}
                          onChange={setDuplicateSelectedMonths}
                          referenceDate={
                            duplicateFormData.start_date ? new Date(`${duplicateFormData.start_date}T00:00:00`) : undefined
                          }
                        />
                      </div>
                    )}
                    {duplicateIsRecurring && duplicatePeriodType !== 'monthly' && (
                      <div className="mt-2">
                        <label className="form-label">{t('budgets.numberOfPeriods')}</label>
                        <input
                          type="number"
                          min={2}
                          max={24}
                          value={duplicateNumberOfPeriods}
                          onChange={(e) => setDuplicateNumberOfPeriods(Number(e.target.value))}
                          className="form-input w-24"
                        />
                      </div>
                    )}
                  </div>
                )}
              </div>
              {duplicateBudget.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.duplicateError')}
                  {getErrorMessage(duplicateBudget.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(duplicateBudget.error)}</span>
                  )}
                </p>
              )}
              {createRecurringBudgets.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.duplicateError')}
                  {getErrorMessage(createRecurringBudgets.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createRecurringBudgets.error)}</span>
                  )}
                </p>
              )}
              {createBudgetsForMonths.isError && (
                <p className="mt-2 text-sm text-red-600">
                  {t('budgets.duplicateError')}
                  {getErrorMessage(createBudgetsForMonths.error) && (
                    <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createBudgetsForMonths.error)}</span>
                  )}
                </p>
              )}
              {recurringProgress && createRecurringBudgets.isPending && (
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  {t('budgets.recurringProgress', {
                    current: recurringProgress.currentPeriod,
                    total: recurringProgress.totalPeriods,
                  })}
                </p>
              )}
              {monthsProgress && createBudgetsForMonths.isPending && (
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  {t('budgets.recurringProgress', {
                    current: monthsProgress.currentIndex,
                    total: monthsProgress.totalMonths,
                  })}
                </p>
              )}
              <div className="mt-6 flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
                <button
                  type="button"
                  onClick={closeDuplicateModal}
                  className="rounded-md bg-white dark:bg-gray-700 px-3 py-2 text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 min-h-[44px]"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={
                    duplicateBudget.isPending ||
                    createRecurringBudgets.isPending ||
                    createBudgetsForMonths.isPending ||
                    (duplicateIsRecurring && duplicatePeriodType === 'monthly' && duplicateSelectedMonths.length === 0)
                  }
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 min-h-[44px]"
                >
                  {duplicateBudget.isPending || createRecurringBudgets.isPending || createBudgetsForMonths.isPending
                    ? `${t('budgets.duplicate')}...`
                    : t('budgets.duplicate')}
                </button>
              </div>
            </form>
        </Dialog>
      )}

      {deletingBudget && (
        <Dialog title={t('budgets.deleteBudget')} onClose={() => { setDeletingBudget(null); deleteMutation.reset(); }}>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
              {t('common.deleteConfirm', { name: deletingBudget.name }).replace(/\*\*/g, '')}
            </p>
            {deleteMutation.isError && (
              <p className="mb-4 text-sm text-red-600">
                {t('budgets.deleteError')}
                {getErrorMessage(deleteMutation.error) && (
                  <span className="block text-xs mt-1 opacity-75">{getErrorMessage(deleteMutation.error)}</span>
                )}
              </p>
            )}
            <div className="flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
              <button
                type="button"
                onClick={() => { setDeletingBudget(null); deleteMutation.reset(); }}
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
