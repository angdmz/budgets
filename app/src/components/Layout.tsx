import { useState } from 'react';
import { Outlet, Link, useLocation, Navigate } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { usePreferences, SUPPORTED_CURRENCIES, SUPPORTED_QUOTE_TYPES } from '../lib/usePreferences';
import { useOnboarding } from '../lib/useOnboarding';
import { SUPPORTED_LANGUAGES } from '../lib/languages';
import Dialog from './Dialog';

export default function Layout() {
  const { user, logout } = useAuth0();
  const location = useLocation();
  const { t } = useTranslation();
  const { theme, updateTheme, updateLanguage, currentLanguage, preferences, updateDisplayCurrency, updatePreferredQuoteType } = usePreferences();
  const { onboarding, isLoading: isOnboardingLoading } = useOnboarding();
  const [moreOpen, setMoreOpen] = useState(false);

  const isOnOnboardingPage = location.pathname === '/onboarding';

  // Show loading spinner while onboarding status is being fetched, so the nav
  // doesn't appear before the onboarding redirect check is complete.
  if (isOnboardingLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
        </div>
      </div>
    );
  }

  // Redirect to onboarding if onboarding is in progress and user is not on the onboarding page
  if (onboarding && onboarding.status === 'in_progress' && !isOnOnboardingPage) {
    return <Navigate to="/onboarding" replace />;
  }

  const navigation = [
    { nameKey: 'nav.dashboard', href: '/dashboard' },
    { nameKey: 'nav.groups', href: '/groups' },
    { nameKey: 'nav.budgets', href: '/budgets' },
    { nameKey: 'nav.categories', href: '/categories' },
    { nameKey: 'nav.expenses', href: '/expenses' },
    { nameKey: 'nav.expected', href: '/expected-expenses' },
  ];

  // Mobile bottom nav: 5 items with Add in the center
  const bottomNav = [
    { nameKey: 'nav.dashboard', href: '/dashboard' },
    { nameKey: 'nav.budgets', href: '/budgets' },
    { nameKey: 'nav.add', href: '/expenses?new=1', isAdd: true },
    { nameKey: 'nav.expenses', href: '/expenses' },
    { nameKey: 'nav.more', href: '', isMore: true },
  ];

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Desktop Navigation */}
      <nav aria-label="Main navigation" className="bg-white shadow-sm dark:bg-gray-800 hidden md:block">
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
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
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
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                  </svg>
                </button>
              </div>
              <select
                value={currentLanguage}
                onChange={(e) => updateLanguage(e.target.value as any)}
                aria-label="Language"
                className="bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-2 py-1 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {SUPPORTED_LANGUAGES.map((lang) => (
                  <option key={lang.code} value={lang.code}>
                    {lang.label}
                  </option>
                ))}
              </select>
              <select
                value={preferences?.display_currency ?? 'USD'}
                onChange={(e) => updateDisplayCurrency(e.target.value as any)}
                aria-label="Display currency"
                className="bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-2 py-1 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {SUPPORTED_CURRENCIES.map((c) => (
                  <option key={c.value} value={c.value}>
                    {c.value}
                  </option>
                ))}
              </select>
              <select
                value={preferences?.preferred_quote_type ?? 'OFFICIAL'}
                onChange={(e) => updatePreferredQuoteType(e.target.value as any)}
                aria-label="Quote type"
                className="bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-2 py-1 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                {SUPPORTED_QUOTE_TYPES.map((q) => (
                  <option key={q.value} value={q.value}>
                    {q.label}
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
                aria-current={isActive(item.href) ? 'page' : undefined}
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

      {/* Mobile Header */}
      <header className="md:hidden sticky top-0 z-30 bg-white dark:bg-gray-800 shadow-sm">
        <div className="flex items-center justify-between h-14 px-4">
          <h1 className="text-lg font-bold text-primary-600">{t('nav.appName')}</h1>
          <button
            onClick={() => updateTheme(theme === 'LIGHT' ? 'DARK' : 'LIGHT')}
            aria-label={theme === 'LIGHT' ? 'Dark theme' : 'Light theme'}
            className="p-2 rounded-md text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
          >
            {theme === 'LIGHT' ? (
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
              </svg>
            ) : (
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <circle cx="12" cy="12" r="4" />
                <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
              </svg>
            )}
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 md:py-6 pb-24 md:pb-6">
        <Outlet />
      </main>

      {/* Mobile Bottom Navigation */}
      <nav aria-label="Mobile navigation" className="md:hidden fixed bottom-0 left-0 right-0 z-30 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 pb-[env(safe-area-inset-bottom)]">
        <div className="flex items-stretch justify-around h-16">
          {bottomNav.map((item) => {
            if (item.isMore) {
              return (
                <button
                  key="more"
                  onClick={() => setMoreOpen(true)}
                  aria-expanded={moreOpen}
                  aria-haspopup="dialog"
                  aria-label={t('nav.more')}
                  className="flex flex-col items-center justify-center flex-1 min-w-0 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                    <circle cx="12" cy="12" r="1" />
                    <circle cx="12" cy="5" r="1" />
                    <circle cx="12" cy="19" r="1" />
                  </svg>
                  <span className="text-xs mt-0.5">{t('nav.more')}</span>
                </button>
              );
            }
            if (item.isAdd) {
              return (
                <Link
                  key="add"
                  to={item.href}
                  className="flex flex-col items-center justify-center flex-1 min-w-0"
                >
                  <div className="flex items-center justify-center w-12 h-12 -mt-4 rounded-full bg-primary-600 text-white shadow-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                      <path d="M12 5v14M5 12h14" />
                    </svg>
                  </div>
                  <span className="text-xs mt-0.5 text-primary-600 font-medium">{t('nav.add')}</span>
                </Link>
              );
            }
            const active = isActive(item.href);
            return (
              <Link
                key={item.nameKey}
                to={item.href}
                aria-current={active ? 'page' : undefined}
                className={`flex flex-col items-center justify-center flex-1 min-w-0 ${
                  active
                    ? 'text-primary-600 dark:text-primary-400'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                }`}
              >
                <span className="text-sm font-medium truncate">{t(item.nameKey)}</span>
              </Link>
            );
          })}
        </div>
      </nav>

      {/* Mobile More Drawer */}
      {moreOpen && (
        <Dialog
          title={t('nav.more')}
          onClose={() => setMoreOpen(false)}
          dismissible
        >
          <div className="space-y-4">
            <div className="space-y-2">
              {navigation.map((item) => (
                <Link
                  key={item.nameKey}
                  to={item.href}
                  onClick={() => setMoreOpen(false)}
                  className={`block px-3 py-3 rounded-md text-sm font-medium min-h-[44px] flex items-center ${
                    isActive(item.href)
                      ? 'bg-primary-50 text-primary-600 dark:bg-primary-900 dark:text-primary-400'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
                  }`}
                >
                  {t(item.nameKey)}
                </Link>
              ))}
            </div>
            <hr className="border-gray-200 dark:border-gray-700" />
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-700 dark:text-gray-300">{user?.name || user?.email}</span>
                <div className="flex items-center space-x-1">
                  <button
                    onClick={() => updateTheme('LIGHT')}
                    aria-label="Light theme"
                    className={`p-2 rounded-md transition-colors ${
                      theme === 'LIGHT'
                        ? 'bg-primary-100 text-primary-600 dark:bg-primary-900 dark:text-primary-400'
                        : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
                    }`}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                      <circle cx="12" cy="12" r="4" />
                      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
                    </svg>
                  </button>
                  <button
                    onClick={() => updateTheme('DARK')}
                    aria-label="Dark theme"
                    className={`p-2 rounded-md transition-colors ${
                      theme === 'DARK'
                        ? 'bg-primary-100 text-primary-600 dark:bg-primary-900 dark:text-primary-400'
                        : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
                    }`}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                    </svg>
                  </button>
                </div>
              </div>
              <select
                value={currentLanguage}
                onChange={(e) => updateLanguage(e.target.value as any)}
                aria-label="Language"
                className="w-full bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-3 py-2 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500 min-h-[44px]"
              >
                {SUPPORTED_LANGUAGES.map((lang) => (
                  <option key={lang.code} value={lang.code}>
                    {lang.label}
                  </option>
                ))}
              </select>
              <select
                value={preferences?.display_currency ?? 'USD'}
                onChange={(e) => updateDisplayCurrency(e.target.value as any)}
                aria-label="Display currency"
                className="w-full bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-3 py-2 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500 min-h-[44px]"
              >
                {SUPPORTED_CURRENCIES.map((c) => (
                  <option key={c.value} value={c.value}>
                    {c.label}
                  </option>
                ))}
              </select>
              <select
                value={preferences?.preferred_quote_type ?? 'OFFICIAL'}
                onChange={(e) => updatePreferredQuoteType(e.target.value as any)}
                aria-label="Quote type"
                className="w-full bg-white text-gray-700 dark:text-gray-300 dark:bg-gray-800 px-3 py-2 rounded-md text-sm border border-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500 min-h-[44px]"
              >
                {SUPPORTED_QUOTE_TYPES.map((q) => (
                  <option key={q.value} value={q.value}>
                    {q.label}
                  </option>
                ))}
              </select>
              <button
                onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}
                className="w-full bg-white text-gray-700 hover:text-gray-900 dark:text-gray-300 dark:bg-gray-700 dark:hover:text-white px-3 py-2 rounded-md text-sm font-medium border border-gray-300 dark:border-gray-600 min-h-[44px]"
              >
                {t('common.logout')}
              </button>
            </div>
          </div>
        </Dialog>
      )}
    </div>
  );
}
