export interface Money {
  amount: string;
  currency: string;
}

export interface Group {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface Budget {
  id: string;
  name: string;
  description: string;
  start_date: string;
  end_date: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  name: string;
  description: string;
  color: string;
  icon: string;
  created_at: string;
  updated_at: string;
}

export interface ExpectedExpense {
  id: string;
  name: string;
  description: string;
  amount: Money;
  category_id: string;
  budget_id: string;
  created_at: string;
  updated_at: string;
}

export interface ActualExpense {
  id: string;
  name: string;
  description: string;
  amount: Money;
  expense_date: string;
  category_id: string;
  budget_id: string;
  expected_expense_id?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateGroupRequest {
  name: string;
  description: string;
}

export interface CreateBudgetRequest {
  name: string;
  description: string;
  start_date: string;
  end_date: string;
}

export interface CreateCategoryRequest {
  name: string;
  description: string;
  color: string;
}

export interface CreateActualExpenseRequest {
  name: string;
  description: string;
  amount: Money;
  expense_date: string;
  expected_expense_id?: string;
}

export interface CreateExpectedExpenseRequest {
  name: string;
  description: string;
  amount: Money;
  category_id: string;
}

export interface Invitation {
  id: string;
  token: string;
  group_id?: string;
  group_name: string;
  inviter_name: string;
  status: string;
  role: string;
  expires_at: string;
  accepted_at?: string;
  created_at: string;
}

export interface InvitationDetail {
  group_name: string;
  inviter_name: string;
  status: string;
  role: string;
  expires_at: string;
}

export interface CreateInvitationRequest {
  role?: string;
}

export interface UpdateCategoryRequest {
  name: string;
  description: string;
  color: string;
}

export interface UpdateBudgetRequest {
  name: string;
  description: string;
  start_date: string;
  end_date: string;
}

export interface UpdateActualExpenseRequest {
  name: string;
  description: string;
  amount: Money;
  expense_date: string;
}
