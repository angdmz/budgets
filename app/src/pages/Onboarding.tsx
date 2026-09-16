import { useState, useCallback, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { createApiClient, type GetAccessTokenSilently } from '../lib/api';
import { useOnboarding, ONBOARDING_STEPS } from '../lib/useOnboarding';
import { computeMonthDates } from '../lib/budgetPeriod';
import { formatCurrency } from '../lib/format';
import CategoryCombobox from '../components/CategoryCombobox';
import type {
  Group,
  Budget,
  ExpectedExpense,
  ActualExpense,
  CreateGroupRequest,
  CreateBudgetRequest,
  CreateExpectedExpenseRequest,
  CreateActualExpenseRequest,
  OnboardingStep,
  OnboardingStepResponse,
} from '../lib/types';

interface DraftExpense {
  name: string;
  amount: string;
  currency: string;
  categoryId: string;
}

const DRAFT_KEY = 'onboarding.draftExpenses';

function loadDraft(): DraftExpense[] {
  try {
    const raw = localStorage.getItem(DRAFT_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveDraft(expenses: DraftExpense[]) {
  localStorage.setItem(DRAFT_KEY, JSON.stringify(expenses));
}

function clearDraft() {
  localStorage.removeItem(DRAFT_KEY);
}

export default function Onboarding() {
  const { getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { onboarding, isLoading, completeStep, skipStep, goBack, steps: onboardingSteps, isPending } = useOnboarding();

  const serverStep: OnboardingStep = onboarding?.current_step ?? 'welcome';

  // Local view step — allows instant back/forward navigation between already-visited steps
  const [viewStep, setViewStep] = useState<OnboardingStep>(serverStep);
  useEffect(() => { setViewStep(serverStep); }, [serverStep]);

  const viewStepIndex = ONBOARDING_STEPS.indexOf(viewStep);

  // Steps that are completed or skipped are navigable
  const completedStepTypes = useMemo(() => {
    const set = new Set<OnboardingStep>();
    for (const s of onboardingSteps) {
      if (s.status === 'completed' || s.status === 'skipped') {
        set.add(s.step);
      }
    }
    return set;
  }, [onboardingSteps]);

  // --- Shared state across steps ---
  const [groupId, setGroupId] = useState<string>('');
  const [budgetId, setBudgetId] = useState<string>('');
  const [cadence, setCadence] = useState<string>('monthly');
  const [draftExpenses, setDraftExpenses] = useState<DraftExpense[]>(loadDraft);
  const [actualExpenseId, setActualExpenseId] = useState<string>('');

  // --- Hydrate state from completed step data (survives page reload) ---
  useEffect(() => {
    if (!onboardingSteps.length) return;
    const findStep = (step: OnboardingStep): OnboardingStepResponse | undefined =>
      onboardingSteps.find(s => s.step === step);

    const groupStep = findStep('choose_group');
    if (groupStep?.data?.group_id && !groupId) {
      setGroupId(groupStep.data.group_id as string);
    }

    const cadenceStep = findStep('choose_cadence');
    if (cadenceStep?.data?.budget_id && !budgetId) {
      setBudgetId(cadenceStep.data.budget_id as string);
    }
    if (cadenceStep?.data?.cadence) {
      setCadence(cadenceStep.data.cadence as string);
    }

    const actualStep = findStep('register_actual_expense');
    if (actualStep?.data?.actual_expense_id && !actualExpenseId) {
      setActualExpenseId(actualStep.data.actual_expense_id as string);
    }
  }, [onboardingSteps, groupId, budgetId, actualExpenseId]);

  // --- Fetch groups ---
  const { data: groups = [] } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Group[]>('/groups');
      return response.data;
    },
  });

  // --- Fetch expected expenses for the budget (for summary & comparison) ---
  const { data: expectedExpenses = [] } = useQuery({
    queryKey: ['expected-expenses', budgetId],
    queryFn: async () => {
      if (!budgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ExpectedExpense[]>(`/budgets/${budgetId}/expected-expenses`);
      return response.data;
    },
    enabled: !!budgetId,
  });

  // --- Fetch actual expenses for the budget (for comparison) ---
  const { data: actualExpenses = [] } = useQuery({
    queryKey: ['actual-expenses', budgetId],
    queryFn: async () => {
      if (!budgetId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<ActualExpense[]>(`/budgets/${budgetId}/actual-expenses`);
      return response.data;
    },
    enabled: !!budgetId,
  });

  // --- Create group mutation ---
  const createGroupMutation = useMutation({
    mutationFn: async (req: CreateGroupRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<Group>('/groups', req);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setGroupId(data.id);
    },
  });

  // --- Create budget mutation ---
  const createBudgetMutation = useMutation({
    mutationFn: async ({ groupId, req }: { groupId: string; req: CreateBudgetRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<Budget>(`/groups/${groupId}/budgets`, req);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['budgets', groupId] });
      setBudgetId(data.id);
    },
  });

  // --- Create expected expense mutation ---
  const createExpectedExpenseMutation = useMutation({
    mutationFn: async ({ budgetId, req }: { budgetId: string; req: CreateExpectedExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<ExpectedExpense>(`/budgets/${budgetId}/expected-expenses`, req);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expected-expenses', budgetId] });
    },
  });

  // --- Create actual expense mutation ---
  const createActualExpenseMutation = useMutation({
    mutationFn: async ({ budgetId, req }: { budgetId: string; req: CreateActualExpenseRequest }) => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.post<ActualExpense>(`/budgets/${budgetId}/actual-expenses`, req);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['actual-expenses', budgetId] });
      setActualExpenseId(data.id);
    },
  });

  // --- Step handlers ---
  const handleCompleteWelcome = useCallback(async () => {
    await completeStep({ step: 'welcome', data: {} });
  }, [completeStep]);

  const handleCompleteChooseGroup = useCallback(async () => {
    if (!groupId) return;
    await completeStep({ step: 'choose_group', data: { group_id: groupId } });
  }, [completeStep, groupId]);

  const handleCompleteChooseCadence = useCallback(async () => {
    if (!groupId) return;
    // Guard: if budget already exists (step already completed), don't create a duplicate
    if (budgetId) {
      await completeStep({
        step: 'choose_cadence',
        data: { budget_id: budgetId, cadence },
      });
      return;
    }
    const now = new Date();
    const { start_date: startDate, end_date: endDate } = computeMonthDates(now.getFullYear(), now.getMonth() + 1);
    const budget = await createBudgetMutation.mutateAsync({
      groupId,
      req: {
        name: `${t('budgets.periodTypes.monthly')} ${now.toLocaleDateString()}`,
        description: '',
        start_date: startDate,
        end_date: endDate,
      },
    });
    setBudgetId(budget.id);
    await completeStep({
      step: 'choose_cadence',
      data: { budget_id: budget.id, cadence },
    });
  }, [completeStep, groupId, budgetId, cadence, createBudgetMutation, t]);

  const handleCompleteAddExpectedExpenses = useCallback(async () => {
    if (!budgetId) return;
    // Guard: if step already completed and no new drafts, use existing expense IDs
    if (draftExpenses.length === 0) {
      const existingIds = expectedExpenses.map(e => e.id);
      if (existingIds.length === 0) return;
      await completeStep({
        step: 'add_expected_expenses',
        data: { expected_expense_ids: existingIds },
      });
      return;
    }
    const createdIds: string[] = [];
    for (const draft of draftExpenses) {
      if (!draft.categoryId) continue;
      const expense = await createExpectedExpenseMutation.mutateAsync({
        budgetId,
        req: {
          name: draft.name,
          description: '',
          amount: { amount: draft.amount, currency: draft.currency },
          category_id: draft.categoryId,
        },
      });
      createdIds.push(expense.id);
    }
    clearDraft();
    await completeStep({
      step: 'add_expected_expenses',
      data: { expected_expense_ids: createdIds },
    });
  }, [completeStep, budgetId, draftExpenses, expectedExpenses, createExpectedExpenseMutation]);

  const handleCompleteBudgetSummary = useCallback(async () => {
    const total = expectedExpenses.reduce((sum, e) => sum + parseFloat(e.amount.amount), 0);
    const currency = expectedExpenses[0]?.amount.currency ?? 'USD';
    await completeStep({
      step: 'budget_summary',
      data: { total_expected: total.toFixed(2), currency },
    });
  }, [completeStep, expectedExpenses]);

  const handleCompleteRegisterActualExpense = useCallback(async () => {
    if (!budgetId || !actualExpenseId) return;
    await completeStep({
      step: 'register_actual_expense',
      data: { actual_expense_id: actualExpenseId },
    });
  }, [completeStep, budgetId, actualExpenseId]);

  const handleCompleteCompareExpenses = useCallback(async () => {
    await completeStep({ step: 'compare_expenses', data: {} });
  }, [completeStep]);

  const handleCompleteDashboardTour = useCallback(async () => {
    await completeStep({ step: 'dashboard_tour', data: {} });
  }, [completeStep]);

  const handleComplete = useCallback(async () => {
    await completeStep({ step: 'complete', data: {} });
    navigate('/dashboard');
  }, [completeStep, navigate]);

  const handleSkip = useCallback(async () => {
    await skipStep(viewStep);
    // Navigate to dashboard if we're at the last navigable step
    if (viewStep === 'dashboard_tour') {
      navigate('/dashboard');
    }
  }, [skipStep, viewStep, navigate]);

  const handleBack = useCallback(() => {
    const prevStep = ONBOARDING_STEPS[viewStepIndex - 1];
    if (prevStep && completedStepTypes.has(prevStep)) {
      setViewStep(prevStep);
      goBack(viewStep); // persist in background
    }
  }, [viewStepIndex, viewStep, completedStepTypes, goBack]);

  // --- Loading ---
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (onboarding?.status === 'completed' || onboarding?.status === 'skipped') {
    navigate('/dashboard');
    return null;
  }

  // --- Step rendering ---
  const progress = ((viewStepIndex + 1) / ONBOARDING_STEPS.length) * 100;
  const canGoBack = viewStepIndex > 0 && completedStepTypes.has(ONBOARDING_STEPS[viewStepIndex - 1]);

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      {/* Progress bar */}
      <div className="mb-6 sm:mb-8">
        <div className="flex items-center justify-between mb-2">
          <h1 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white">
            {t('onboarding.title')}
          </h1>
          <span className="text-sm text-gray-500">
            {viewStepIndex + 1} / {ONBOARDING_STEPS.length}
          </span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2 dark:bg-gray-700">
          <div
            className="bg-primary-600 h-2 rounded-full transition-all duration-300"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Step content */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 sm:p-6">
        {viewStep === 'welcome' && (
          <WelcomeStep onNext={handleCompleteWelcome} onSkip={handleSkip} onBack={canGoBack ? handleBack : undefined} isPending={isPending} />
        )}
        {viewStep === 'choose_group' && (
          <ChooseGroupStep
            groups={groups}
            groupId={groupId}
            setGroupId={setGroupId}
            onCreateGroup={(name) => createGroupMutation.mutate({ name, description: '' })}
            onNext={handleCompleteChooseGroup}
            onSkip={handleSkip}
            onBack={canGoBack ? handleBack : undefined}
            isPending={isPending}
          />
        )}
        {viewStep === 'choose_cadence' && (
          <ChooseCadenceStep
            cadence={cadence}
            setCadence={setCadence}
            onNext={handleCompleteChooseCadence}
            onSkip={handleSkip}
            onBack={canGoBack ? handleBack : undefined}
            isPending={isPending || createBudgetMutation.isPending}
          />
        )}
        {viewStep === 'add_expected_expenses' && (
          <AddExpectedExpensesStep
            groupId={groupId}
            getAccessTokenSilently={getAccessTokenSilently}
            draftExpenses={draftExpenses}
            setDraftExpenses={(exp) => { setDraftExpenses(exp); saveDraft(exp); }}
            onNext={handleCompleteAddExpectedExpenses}
            onSkip={handleSkip}
            onBack={canGoBack ? handleBack : undefined}
            isPending={isPending || createExpectedExpenseMutation.isPending}
          />
        )}
        {viewStep === 'budget_summary' && (
          <BudgetSummaryStep
            expectedExpenses={expectedExpenses}
            onNext={handleCompleteBudgetSummary}
            onSkip={handleSkip}
            onBack={canGoBack ? handleBack : undefined}
            isPending={isPending}
          />
        )}
        {viewStep === 'register_actual_expense' && (
          <RegisterActualExpenseStep
            groupId={groupId}
            getAccessTokenSilently={getAccessTokenSilently}
            actualExpenseId={actualExpenseId}
            onCreate={(req) => createActualExpenseMutation.mutate({ budgetId, req })}
            onNext={handleCompleteRegisterActualExpense}
            onSkip={handleSkip}
            onBack={canGoBack ? handleBack : undefined}
            isPending={isPending || createActualExpenseMutation.isPending}
          />
        )}
        {viewStep === 'compare_expenses' && (
          <CompareExpensesStep
            expectedExpenses={expectedExpenses}
            actualExpenses={actualExpenses}
            onNext={handleCompleteCompareExpenses}
            onSkip={handleSkip}
            onBack={canGoBack ? handleBack : undefined}
            isPending={isPending}
          />
        )}
        {viewStep === 'dashboard_tour' && (
          <DashboardTourStep onNext={handleCompleteDashboardTour} onSkip={handleSkip} onBack={canGoBack ? handleBack : undefined} isPending={isPending} />
        )}
        {viewStep === 'complete' && (
          <CompleteStep onNext={handleComplete} isPending={isPending} />
        )}
      </div>
    </div>
  );
}

