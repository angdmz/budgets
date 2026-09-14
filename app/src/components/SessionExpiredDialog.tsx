import { useEffect, useState } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import { useQueryClient } from '@tanstack/react-query';
import { onSessionExpired } from '../lib/session';
import Dialog from './Dialog';

const AUTO_LOGOUT_MS = 5000;

export default function SessionExpiredDialog() {
  const { logout } = useAuth0();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const handleLogout = () =>
    logout({ logoutParams: { returnTo: window.location.origin } });

  useEffect(
    () =>
      onSessionExpired(() => {
        queryClient.cancelQueries();
        queryClient.clear();
        setOpen(true);
      }),
    [queryClient]
  );

  useEffect(() => {
    if (!open) return;
    const timer = setTimeout(handleLogout, AUTO_LOGOUT_MS);
    return () => clearTimeout(timer);
  }, [open]);

  if (!open) return null;

  return (
    <Dialog
      title={t('auth.sessionExpiredTitle')}
      onClose={handleLogout}
      dismissible={false}
      data-testid="session-expired-dialog"
    >
      <p className="mt-2 text-sm text-gray-600 dark:text-gray-300">
        {t('auth.sessionExpiredMessage')}
      </p>
      <button
        data-testid="session-expired-logout"
        onClick={handleLogout}
        className="mt-4 w-full rounded-md bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700 min-h-[44px]"
      >
        {t('auth.sessionExpiredAction')}
      </button>
    </Dialog>
  );
}
