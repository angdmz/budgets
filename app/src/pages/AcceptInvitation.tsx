import { useState, useEffect } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import { createApiClient } from '../lib/api';
import { formatDate } from '../lib/format';
import type { InvitationDetail } from '../lib/types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

export default function AcceptInvitation() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, isLoading: isAuthLoading, loginWithRedirect, getAccessTokenSilently } = useAuth0();
  const { t } = useTranslation();

  const [detail, setDetail] = useState<InvitationDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    async function fetchInvitation() {
      if (!token) {
        setError(t('invitation.errorInvalid'));
        setLoading(false);
        return;
      }
      try {
        const response = await axios.get<InvitationDetail>(`${API_BASE_URL}/invitations/token/${token}`);
        setDetail(response.data);
      } catch (err: any) {
        const status = err.response?.status;
        if (status === 404) {
          setError(t('invitation.errorNotFound'));
        } else if (status === 410) {
          setError(t('invitation.errorExpired'));
        } else {
          setError(t('invitation.errorLoadFailed'));
        }
      } finally {
        setLoading(false);
      }
    }
    fetchInvitation();
  }, [token]);

  const handleAccept = async () => {
    if (!token) return;
    if (!isAuthenticated) {
      loginWithRedirect({ appState: { returnTo: location.pathname } });
      return;
    }
    setAccepting(true);
    setError(null);
    try {
      const api = await createApiClient(getAccessTokenSilently);
      await api.post(`/invitations/token/${token}/accept`);
      setSuccess(true);
      setTimeout(() => navigate('/groups'), 2000);
    } catch (err: any) {
      const status = err.response?.status;
      if (status === 409) {
        setError(t('invitation.errorAlreadyMember'));
      } else if (status === 410) {
        setError(t('invitation.errorExpiredOrRevoked'));
      } else if (status === 404) {
        setError(t('invitation.errorNotFound'));
      } else {
        setError(t('invitation.errorAcceptFailed'));
      }
    } finally {
      setAccepting(false);
    }
  };

  if (loading || isAuthLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">{t('invitation.loading')}</p>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 max-w-md w-full text-center">
          <div className="w-16 h-16 bg-green-100 dark:bg-green-900 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">{t('invitation.joined')}</h1>
          <p className="text-gray-600 dark:text-gray-400">{t('invitation.redirecting')}</p>
        </div>
      </div>
    );
  }

  if (error && !detail) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 max-w-md w-full text-center">
          <div className="w-16 h-16 bg-red-100 dark:bg-red-900 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">{t('invitation.unavailable')}</h1>
          <p className="text-gray-600 dark:text-gray-400 mb-6">{error}</p>
          <button
            onClick={() => navigate('/dashboard')}
            className="rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
          >
            {t('invitation.goToDashboard')}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-semibold text-gray-900 dark:text-white mb-2">{t('invitation.title')}</h1>
        <p className="text-gray-600 dark:text-gray-400 mb-6">{t('invitation.subtitle')}</p>

        {detail && (
          <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4 mb-6 space-y-3">
            <div>
              <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{t('invitation.group')}</p>
              <p className="text-gray-900 dark:text-white font-semibold mt-0.5">{detail.group_name}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{t('invitation.invitedBy')}</p>
              <p className="text-gray-900 dark:text-white mt-0.5">{detail.inviter_name}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{t('invitation.role')}</p>
              <p className="text-gray-900 dark:text-white capitalize mt-0.5">{detail.role}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">{t('invitation.expires')}</p>
              <p className="text-gray-900 dark:text-white mt-0.5">{formatDate(detail.expires_at)}</p>
            </div>
          </div>
        )}

        {error && (
          <div className="mb-4 rounded-md bg-red-50 dark:bg-red-900/30 p-3 text-sm text-red-700 dark:text-red-400">
            {error}
          </div>
        )}

        <button
          onClick={handleAccept}
          disabled={accepting}
          className="w-full rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {accepting ? t('invitation.joining') : t('invitation.joinGroup')}
        </button>
        <button
          onClick={() => navigate('/dashboard')}
          className="mt-3 w-full rounded-md bg-white dark:bg-gray-700 px-4 py-2 text-sm font-semibold text-gray-900 dark:text-white shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600"
        >
          {t('invitation.decline')}
        </button>
      </div>
    </div>
  );
}