// --- Step components ---

function StepShell({
  title,
  description,
  children,
  onNext,
  onSkip,
  onBack,
  nextLabel,
  isPending,
}: {
  title: string;
  description: string;
  children: React.ReactNode;
  onNext: () => void;
  onSkip?: () => void;
  onBack?: () => void;
  nextLabel: string;
  isPending: boolean;
}) {
  const { t } = useTranslation();
  return (
    <div>
      <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">{title}</h2>
      <p className="text-gray-600 dark:text-gray-400 mb-6">{description}</p>
      {children}
      <div className="mt-8 flex items-center justify-between">
        <div>
          {onBack && (
            <button
              onClick={onBack}
              disabled={isPending}
              className="px-4 py-2 text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50"
            >
              ← {t('onboarding.back')}
            </button>
          )}
        </div>
        <div className="flex items-center gap-4">
          {onSkip && (
            <button
              onClick={onSkip}
              disabled={isPending}
              className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            >
              {t('onboarding.skip')}
            </button>
          )}
          <button
            onClick={onNext}
            disabled={isPending}
            className="px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50"
          >
            {isPending ? t('common.loading') : nextLabel}
          </button>
        </div>
      </div>
    </div>
  );
}

function WelcomeStep({ onNext, onSkip, onBack, isPending }: { onNext: () => void; onSkip: () => void; onBack?: () => void; isPending: boolean }) {
  const { t } = useTranslation();
  return (
    <StepShell
      title={t('onboarding.welcome.title')}
      description={t('onboarding.welcome.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.start')}
      isPending={isPending}
    >
      <div className="space-y-3">
        <p className="text-gray-700 dark:text-gray-300">{t('onboarding.welcome.bullet1')}</p>
        <p className="text-gray-700 dark:text-gray-300">{t('onboarding.welcome.bullet2')}</p>
        <p className="text-gray-700 dark:text-gray-300">{t('onboarding.welcome.bullet3')}</p>
      </div>
    </StepShell>
  );
}

