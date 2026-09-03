import type { AxiosInstance } from 'axios';
import type { Budget, ExpectedExpense } from '../types';
import {
  advancePeriod,
  computeMonthDates,
  monthLabel,
  generatePeriods,
  type BudgetPeriodType,
  type MonthOption,
} from '../budgetPeriod';

export interface DuplicateBudgetOptions {
  sourceBudgetId: string;
  groupId: string;
  name: string;
  description: string;
  startDate: string;
  endDate: string;
  copyExpectedExpenses: boolean;
}

export interface DuplicateBudgetResult {
  budget: Budget;
  copiedExpensesCount: number;
}

export interface CreateRecurringOptions {
  groupId: string;
  baseName: string;
  description: string;
  periodType: Exclude<BudgetPeriodType, 'custom'>;
  firstStartDate: string;
  firstEndDate: string;
  numberOfPeriods: number;
  sourceBudgetId?: string;
  copyExpectedExpenses: boolean;
}

export interface CreateRecurringProgress {
  currentPeriod: number;
  totalPeriods: number;
  stage: 'creating-budget' | 'copying-expenses';
}

export interface CreateBudgetsForMonthsOptions {
  groupId: string;
  baseName: string;
  description: string;
  months: MonthOption[];
  sourceBudgetId?: string;
  copyExpectedExpenses: boolean;
  locale?: string;
}

export interface CreateBudgetsForMonthsProgress {
  currentIndex: number;
  totalMonths: number;
  stage: 'creating-budget' | 'copying-expenses';
}

export interface BudgetPlanTemplateExpense {
  name: string;
  description: string;
  amount: string;
  currency: string;
  category_id: string;
}

export interface CreateBudgetPlanOptions {
  groupId: string;
  baseName: string;
  description: string;
  cadence: Exclude<BudgetPeriodType, 'custom'>;
  startDate: string;
  numberOfPeriods: number;
  templateExpenses: BudgetPlanTemplateExpense[];
  locale?: string;
}

export interface CreateBudgetPlanProgress {
  currentIndex: number;
  totalPeriods: number;
  stage: 'creating-budget' | 'adding-expenses';
  budgetName?: string;
}

export interface BudgetService {
  duplicateBudget(options: DuplicateBudgetOptions): Promise<DuplicateBudgetResult>;
  createRecurringBudgets(
    options: CreateRecurringOptions,
    onProgress?: (progress: CreateRecurringProgress) => void
  ): Promise<Budget[]>;
  createBudgetsForMonths(
    options: CreateBudgetsForMonthsOptions,
    onProgress?: (progress: CreateBudgetsForMonthsProgress) => void
  ): Promise<Budget[]>;
  createBudgetPlan(
    options: CreateBudgetPlanOptions,
    onProgress?: (progress: CreateBudgetPlanProgress) => void
  ): Promise<Budget[]>;
}

/**
 * Frontend-orchestrated implementation. All duplicate/recurring logic lives
 * here so that if a backend endpoint (e.g. POST /budgets/{id}/duplicate) is
 * added later, only this file needs to change — callers (hooks/components)
 * are unaffected since they depend on the BudgetService interface only.
 */
