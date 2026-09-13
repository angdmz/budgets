import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import Dashboard from './pages/Dashboard';
import Groups from './pages/Groups';
import Budgets from './pages/Budgets';
import CreateBudgetPlan from './pages/CreateBudgetPlan';
import BudgetDetail from './pages/BudgetDetail';
import Categories from './pages/Categories';
import Expenses from './pages/Expenses';
import ExpectedExpenses from './pages/ExpectedExpenses';
import AcceptInvitation from './pages/AcceptInvitation';
import Onboarding from './pages/Onboarding';

function App() {
  const { isLoading } = useAuth0();
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">{t('common.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <Routes>
      <Route element={<ProtectedRoute />}>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="/onboarding" element={<Onboarding />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/groups" element={<Groups />} />
          <Route path="/budgets" element={<Budgets />} />
          <Route path="/budgets/new" element={<CreateBudgetPlan />} />
          <Route path="/budgets/:budgetId" element={<BudgetDetail />} />
          <Route path="/categories" element={<Categories />} />
          <Route path="/expenses" element={<Expenses />} />
          <Route path="/expected-expenses" element={<ExpectedExpenses />} />
        </Route>
      </Route>
      <Route path="/invite/:token" element={<AcceptInvitation />} />
    </Routes>
  );
}

export default App;
