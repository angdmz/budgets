import { useState, useMemo, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { createApiClient, getErrorMessage } from '../lib/api';
import { formatCurrency } from '../lib/format';
import {
  computePeriodDates,
  generatePeriods,
  type BudgetPeriodType,
  type GeneratedPeriod,
} from '../lib/budgetPeriod';
import type { Group, Category } from '../lib/types';
import type { BudgetPlanTemplateExpense, CreateBudgetPlanProgress } from '../lib/services/budgetService';
import { useBudgetService } from '../hooks/useBudgetService';
import CategoryCombobox from '../components/CategoryCombobox';
import CurrencySelect from '../components/CurrencySelect';

const DRAFT_KEY = 'budgetPlanDraft';

interface WizardDraft {
  step: number;
  selectedGroupId: string;
  baseName: string;
  description: string;
  cadence: Exclude<BudgetPeriodType, 'custom'>;
  startDate: string;
  numberOfPeriods: number;
  templateExpenses: TemplateExpenseRow[];
}

function loadDraft(): WizardDraft | null {
  try {
    const raw = localStorage.getItem(DRAFT_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as WizardDraft;
  } catch {
    return null;
  }
}

function saveDraft(draft: WizardDraft) {
  try {
    localStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
  } catch {
    // ignore quota errors
  }
}

function clearDraft() {
  try {
    localStorage.removeItem(DRAFT_KEY);
  } catch {
    // ignore
  }
}

const CADENCE_OPTIONS: Exclude<BudgetPeriodType, 'custom'>[] = ['weekly', 'biweekly', 'monthly'];

interface TemplateExpenseRow extends BudgetPlanTemplateExpense {
  id: string;
}

let rowIdCounter = 0;
function nextRowId(): string {
  rowIdCounter += 1;
  return `row-${rowIdCounter}`;
}

function emptyRow(): TemplateExpenseRow {
  return {
    id: nextRowId(),
    name: '',
    description: '',
    amount: '',
    currency: 'USD',
    category_id: '',
  };
}

export default function CreateBudgetPlan() {
  const navigate = useNavigate();
  const { getAccessTokenSilently } = useAuth0();
  const { t, i18n } = useTranslation();
  const { createBudgetPlan } = useBudgetService();
  const queryClient = useQueryClient();

  const [step, setStep] = useState(1);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [baseName, setBaseName] = useState('');
  const [description, setDescription] = useState('');
  const [cadence, setCadence] = useState<Exclude<BudgetPeriodType, 'custom'>>('monthly');
  const [startDate, setStartDate] = useState(() => {
    const today = new Date();
    return `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
  });
  const [numberOfPeriods, setNumberOfPeriods] = useState(3);
  const [templateExpenses, setTemplateExpenses] = useState<TemplateExpenseRow[]>([emptyRow()]);
  const [progress, setProgress] = useState<CreateBudgetPlanProgress | null>(null);
  const [showResumePrompt, setShowResumePrompt] = useState(false);

  // ── Load draft on mount ──
  useEffect(() => {
    const draft = loadDraft();
    if (draft && draft.selectedGroupId && draft.baseName) {
      setShowResumePrompt(true);
    }
  }, []);

  const resumeDraft = useCallback(() => {
    const draft = loadDraft();
    if (draft) {
      setStep(draft.step);
      setSelectedGroupId(draft.selectedGroupId);
      setBaseName(draft.baseName);
      setDescription(draft.description);
      setCadence(draft.cadence);
      setStartDate(draft.startDate);
      setNumberOfPeriods(draft.numberOfPeriods);
      setTemplateExpenses(draft.templateExpenses.length > 0 ? draft.templateExpenses : [emptyRow()]);
    }
    setShowResumePrompt(false);
  }, []);

  const discardDraft = useCallback(() => {
    clearDraft();
    setShowResumePrompt(false);
  }, []);

  // ── Save draft on state change ──
  useEffect(() => {
    if (selectedGroupId || baseName || step > 1) {
      saveDraft({
        step,
        selectedGroupId,
        baseName,
        description,
        cadence,
        startDate,
        numberOfPeriods,
        templateExpenses,
      });
    }
  }, [step, selectedGroupId, baseName, description, cadence, startDate, numberOfPeriods, templateExpenses]);

  const { data: groups } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Group[]>('/groups');
      return response.data;
    },
  });

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

  const periods: GeneratedPeriod[] = useMemo(
    () => generatePeriods(cadence, startDate, numberOfPeriods, i18n.language),
    [cadence, startDate, numberOfPeriods, i18n.language]
  );

  const validTemplateExpenses = useMemo(
    () => templateExpenses.filter((e) => e.name.trim() && e.amount && e.category_id),
    [templateExpenses]
  );

  const step1Valid = selectedGroupId && baseName.trim() && startDate && numberOfPeriods >= 1;
  const step2Valid = true;

  const handleCadenceChange = (newCadence: Exclude<BudgetPeriodType, 'custom'>) => {
    setCadence(newCadence);
    const dates = computePeriodDates(newCadence, new Date(`${startDate}T00:00:00`));
    setStartDate(dates.start_date);
  };

  const addExpenseRow = () => {
    setTemplateExpenses((prev) => [...prev, emptyRow()]);
  };

  const removeExpenseRow = (id: string) => {
    setTemplateExpenses((prev) => (prev.length > 1 ? prev.filter((e) => e.id !== id) : prev));
  };

  const updateExpenseRow = (id: string, patch: Partial<TemplateExpenseRow>) => {
    setTemplateExpenses((prev) => prev.map((e) => (e.id === id ? { ...e, ...patch } : e)));
  };

  const handleCreate = () => {
    setProgress(null);
    // Navigate immediately — creation runs in the background
    navigate('/budgets');
    createBudgetPlan.mutate(
      {
        options: {
          groupId: selectedGroupId,
          baseName: baseName.trim(),
          description: description.trim(),
          cadence,
          startDate,
          numberOfPeriods,
          templateExpenses: validTemplateExpenses.map(({ id: _id, ...rest }) => rest),
          locale: i18n.language,
        },
        onProgress: setProgress,
      },
      {
        onSuccess: () => {
          clearDraft();
          queryClient.invalidateQueries({ queryKey: ['budgets'] });
        },
      }
    );
  };

  const isCreating = createBudgetPlan.isPending;

  return (
    <div className="px-4 sm:px-6 lg:px-8 max-w-2xl mx-auto">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">{t('budgetPlan.title')}</h1>
          <p className="mt-2 text-sm text-gray-700 dark:text-gray-300">{t('budgetPlan.subtitle')}</p>
        </div>
      </div>

      {/* Resume draft prompt */}
      {showResumePrompt && (
        <div className="mt-4 rounded-md bg-blue-50 p-4 ring-1 ring-blue-200 dark:bg-blue-900/30 dark:ring-blue-800">
          <p className="text-sm text-blue-800 dark:text-blue-200">{t('budgetPlan.draftFound')}</p>
          <div className="mt-3 flex gap-3">
            <button
              type="button"
              onClick={resumeDraft}
              className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-blue-500"
            >
              {t('budgetPlan.resumeDraft')}
            </button>
            <button
              type="button"
              onClick={discardDraft}
              className="rounded-md bg-white px-3 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600"
            >
              {t('budgetPlan.discardDraft')}
            </button>
          </div>
        </div>
      )}

      {/* Stepper */}
      <div className="mt-6 flex items-center gap-2 overflow-x-auto pb-2">
        {[1, 2, 3].map((s) => (
          <div key={s} className="flex items-center gap-2">
            <div
              className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold ${
                step >= s ? 'bg-primary-600 text-white' : 'bg-gray-200 text-gray-500 dark:bg-gray-700 dark:text-gray-400'
              }`}
            >
              {s}
            </div>
            <span
              className={`text-sm ${
                step >= s ? 'text-gray-900 font-medium dark:text-white' : 'text-gray-500 dark:text-gray-400'
              }`}
            >
              {t(`budgetPlan.step${s}`)}
            </span>
            {s < 3 && <div className="mx-2 h-px w-8 bg-gray-300 dark:bg-gray-600" />}
          </div>
        ))}
      </div>

      <div className="mt-6 bg-white rounded-lg shadow-sm ring-1 ring-gray-200 p-6 dark:bg-gray-800 dark:ring-gray-700">
        {/* Step 1: Cadence & Periods */}
        {step === 1 && (
          <div className="space-y-4">
            <div>
              <label className="form-label">{t('common.selectGroup')}</label>
              <select
                value={selectedGroupId}
                onChange={(e) => setSelectedGroupId(e.target.value)}
                className="form-select"
              >
                <option value="">{t('common.selectGroupPlaceholder')}</option>
                {groups?.map((group) => (
                  <option key={group.id} value={group.id}>{group.name}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="form-label">{t('common.name')}</label>
              <input
                type="text"
                required
                value={baseName}
                onChange={(e) => setBaseName(e.target.value)}
                placeholder={t('budgetPlan.namePlaceholder')}
                className="form-input"
              />
            </div>

            <div>
              <label className="form-label">{t('common.description')}</label>
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="form-input"
              />
            </div>

            <div>
              <label className="form-label">{t('budgetPlan.cadence')}</label>
              <div className="mt-1 grid grid-cols-3 gap-2">
                {CADENCE_OPTIONS.map((option) => (
                  <button
                    key={option}
                    type="button"
                    onClick={() => handleCadenceChange(option)}
                    className={`rounded-md px-3 py-2 text-sm font-medium border ${
                      cadence === option
                        ? 'bg-primary-600 text-white border-primary-600'
                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600 dark:hover:bg-gray-600'
                    }`}
                  >
                    {t(`budgets.periodTypes.${option}`)}
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label className="form-label">{t('budgets.startDate')}</label>
                <input
                  type="date"
                  required
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="form-input"
                />
              </div>
              <div>
                <label className="form-label">{t('budgetPlan.numberOfPeriods')}</label>
                <input
                  type="number"
                  min={1}
                  max={52}
                  value={numberOfPeriods}
                  onChange={(e) => setNumberOfPeriods(Math.max(1, Number(e.target.value)))}
                  className="form-input"
                />
              </div>
            </div>

            {/* Preview of periods */}
            <div className="rounded-md bg-gray-50 p-3 dark:bg-gray-900/50">
              <p className="text-sm font-medium text-gray-700 mb-2 dark:text-gray-300">{t('budgetPlan.periodPreview')}</p>
              <ul className="space-y-1">
                {periods.slice(0, 6).map((p, i) => (
                  <li key={i} className="text-sm text-gray-600 dark:text-gray-400">
                    {numberOfPeriods > 1 ? `${baseName || '—'} - ${p.label}` : (baseName || '—')}
                    <span className="ml-2 text-gray-400 dark:text-gray-500">({p.start_date} → {p.end_date})</span>
                  </li>
                ))}
                {periods.length > 6 && (
                  <li className="text-sm text-gray-400 dark:text-gray-500">... {periods.length - 6} more</li>
                )}
              </ul>
            </div>

            <div className="flex justify-end">
              <button
                type="button"
                disabled={!step1Valid}
                onClick={() => setStep(2)}
                className="rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
              >
                {t('budgetPlan.next')}
              </button>
            </div>
          </div>
        )}

        {/* Step 2: Template Expenses */}
        {step === 2 && (
          <div className="space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">{t('budgetPlan.templateDescription')}</p>

            {templateExpenses.map((row, index) => (
              <div key={row.id} className="rounded-md border border-gray-200 p-4 space-y-3 dark:border-gray-700">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('budgetPlan.expense')} #{index + 1}</span>
                  {templateExpenses.length > 1 && (
                    <button
                      type="button"
                      onClick={() => removeExpenseRow(row.id)}
                      className="text-sm text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                    >
                      {t('common.delete')}
                    </button>
                  )}
                </div>

                <div>
                  <label className="form-label">{t('common.name')}</label>
                  <input
                    type="text"
                    value={row.name}
                    onChange={(e) => updateExpenseRow(row.id, { name: e.target.value })}
                    className="form-input"
                  />
                </div>

                <div>
                  <label className="form-label">{t('expectedExpenses.amount')}</label>
                  <div className="mt-1 flex gap-2">
                    <input
                      type="number"
                      step="0.01"
                      value={row.amount}
                      onChange={(e) => updateExpenseRow(row.id, { amount: e.target.value })}
                      className="form-input"
                    />
                    <div className="w-28">
                      <CurrencySelect
                        value={row.currency}
                        onChange={(currency) => updateExpenseRow(row.id, { currency })}
                      />
                    </div>
                  </div>
                </div>

                <div>
                  <label className="form-label">{t('common.description')}</label>
                  <input
                    type="text"
                    value={row.description}
                    onChange={(e) => updateExpenseRow(row.id, { description: e.target.value })}
                    className="form-input"
                  />
                </div>

                <div>
                  <label className="form-label">{t('expectedExpenses.category')}</label>
                  <CategoryCombobox
                    groupId={selectedGroupId}
                    value={row.category_id}
                    onChange={(categoryId) => updateExpenseRow(row.id, { category_id: categoryId })}
                    getAccessTokenSilently={getAccessTokenSilently}
                  />
                </div>
              </div>
            ))}

            <button
              type="button"
              onClick={addExpenseRow}
              className="w-full rounded-md border border-dashed border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
            >
              {t('budgetPlan.addExpense')}
            </button>

            <div className="flex justify-between">
              <button
                type="button"
                onClick={() => setStep(1)}
                className="rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600"
              >
                {t('budgetPlan.back')}
              </button>
              <button
                type="button"
                disabled={!step2Valid}
                onClick={() => setStep(3)}
                className="rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
              >
                {t('budgetPlan.next')}
              </button>
            </div>
          </div>
        )}

        {/* Step 3: Review & Create */}
        {step === 3 && (
          <div className="space-y-4">
            <div className="rounded-md bg-gray-50 p-4 space-y-2 dark:bg-gray-900/50">
              <div className="flex justify-between text-sm">
                <span className="font-medium text-gray-700 dark:text-gray-300">{t('common.selectGroup')}</span>
                <span className="text-gray-900 dark:text-white">{groups?.find((g) => g.id === selectedGroupId)?.name ?? '—'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="font-medium text-gray-700 dark:text-gray-300">{t('common.name')}</span>
                <span className="text-gray-900 dark:text-white">{baseName}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="font-medium text-gray-700 dark:text-gray-300">{t('budgetPlan.cadence')}</span>
                <span className="text-gray-900 dark:text-white">{t(`budgets.periodTypes.${cadence}`)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="font-medium text-gray-700 dark:text-gray-300">{t('budgetPlan.numberOfPeriods')}</span>
                <span className="text-gray-900 dark:text-white">{numberOfPeriods}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="font-medium text-gray-700 dark:text-gray-300">{t('budgets.startDate')}</span>
                <span className="text-gray-900 dark:text-white">{startDate}</span>
              </div>
            </div>

            {/* Periods list */}
            <div>
              <p className="text-sm font-medium text-gray-700 mb-2 dark:text-gray-300">{t('budgetPlan.periodPreview')}</p>
              <div className="max-h-40 overflow-y-auto rounded-md border border-gray-200 dark:border-gray-700">
                <ul className="divide-y divide-gray-100 dark:divide-gray-700">
                  {periods.map((p, i) => (
                    <li key={i} className="px-3 py-2 text-sm text-gray-700 dark:text-gray-300">
                      <span className="font-medium">
                        {numberOfPeriods > 1 ? `${baseName} - ${p.label}` : baseName}
                      </span>
                      <span className="ml-2 text-gray-400 dark:text-gray-500">({p.start_date} → {p.end_date})</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>

            {/* Template expenses summary */}
            <div>
              <p className="text-sm font-medium text-gray-700 mb-2 dark:text-gray-300">{t('budgetPlan.templateExpenses')}</p>
              {validTemplateExpenses.length === 0 ? (
                <p className="text-sm text-gray-500 dark:text-gray-400">{t('budgetPlan.noExpenses')}</p>
              ) : (
                <div className="overflow-x-auto shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg dark:ring-white/10">
                  <table className="min-w-full divide-y divide-gray-300 dark:divide-gray-700">
                    <thead className="bg-gray-50 dark:bg-gray-900/50">
                      <tr>
                        <th className="py-2 pl-3 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-white">{t('common.name')}</th>
                        <th className="px-3 py-2 text-left text-sm font-semibold text-gray-900 dark:text-white">{t('expectedExpenses.amount')}</th>
                        <th className="px-3 py-2 text-left text-sm font-semibold text-gray-900 dark:text-white">{t('expectedExpenses.category')}</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-800">
                      {validTemplateExpenses.map((expense) => {
                        const category = categories.find((c) => c.id === expense.category_id);
                        return (
                          <tr key={expense.id}>
                            <td className="whitespace-nowrap py-2 pl-3 pr-3 text-sm text-gray-900 dark:text-white">{expense.name}</td>
                            <td className="whitespace-nowrap px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
                              {formatCurrency(expense.amount, expense.currency)}
                            </td>
                            <td className="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
                              {category ? (
                                <div className="flex items-center gap-2">
                                  <div className="w-3 h-3 rounded-full" style={{ backgroundColor: category.color }}></div>
                                  <span>{category.name}</span>
                                </div>
                              ) : '—'}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            {createBudgetPlan.isError && (
              <p className="text-sm text-red-600">
                {t('budgetPlan.createError')}
                {getErrorMessage(createBudgetPlan.error) && (
                  <span className="block text-xs mt-1 opacity-75">{getErrorMessage(createBudgetPlan.error)}</span>
                )}
              </p>
            )}

            {progress && isCreating && (
              <p className="text-sm text-gray-600 dark:text-gray-400">
                {t('budgetPlan.creatingProgress', {
                  current: progress.currentIndex,
                  total: progress.totalPeriods,
                  name: progress.budgetName ?? '',
                })}
              </p>
            )}

            <div className="flex justify-between">
              <button
                type="button"
                onClick={() => setStep(2)}
                disabled={isCreating}
                className="rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:opacity-50 dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600"
              >
                {t('budgetPlan.back')}
              </button>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => navigate('/budgets')}
                  disabled={isCreating}
                  className="rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:opacity-50 dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="button"
                  onClick={handleCreate}
                  disabled={isCreating}
                  className="rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
                >
                  {isCreating ? `${t('budgetPlan.create')}...` : t('budgetPlan.create')}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
