import { useQuery } from '@tanstack/react-query'
import { GetUser } from '../../wailsjs/go/main/UserService'
import { useAuth } from '../contexts/AuthContext'

export function useUser(userId: string | undefined) {
  const {user: me} = useAuth()

  return useQuery({
    queryKey: ['users', userId],
    queryFn: async () => {
      if (userId === me?.ID) {
        return me
      }
      return await GetUser(userId!)
    },
    enabled: !!userId,
    staleTime: 5 * 60 * 1000,  // Consider data fresh for 5 minutes
  })
}