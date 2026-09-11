import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import { getAuthToken } from '@/lib/auth';
import notification from '../app/(tabs)/notification';

const API = process.env.EXPO_PUBLIC_API_URL;

async function authedFetch(path: string, options: RequestInit={}) {
    const res = await fetch(`${API}${path}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${await getAuthToken()}`,
            ...options.headers,
        }
    })
    const json = await res.json();
    if (!res.ok) {
        throw new Error(json.message || 'Request failed');
    }
    return json;
}

export function useNotifications() {
    return useQuery({
        queryKey: ['notifications'],
        queryFn: async () => {
            const json = await authedFetch('/api/admin/notifications')
            return json.data.notifications ?? json.data
        }
    })
}

export function useUnreadCount() {
  return useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: async () => {
      const json = await authedFetch('/api/notifications/unread-count')
      return json.data.count ?? json.data
    },
  })
}


export function useMarkAllRead() {
    const markedNotifications = useQueryClient()
    return useMutation({
        mutationFn: (id: string) => authedFetch('/api/notifications/read-all', {
            method: 'PUT'
        }),
        onSuccess: () => markedNotifications.invalidateQueries({ queryKey: ['notifications']})
    })
}

export function useMarkOneRead() {
    const markOneAsRead= useQueryClient()
    return useMutation({
        mutationFn: (id: string) => authedFetch('/api/notifications/${id}/read', {
            method: 'PUT'
        }),
        onSuccess: () => markOneAsRead.invalidateQueries({ 
            queryKey: ['notifications']
        })
    })
}


export function useDeleteSelectedNotification() {
    const useDeleteSelected = useQueryClient()
    return useMutation({
        mutationFn: (ids: string[]) => 
            authedFetch('/api/notifications/bulk', {
                method: 'DELETE',
                body: JSON.stringify({ ids })
            }),
            onSuccess: () => useDeleteSelected.invalidateQueries({ 
                queryKey: ['notifications']
            })
    })
}


