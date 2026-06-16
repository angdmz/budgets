import { useState, useEffect } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import axios from 'axios';
import { createApiClient } from '../lib/api';
import type { InvitationDetail } from '../lib/types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

export default function AcceptInvitation() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, loginWithRedirect, getAccessTokenSilently } = useAuth0();

  const [detail, setDetail] = useState<InvitationDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    async function fetchInvitation() {
      if (!token) {
        setError('Invalid invitation link.');
        setLoading(false);
        return;
      }
      try {
        const response = await axios.get<InvitationDetail>(`${API_BASE_URL}/invitations/token/${token}`);
        setDetail(response.data);
      } catch (err: any) {
        const status = err.response?.status;
        if (status === 404) {
          setError('Invitation not found.');
        } else if (status === 410) {
          setError('This invitation is no longer valid (expired or revoked).');
        } else {
          setError('Failed to load invitation details.');
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
        setError('You are already a member of this group.');
      } else if (status === 410) {
        setError('This invitation has expired or been revoked.');
      } else if (status === 404) {
        setError('Invitation not found.');
      } else {
        setError('Failed to accept invitation. Please try again.');
      }
    } finally {
      setAccepting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading invitation...</p>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="bg-white rounded-lg shadow p-8 max-w-md w-full text-center">
          <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold text-gray-900 mb-2">You've joined the group!</h1>
          <p className="text-gray-600">Redirecting to Groups...</p>
        </div>
      </div>
    );
  }

  if (error && !detail) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="bg-white rounded-lg shadow p-8 max-w-md w-full text-center">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold text-gray-900 mb-2">Invitation Unavailable</h1>
          <p className="text-gray-600 mb-6">{error}</p>
          <button
            onClick={() => navigate('/dashboard')}
            className="rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
          >
            Go to Dashboard
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white rounded-lg shadow p-8 max-w-md w-full">
        <h1 className="text-2xl font-semibold text-gray-900 mb-2">Group Invitation</h1>
        <p className="text-gray-600 mb-6">You've been invited to join a group.</p>

        {detail && (
          <div className="bg-gray-50 rounded-lg p-4 mb-6 space-y-3">
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Group</p>
              <p className="text-gray-900 font-semibold mt-0.5">{detail.group_name}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Invited by</p>
              <p className="text-gray-900 mt-0.5">{detail.inviter_name}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Role</p>
              <p className="text-gray-900 capitalize mt-0.5">{detail.role}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Expires</p>
              <p className="text-gray-900 mt-0.5">{new Date(detail.expires_at).toLocaleDateString()}</p>
            </div>
          </div>
        )}

        {error && (
          <div className="mb-4 rounded-md bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        <button
          onClick={handleAccept}
          disabled={accepting}
          className="w-full rounded-md bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {accepting ? 'Joining...' : 'Join Group'}
        </button>
        <button
          onClick={() => navigate('/dashboard')}
          className="mt-3 w-full rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
        >
          Decline
        </button>
      </div>
    </div>
  );
}