export function createBudgetService(api: AxiosInstance): BudgetService {
  async function fetchExpectedExpenses(budgetId: string): Promise<ExpectedExpense[]> {
    const response = await api.get<ExpectedExpense[]>(`/budgets/${budgetId}/expected-expenses`);
    return response.data;
  }

  async function copyExpensesToBudget(sourceBudgetId: string, targetBudgetId: string): Promise<number> {
    const expenses = await fetchExpectedExpenses(sourceBudgetId);
    await Promise.all(
      expenses.map((expense) =>
        api.post(`/budgets/${targetBudgetId}/expected-expenses`, {
          name: expense.name,
          description: expense.description,
          amount: expense.amount,
          category_id: expense.category_id,
        })
      )
    );
    return expenses.length;
  }

  return {
    async duplicateBudget(options: DuplicateBudgetOptions): Promise<DuplicateBudgetResult> {
      const { data: budget } = await api.post<Budget>(`/groups/${options.groupId}/budgets`, {
        name: options.name,
        description: options.description,
        start_date: options.startDate,
        end_date: options.endDate,
      });

      let copiedExpensesCount = 0;
      if (options.copyExpectedExpenses) {
        copiedExpensesCount = await copyExpensesToBudget(options.sourceBudgetId, budget.id);
      }

      return { budget, copiedExpensesCount };
    },

    async createRecurringBudgets(
      options: CreateRecurringOptions,
      onProgress?: (progress: CreateRecurringProgress) => void
    ): Promise<Budget[]> {
      const budgets: Budget[] = [];
      let periodStart = new Date(`${options.firstStartDate}T00:00:00`);
      let periodEnd = new Date(`${options.firstEndDate}T00:00:00`);

      for (let i = 0; i < options.numberOfPeriods; i++) {
        onProgress?.({ currentPeriod: i + 1, totalPeriods: options.numberOfPeriods, stage: 'creating-budget' });

        const start_date = periodStart.toISOString().slice(0, 10);
        const end_date = periodEnd.toISOString().slice(0, 10);
        const suffix = options.numberOfPeriods > 1 ? ` #${i + 1}` : '';

        const { data: budget } = await api.post<Budget>(`/groups/${options.groupId}/budgets`, {
          name: `${options.baseName}${suffix}`,
          description: options.description,
          start_date,
          end_date,
        });
        budgets.push(budget);

        if (options.copyExpectedExpenses && options.sourceBudgetId) {
          onProgress?.({ currentPeriod: i + 1, totalPeriods: options.numberOfPeriods, stage: 'copying-expenses' });
          await copyExpensesToBudget(options.sourceBudgetId, budget.id);
        }

        const nextStart = advancePeriod(options.periodType, periodStart);
        const nextEnd = advancePeriod(options.periodType, periodEnd);
        periodStart = nextStart;
        periodEnd = nextEnd;
      }

      return budgets;
    },

    async createBudgetsForMonths(
      options: CreateBudgetsForMonthsOptions,
      onProgress?: (progress: CreateBudgetsForMonthsProgress) => void
    ): Promise<Budget[]> {
      const sortedMonths = [...options.months].sort((a, b) => a.year - b.year || a.month - b.month);
      const budgets: Budget[] = [];

      for (let i = 0; i < sortedMonths.length; i++) {
        const { year, month } = sortedMonths[i];
        onProgress?.({ currentIndex: i + 1, totalMonths: sortedMonths.length, stage: 'creating-budget' });

        const { start_date, end_date } = computeMonthDates(year, month);
        const name = `${options.baseName} - ${monthLabel(year, month, options.locale)}`;

        const { data: budget } = await api.post<Budget>(`/groups/${options.groupId}/budgets`, {
          name,
          description: options.description,
          start_date,
          end_date,
        });
        budgets.push(budget);

        if (options.copyExpectedExpenses && options.sourceBudgetId) {
          onProgress?.({ currentIndex: i + 1, totalMonths: sortedMonths.length, stage: 'copying-expenses' });
          await copyExpensesToBudget(options.sourceBudgetId, budget.id);
        }
      }

      return budgets;
    },

    async createBudgetPlan(
      options: CreateBudgetPlanOptions,
      onProgress?: (progress: CreateBudgetPlanProgress) => void
    ): Promise<Budget[]> {
      const periods = generatePeriods(
        options.cadence,
        options.startDate,
        options.numberOfPeriods,
        options.locale
      );
      const budgets: Budget[] = [];

      for (let i = 0; i < periods.length; i++) {
        const period = periods[i];
        const budgetName =
          options.numberOfPeriods > 1
            ? `${options.baseName} - ${period.label}`
            : options.baseName;

        onProgress?.({
          currentIndex: i + 1,
          totalPeriods: periods.length,
          stage: 'creating-budget',
          budgetName,
        });

        const { data: budget } = await api.post<Budget>(
          `/groups/${options.groupId}/budgets`,
          {
            name: budgetName,
            description: options.description,
            start_date: period.start_date,
            end_date: period.end_date,
          }
        );
        budgets.push(budget);

        if (options.templateExpenses.length > 0) {
          onProgress?.({
            currentIndex: i + 1,
            totalPeriods: periods.length,
            stage: 'adding-expenses',
            budgetName,
          });

          for (const expense of options.templateExpenses) {
            await api.post(`/budgets/${budget.id}/expected-expenses`, {
              name: expense.name,
              description: expense.description,
              amount: { amount: expense.amount, currency: expense.currency },
              category_id: expense.category_id,
            });
          }
        }
      }

      return budgets;
    },
  };
}