function ChooseGroupStep({
  groups,
  groupId,
  setGroupId,
  onCreateGroup,
  onNext,
  onSkip,
  onBack,
  isPending,
}: {
  groups: Group[];
  groupId: string;
  setGroupId: (id: string) => void;
  onCreateGroup: (name: string) => void;
  onNext: () => void;
  onSkip: () => void;
  onBack?: () => void;
  isPending: boolean;
}) {
  const { t } = useTranslation();
  const [newGroupName, setNewGroupName] = useState('');
  return (
    <StepShell
      title={t('onboarding.chooseGroup.title')}
      description={t('onboarding.chooseGroup.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.next')}
      isPending={isPending}
    >
      {groups.length > 0 && (
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            {t('onboarding.chooseGroup.existing')}
          </label>
          <select
            value={groupId}
            onChange={(e) => setGroupId(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          >
            <option value="">{t('common.selectGroupPlaceholder')}</option>
            {groups.map((g) => (
              <option key={g.id} value={g.id}>{g.name}</option>
            ))}
          </select>
        </div>
      )}
      <div>
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          {t('onboarding.chooseGroup.createNew')}
        </label>
        <div className="flex gap-2">
          <input
            type="text"
            value={newGroupName}
            onChange={(e) => setNewGroupName(e.target.value)}
            placeholder={t('onboarding.chooseGroup.namePlaceholder')}
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          />
          <button
            onClick={() => { if (newGroupName.trim()) { onCreateGroup(newGroupName.trim()); setNewGroupName(''); } }}
            className="px-4 py-2 bg-gray-200 dark:bg-gray-600 text-gray-800 dark:text-white rounded-md hover:bg-gray-300 dark:hover:bg-gray-500"
          >
            {t('common.create')}
          </button>
        </div>
      </div>
    </StepShell>
  );
}

function ChooseCadenceStep({
  cadence,
  setCadence,
  onNext,
  onSkip,
  onBack,
  isPending,
}: {
  cadence: string;
  setCadence: (c: string) => void;
  onNext: () => void;
  onSkip: () => void;
  onBack?: () => void;
  isPending: boolean;
}) {
  const { t } = useTranslation();
  const options = ['weekly', 'biweekly', 'monthly', 'custom'];
  return (
    <StepShell
      title={t('onboarding.chooseCadence.title')}
      description={t('onboarding.chooseCadence.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.next')}
      isPending={isPending}
    >
      <div className="grid grid-cols-2 gap-3">
        {options.map((opt) => (
          <button
            key={opt}
            onClick={() => setCadence(opt)}
            className={`px-4 py-3 rounded-md border-2 text-center transition-colors ${
              cadence === opt
                ? 'border-primary-600 bg-primary-50 dark:bg-primary-900 text-primary-700 dark:text-primary-300'
                : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
            }`}
          >
            {t(`budgets.periodTypes.${opt}`)}
          </button>
        ))}
      </div>
    </StepShell>
  );
}

function AddExpectedExpensesStep({
  groupId,
  getAccessTokenSilently,
  draftExpenses,
  setDraftExpenses,
  onNext,
  onSkip,
  onBack,
  isPending,
}: {
  groupId: string;
  getAccessTokenSilently: GetAccessTokenSilently;
  draftExpenses: DraftExpense[];
  setDraftExpenses: (e: DraftExpense[]) => void;
  onNext: () => void;
  onSkip: () => void;
  onBack?: () => void;
  isPending: boolean;
}) {
  const { t } = useTranslation();

  const addExpense = () => {
    setDraftExpenses([...draftExpenses, { name: '', amount: '', currency: 'USD', categoryId: '' }]);
  };
  const removeExpense = (idx: number) => {
    setDraftExpenses(draftExpenses.filter((_, i) => i !== idx));
  };
  const updateExpense = (idx: number, field: keyof DraftExpense, value: string) => {
    setDraftExpenses(draftExpenses.map((e, i) => (i === idx ? { ...e, [field]: value } : e)));
  };

  return (
    <StepShell
      title={t('onboarding.addExpectedExpenses.title')}
      description={t('onboarding.addExpectedExpenses.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.next')}
      isPending={isPending}
    >
      {draftExpenses.map((expense, idx) => (
        <div key={idx} className="mb-3 p-3 border border-gray-200 dark:border-gray-600 rounded-md">
          <div className="flex flex-col gap-2 mb-2 sm:flex-row">
            <input
              type="text"
              value={expense.name}
              onChange={(e) => updateExpense(idx, 'name', e.target.value)}
              placeholder={t('common.name')}
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
            <button
              onClick={() => removeExpense(idx)}
              className="px-3 py-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900 rounded-md self-end sm:self-auto"
            >
              ✕
            </button>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row">
            <input
              type="number"
              value={expense.amount}
              onChange={(e) => updateExpense(idx, 'amount', e.target.value)}
              placeholder={t('expenses.amount')}
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
            <select
              value={expense.currency}
              onChange={(e) => updateExpense(idx, 'currency', e.target.value)}
              className="w-full sm:w-24 px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            >
              {['USD', 'EUR', 'GBP', 'ARS', 'BRL', 'MXN'].map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
            <div className="flex-1">
              <CategoryCombobox
                groupId={groupId}
                getAccessTokenSilently={getAccessTokenSilently}
                value={expense.categoryId}
                onChange={(id) => updateExpense(idx, 'categoryId', id)}
                allowCreate
                placeholder={t('categories.selectCategory')}
              />
            </div>
          </div>
        </div>
      ))}
      <button
        onClick={addExpense}
        className="w-full py-2 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-md text-gray-500 hover:border-primary-500 hover:text-primary-600"
      >
        + {t('onboarding.addExpectedExpenses.add')}
      </button>
    </StepShell>
  );
}

function BudgetSummaryStep({
  expectedExpenses,
  onNext,
  onSkip,
  onBack,
  isPending,
}: {
  expectedExpenses: ExpectedExpense[];
  onNext: () => void;
  onSkip: () => void;
  onBack?: () => void;
  isPending: boolean;
}) {
  const { t } = useTranslation();
  const total = expectedExpenses.reduce((sum, e) => sum + parseFloat(e.amount.amount), 0);
  const currency = expectedExpenses[0]?.amount.currency ?? 'USD';
  return (
    <StepShell
      title={t('onboarding.budgetSummary.title')}
      description={t('onboarding.budgetSummary.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.next')}
      isPending={isPending}
    >
      <div className="text-center py-8">
        <p className="text-sm text-gray-500 mb-2">{t('onboarding.budgetSummary.totalExpected')}</p>
        <p className="text-4xl font-bold text-primary-600">
          {formatCurrency(total.toFixed(2), currency)}
        </p>
      </div>
      <div className="space-y-2">
        {expectedExpenses.map((e) => (
          <div key={e.id} className="flex justify-between py-2 border-b border-gray-100 dark:border-gray-700">
            <span className="text-gray-700 dark:text-gray-300">{e.name}</span>
            <span className="text-gray-900 dark:text-white font-medium">
              {formatCurrency(e.amount.amount, e.amount.currency)}
            </span>
          </div>
        ))}
      </div>
    </StepShell>
  );
}

function RegisterActualExpenseStep({
  groupId,
  getAccessTokenSilently,
  actualExpenseId,
  onCreate,
  onNext,
  onSkip,
  onBack,
  isPending,
}: {
  groupId: string;
  getAccessTokenSilently: GetAccessTokenSilently;
  actualExpenseId: string;
  onCreate: (req: CreateActualExpenseRequest) => void;
  onNext: () => void;
  onSkip: () => void;
  onBack?: () => void;
  isPending: boolean;
}) {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [amount, setAmount] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const today = new Date().toISOString().split('T')[0];

  return (
    <StepShell
      title={t('onboarding.registerActualExpense.title')}
      description={t('onboarding.registerActualExpense.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.next')}
      isPending={isPending}
    >
      {actualExpenseId ? (
        <div className="p-4 bg-green-50 dark:bg-green-900 rounded-md">
          <p className="text-green-700 dark:text-green-300">{t('onboarding.registerActualExpense.done')}</p>
        </div>
      ) : (
        <div className="space-y-3">
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('common.name')}
            className="w-full px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          />
          <div className="flex gap-2">
            <input
              type="number"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder={t('expenses.amount')}
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
            <div className="flex-1">
              <CategoryCombobox
                groupId={groupId}
                getAccessTokenSilently={getAccessTokenSilently}
                value={categoryId}
                onChange={setCategoryId}
                allowCreate
                placeholder={t('categories.selectCategory')}
              />
            </div>
          </div>
          <div className="flex items-center gap-4">
            <button
              onClick={() => {
                if (name.trim() && amount && categoryId) {
                  onCreate({
                    name: name.trim(),
                    description: '',
                    amount: { amount, currency: 'USD' },
                    expense_date: today,
                    category_id: categoryId,
                  });
                }
              }}
              disabled={!name.trim() || !amount || !categoryId}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-600 text-gray-800 dark:text-white rounded-md hover:bg-gray-300 dark:hover:bg-gray-500 disabled:opacity-50"
            >
              {t('onboarding.registerActualExpense.register')}
            </button>
            <button
              onClick={onSkip}
              disabled={isPending}
              className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            >
              {t('onboarding.registerActualExpense.skipActual')}
            </button>
          </div>
        </div>
      )}
    </StepShell>
  );
}

function CompareExpensesStep({
  expectedExpenses,
  actualExpenses,
  onNext,
  onSkip,
  onBack,
  isPending,
}: {
  expectedExpenses: ExpectedExpense[];
  actualExpenses: ActualExpense[];
  onNext: () => void;
  onSkip: () => void;
  onBack?: () => void;
  isPending: boolean;
}) {
  const { t } = useTranslation();
  const expectedTotal = expectedExpenses.reduce((s, e) => s + parseFloat(e.amount.amount), 0);
  const actualTotal = actualExpenses.reduce((s, e) => s + parseFloat(e.amount.amount), 0);
  const diff = expectedTotal - actualTotal;
  const currency = expectedExpenses[0]?.amount.currency ?? actualExpenses[0]?.amount.currency ?? 'USD';
  const hasActualExpenses = actualExpenses.length > 0;

  return (
    <StepShell
      title={t('onboarding.compareExpenses.title')}
      description={t('onboarding.compareExpenses.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.next')}
      isPending={isPending}
    >
      {hasActualExpenses ? (
        <>
          <div className="grid grid-cols-1 gap-4 mb-6 sm:grid-cols-3">
            <div className="text-center p-4 bg-blue-50 dark:bg-blue-900 rounded-md">
              <p className="text-sm text-gray-500">{t('dashboard.expected')}</p>
              <p className="text-xl font-bold text-blue-600 dark:text-blue-400">
                {formatCurrency(expectedTotal.toFixed(2), currency)}
              </p>
            </div>
            <div className="text-center p-4 bg-amber-50 dark:bg-amber-900 rounded-md">
              <p className="text-sm text-gray-500">{t('dashboard.actual')}</p>
              <p className="text-xl font-bold text-amber-600 dark:text-amber-400">
                {formatCurrency(actualTotal.toFixed(2), currency)}
              </p>
            </div>
            <div className={`text-center p-4 rounded-md ${diff >= 0 ? 'bg-green-50 dark:bg-green-900' : 'bg-red-50 dark:bg-red-900'}`}>
              <p className="text-sm text-gray-500">{t('dashboard.difference')}</p>
              <p className={`text-xl font-bold ${diff >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                {formatCurrency(Math.abs(diff).toFixed(2), currency)}
              </p>
              <p className="text-xs text-gray-500">{diff >= 0 ? t('dashboard.under') : t('dashboard.over')}</p>
            </div>
          </div>
          <p className="text-gray-600 dark:text-gray-400 text-sm">{t('onboarding.compareExpenses.explanation')}</p>
        </>
      ) : (
        <div className="text-center py-8">
          <p className="text-gray-500 dark:text-gray-400">{t('onboarding.compareExpenses.noActual')}</p>
        </div>
      )}
    </StepShell>
  );
}

function DashboardTourStep({ onNext, onSkip, onBack, isPending }: { onNext: () => void; onSkip: () => void; onBack?: () => void; isPending: boolean }) {
  const { t } = useTranslation();
  return (
    <StepShell
      title={t('onboarding.dashboardTour.title')}
      description={t('onboarding.dashboardTour.description')}
      onNext={onNext}
      onSkip={onSkip}
      onBack={onBack}
      nextLabel={t('onboarding.next')}
      isPending={isPending}
    >
      <div className="space-y-3">
        <p className="text-gray-700 dark:text-gray-300">{t('onboarding.dashboardTour.tip1')}</p>
        <p className="text-gray-700 dark:text-gray-300">{t('onboarding.dashboardTour.tip2')}</p>
        <p className="text-gray-700 dark:text-gray-300">{t('onboarding.dashboardTour.tip3')}</p>
      </div>
    </StepShell>
  );
}

function CompleteStep({ onNext, isPending }: { onNext: () => void; isPending: boolean }) {
  const { t } = useTranslation();
  return (
    <div className="text-center py-8">
      <div className="text-5xl mb-4">🎉</div>
      <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
        {t('onboarding.complete.title')}
      </h2>
      <p className="text-gray-600 dark:text-gray-400 mb-6">
        {t('onboarding.complete.description')}
      </p>
      <button
        onClick={onNext}
        disabled={isPending}
        className="px-8 py-3 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50"
      >
        {isPending ? t('common.loading') : t('onboarding.complete.goToDashboard')}
      </button>
    </div>
  );
}
