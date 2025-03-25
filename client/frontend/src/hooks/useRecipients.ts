import { useQueries, useQuery, useQueryClient } from '@tanstack/react-query'
import { GetUser } from '../../wailsjs/go/main/UserService'
import { useAuth } from '../contexts/AuthContext'
import { models } from '../../wailsjs/go/models'
import User = models.User

export function useRecipients(recipientIds: string[] | undefined): {recipients: (User | undefined)[], isLoading: boolean} {
  const queryClient = useQueryClient()

  const queries = useQueries({
    queries: (recipientIds || []).map(id => ({
      queryKey: ['users', id],
      queryFn: async () => GetUser(id!),
      initialData: () => queryClient.getQueryData<User>(['users', id]),
      staleTime: 5 * 60 * 1000,  // Consider data fresh for 5 minutes
      enabled: !!recipientIds,
    })),
  });

  return {recipients: queries.map(query => query.data), isLoading: !!queries.find(query => query.isLoading)}
}