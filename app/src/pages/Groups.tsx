import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth0 } from '@auth0/auth0-react';
import { createApiClient } from '../lib/api';
import type { Budget, Group, CreateGroupRequest, Invitation } from '../lib/types';

export default function Groups() {
  const { getAccessTokenSilently } = useAuth0();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [formData, setFormData] = useState<CreateGroupRequest>({ name: '', description: '' });
  const [inviteModalGroupId, setInviteModalGroupId] = useState<string | null>(null);
  const [createdInviteLink, setCreatedInviteLink] = useState<string | null>(null);
  const [copySuccess, setCopySuccess] = useState(false);

  const { data: groups, isLoading } = useQuery({
    queryKey: ['groups'],
    queryFn: async () => {
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Group[]>('/groups');
      return response.data;
    },
  });

  const { data: budgets, isLoading: isBudgetsLoading } = useQuery({
    queryKey: ['budgets', selectedGroupId],
    queryFn: async () => {
      if (!selectedGroupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Budget[]>(`/groups/${selectedGroupId}/budgets`);
      return response.data;
    },
    enabled: !!selectedGroupId,
  });

  const createMutation = useMutation({
    mutationFn: async (data: CreateGroupRequest) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post('/groups', data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setIsModalOpen(false);
      setFormData({ name: '', description: '' });
    },
    onError: (error) => {
      console.error('[createGroup] mutation failed:', error);
    },
  });

  const { data: invitations, refetch: refetchInvitations } = useQuery({
    queryKey: ['invitations', inviteModalGroupId],
    queryFn: async () => {
      if (!inviteModalGroupId) return [];
      const api = await createApiClient(getAccessTokenSilently);
      const response = await api.get<Invitation[]>(`/groups/${inviteModalGroupId}/invitations`);
      return response.data;
    },
    enabled: !!inviteModalGroupId,
  });

  const createInvitationMutation = useMutation({
    mutationFn: async (groupId: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.post<Invitation>(`/groups/${groupId}/invitations`, { role: 'member' });
    },
    onSuccess: (response) => {
      const token = response.data.token;
      const link = `${window.location.origin}/app/invite/${token}`;
      setCreatedInviteLink(link);
      refetchInvitations();
    },
  });

  const revokeInvitationMutation = useMutation({
    mutationFn: async (invitationId: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/invitations/${invitationId}`);
    },
    onSuccess: () => {
      refetchInvitations();
    },
  });

  const handleCopyLink = () => {
    if (createdInviteLink) {
      navigator.clipboard.writeText(createdInviteLink);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    }
  };

  const openInviteModal = (groupId: string) => {
    setInviteModalGroupId(groupId);
    setCreatedInviteLink(null);
    setCopySuccess(false);
  };

  const closeInviteModal = () => {
    setInviteModalGroupId(null);
    setCreatedInviteLink(null);
    setCopySuccess(false);
  };

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      const api = await createApiClient(getAccessTokenSilently);
      return api.delete(`/groups/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(formData);
  };

  if (isLoading) {
    return <div className="text-center py-12">Loading...</div>;
  }

  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="sm:flex sm:items-center">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Groups</h1>
          <p className="mt-2 text-sm text-gray-700">
            Manage your budgeting groups
          </p>
        </div>
        <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
          <button
            onClick={() => { setIsModalOpen(true); createMutation.reset(); }}
            className="block rounded-md bg-primary-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
          >
            Add Group
          </button>
        </div>
      </div>

      <div className="mt-8 flow-root">
        <div className="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div className="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
            <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-300">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">
                      Name
                    </th>
                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Description
                    </th>
                    <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">
                      Created
                    </th>
                    <th className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                      <span className="sr-only">Actions</span>
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 bg-white">
                  {groups?.map((group) => (
                    <tr key={group.id}>
                      <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                        {group.name}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        {group.description || '-'}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                        {new Date(group.created_at).toLocaleDateString()}
                      </td>
                      <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                        <button
                          onClick={() => openInviteModal(group.id)}
                          className="text-primary-600 hover:text-primary-900 mr-4"
                        >
                          Invite
                        </button>
                        <button
                          onClick={() => deleteMutation.mutate(group.id)}
                          className="text-red-600 hover:text-red-900"
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <div className="mt-10">
        <h2 className="text-lg font-semibold text-gray-900">Budgets</h2>
        <p className="mt-1 text-sm text-gray-700">Select a group to view its budgets</p>
        <div className="mt-3">
          <label className="block text-sm font-medium text-gray-700">Select Group</label>
          <select
            value={selectedGroupId}
            onChange={(e) => setSelectedGroupId(e.target.value)}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
          >
            <option value="">Select a group...</option>
            {groups?.map((group) => (
              <option key={group.id} value={group.id}>{group.name}</option>
            ))}
          </select>
        </div>

        {selectedGroupId && (
          <div className="mt-6 flow-root">
            {isBudgetsLoading ? (
              <div className="text-center py-6 text-sm text-gray-500">Loading budgets...</div>
            ) : budgets && budgets.length > 0 ? (
              <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
                <table className="min-w-full divide-y divide-gray-300">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">Name</th>
                      <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Period</th>
                      <th className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Description</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 bg-white">
                    {budgets.map((budget) => (
                      <tr key={budget.id}>
                        <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">
                          {budget.name}
                        </td>
                        <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                          {new Date(budget.start_date).toLocaleDateString()} - {new Date(budget.end_date).toLocaleDateString()}
                        </td>
                        <td className="px-3 py-4 text-sm text-gray-500">{budget.description || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-6 text-sm text-gray-500">No budgets found for this group.</div>
            )}
          </div>
        )}
      </div>

      {/* Invite Modal */}
      {inviteModalGroupId && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-lg w-full">
            <h2 className="text-lg font-semibold mb-1">Invite to Group</h2>
            <p className="text-sm text-gray-600 mb-4">Generate a link to invite someone to join this group.</p>

            {!createdInviteLink ? (
              <button
                onClick={() => createInvitationMutation.mutate(inviteModalGroupId)}
                disabled={createInvitationMutation.isPending}
                className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
              >
                {createInvitationMutation.isPending ? 'Generating...' : 'Generate Invite Link'}
              </button>
            ) : (
              <div>
                <p className="text-sm font-medium text-gray-700 mb-2">Share this link (expires in 7 days):</p>
                <div className="flex gap-2">
                  <input
                    type="text"
                    readOnly
                    value={createdInviteLink}
                    className="flex-1 rounded-md border border-gray-300 bg-gray-50 px-3 py-2 text-sm text-gray-700"
                  />
                  <button
                    onClick={handleCopyLink}
                    className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500"
                  >
                    {copySuccess ? 'Copied!' : 'Copy'}
                  </button>
                </div>
              </div>
            )}

            {invitations && invitations.length > 0 && (
              <div className="mt-5">
                <h3 className="text-sm font-medium text-gray-700 mb-2">Existing Invitations</h3>
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {invitations.map((inv) => (
                    <div key={inv.id} className="flex items-center justify-between rounded-md bg-gray-50 px-3 py-2">
                      <div className="flex items-center gap-2 text-sm">
                        <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ${
                          inv.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                          inv.status === 'accepted' ? 'bg-green-100 text-green-800' :
                          'bg-red-100 text-red-800'
                        }`}>
                          {inv.status}
                        </span>
                        <span className="text-gray-600 capitalize">{inv.role}</span>
                        <span className="text-gray-400 text-xs">
                          expires {new Date(inv.expires_at).toLocaleDateString()}
                        </span>
                      </div>
                      {inv.status === 'pending' && (
                        <button
                          onClick={() => revokeInvitationMutation.mutate(inv.id)}
                          disabled={revokeInvitationMutation.isPending}
                          className="text-red-600 hover:text-red-900 text-xs font-medium disabled:opacity-50"
                        >
                          Revoke
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="mt-6 flex justify-end">
              <button
                onClick={closeInviteModal}
                className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h2 className="text-lg font-semibold mb-4">Create Group</h2>
            <form onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Name</label>
                  <input
                    type="text"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
                    rows={3}
                  />
                </div>
              </div>
              {createMutation.isError && (
                <p className="mt-2 text-sm text-red-600">
                  Failed to create group. Please try again.
                  {createMutation.error instanceof Error && (
                    <span className="block text-xs mt-1 opacity-75">{createMutation.error.message}</span>
                  )}
                </p>
              )}
              <div className="mt-6 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => { setIsModalOpen(false); createMutation.reset(); }}
                  className="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  disabled={createMutation.isPending}
                  onClick={() => createMutation.mutate(formData)}
                  className="rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 disabled:opacity-50"
                >
                  {createMutation.isPending ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
