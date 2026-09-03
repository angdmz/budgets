export type BudgetPeriodType = 'weekly' | 'biweekly' | 'monthly' | 'custom';

function toISODate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * Computes the [start_date, end_date] pair (ISO yyyy-mm-dd strings) for a given
 * period type, anchored to a reference date (defaults to today).
 *
 * - weekly:   Monday -> Sunday of the reference date's week
 * - biweekly: Monday of the reference date's week -> Sunday of the following week
 * - monthly:  1st -> last day of the reference date's month
 */
export function computePeriodDates(
  periodType: Exclude<BudgetPeriodType, 'custom'>,
  referenceDate: Date = new Date()
): { start_date: string; end_date: string } {
  const ref = new Date(referenceDate.getFullYear(), referenceDate.getMonth(), referenceDate.getDate());

  if (periodType === 'monthly') {
    const start = new Date(ref.getFullYear(), ref.getMonth(), 1);
    const end = new Date(ref.getFullYear(), ref.getMonth() + 1, 0);
    return { start_date: toISODate(start), end_date: toISODate(end) };
  }

  // weekly / biweekly: find Monday of the reference week
  const dayOfWeek = ref.getDay(); // 0 = Sunday ... 6 = Saturday
  const diffToMonday = dayOfWeek === 0 ? -6 : 1 - dayOfWeek;
  const monday = new Date(ref);
  monday.setDate(ref.getDate() + diffToMonday);

  const sunday = new Date(monday);
  sunday.setDate(monday.getDate() + (periodType === 'biweekly' ? 13 : 6));

  return { start_date: toISODate(monday), end_date: toISODate(sunday) };
}

/**
 * Advances a reference date forward by one period, used to compute subsequent
 * periods when creating recurring budgets.
 */
export function advancePeriod(
  periodType: Exclude<BudgetPeriodType, 'custom'>,
  referenceDate: Date
): Date {
  const next = new Date(referenceDate.getFullYear(), referenceDate.getMonth(), referenceDate.getDate());
  if (periodType === 'monthly') {
    next.setMonth(next.getMonth() + 1);
  } else if (periodType === 'biweekly') {
    next.setDate(next.getDate() + 14);
  } else {
    next.setDate(next.getDate() + 7);
  }
  return next;
}

/**
 * A single calendar month, used to let users pick specific months (e.g. "the
 * rest of the year") when creating monthly budgets in bulk.
 */
export interface MonthOption {
  year: number;
  month: number; // 1-12
}

/**
 * Computes the [start_date, end_date] pair (ISO yyyy-mm-dd strings) spanning
 * the full given calendar month (1st -> last day).
 */
export function computeMonthDates(year: number, month: number): { start_date: string; end_date: string } {
  const start = new Date(year, month - 1, 1);
  const end = new Date(year, month, 0);
  return { start_date: toISODate(start), end_date: toISODate(end) };
}

/**
 * Human-readable "Month Year" label (e.g. "January 2026"), used to name
 * budgets generated from a month picker.
 */
export function monthLabel(year: number, month: number, locale?: string): string {
  const date = new Date(year, month - 1, 1);
  return date.toLocaleDateString(locale ?? 'en-US', { month: 'long', year: 'numeric' });
}

/**
 * Returns every month strictly after the reference date's month through the
 * end of that same calendar year (e.g. reference = March 2026 -> Apr..Dec 2026).
 */
export function remainingMonthsOfYear(referenceDate: Date = new Date()): MonthOption[] {
  const year = referenceDate.getFullYear();
  const months: MonthOption[] = [];
  for (let month = referenceDate.getMonth() + 2; month <= 12; month++) {
    months.push({ year, month });
  }
  return months;
}

/**
 * Returns the three months belonging to the given quarter (1-4) of the given year.
 */
export function quarterMonths(quarter: 1 | 2 | 3 | 4, year: number): MonthOption[] {
  const startMonth = (quarter - 1) * 3 + 1;
  return [startMonth, startMonth + 1, startMonth + 2].map((month) => ({ year, month }));
}

export interface GeneratedPeriod {
  start_date: string;
  end_date: string;
  label: string;
}

/**
 * Generates `count` consecutive periods starting from `startDateISO`, advancing
 * by the given cadence each iteration. Each period gets a human-readable label
 * (e.g. "January 2026" for monthly, "#1" for weekly/biweekly).
 */
export function generatePeriods(
  cadence: Exclude<BudgetPeriodType, 'custom'>,
  startDateISO: string,
  count: number,
  locale?: string
): GeneratedPeriod[] {
  const periods: GeneratedPeriod[] = [];
  let ref = new Date(`${startDateISO}T00:00:00`);

  for (let i = 0; i < count; i++) {
    const { start_date, end_date } = computePeriodDates(cadence, ref);

    let label: string;
    if (cadence === 'monthly') {
      label = ref.toLocaleDateString(locale ?? 'en-US', { month: 'long', year: 'numeric' });
    } else {
      label = `#${i + 1}`;
    }

    periods.push({ start_date, end_date, label });
    ref = advancePeriod(cadence, ref);
  }

  return periods;
}
