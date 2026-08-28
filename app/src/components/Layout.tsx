import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { usePreferences } from '../lib/usePreferences';
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
              <div className="flex items-center space-x-1">
                <button
                  onClick={() => updateTheme('LIGHT')}
                  aria-label="Light theme"
                  className={`p-1.5 rounded-md transition-colors ${
                    theme === 'LIGHT'
                      ? 'bg-primary-100 text-primary-600 dark:bg-primary-900 dark:text-primary-400'
                      : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
                  }`}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="4" />
                    <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
                  </svg>
                </button>
                <button
                  onClick={() => updateTheme('DARK')}
                  aria-label="Dark theme"
                  className={`p-1.5 rounded-md transition-colors ${
                    theme === 'DARK'
                      ? 'bg-primary-100 text-primary-600 dark:bg-primary-900 dark:text-primary-400'
                      : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
                  }`}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                  </svg>
                </button>
              </div>
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
