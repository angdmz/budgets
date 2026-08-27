import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { usePreferences, SUPPORTED_THEMES } from '../lib/usePreferences';
import { SUPPORTED_LANGUAGES } from '../lib/languages';

export default function Layout() {
  const { user, logout } = useAuth0();
  const location = useLocation();
  const { t } = useTranslation();
  const { theme, updateTheme, updateLanguage, currentLanguage } = usePreferences();

  const navigation = [
    { nameKey: 'nav.dashboard', href: '/dashboard' },
    { nameKey: 'nav.groups', href: '/groups' },
    { nameKey: 'nav.budgets', href: '/budgets' },
    { nameKey: 'nav.categories', href: '/categories' },
    { nameKey: 'nav.expenses', href: '/expenses' },
    { nameKey: 'nav.expected', href: '/expected-expenses' },
  ];

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Navigation */}
      <nav className="bg-white shadow-sm dark:bg-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-14">
            <h1 className="text-xl font-bold text-primary-600">{t('nav.appName')}</h1>
            <div className="flex items-center space-x-3">
              <span className="text-sm text-gray-700 dark:text-gray-300">{user?.name || user?.email}</span>
              <select
                value={theme}
                onChange={(e) => updateTheme(e.target.value as any)}
                className="bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-2 py-1 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {SUPPORTED_THEMES.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              {/* <select
                value={preferences?.display_currency || 'USD'}
                onChange={(e) => updateDisplayCurrency(e.target.value as any)}
                className="bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-2 py-1 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {SUPPORTED_CURRENCIES.map((currency) => (
                  <option key={currency.value} value={currency.value}>
                    {currency.value}
                  </option>
                ))}
              </select> */}
              <select
                value={currentLanguage}
                onChange={(e) => updateLanguage(e.target.value as any)}
                className="bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-2 py-1 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {SUPPORTED_LANGUAGES.map((lang) => (
                  <option key={lang.code} value={lang.code}>
                    {lang.label}
                  </option>
                ))}
              </select>
              <button
                onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}
                className="bg-white text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:bg-gray-700 dark:hover:text-white px-3 py-2 rounded-md text-sm font-medium border border-gray-300 dark:border-gray-600"
              >
                {t('common.logout')}
              </button>
            </div>
          </div>
          <div className="flex space-x-4 overflow-x-auto pb-0">
            {navigation.map((item) => (
              <Link
                key={item.nameKey}
                to={item.href}
                className={`whitespace-nowrap px-1 py-2 border-b-2 text-sm font-medium ${
                  isActive(item.href)
                    ? 'border-primary-500 text-gray-900 dark:text-white'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
                }`}
              >
                {t(item.nameKey)}
              </Link>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
}
