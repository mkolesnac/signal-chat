import { useQuery } from '@tanstack/react-query'
import { GetUser } from '../../wailsjs/go/main/UserService'
import { useAuth } from '../contexts/AuthContext'

export function useUser(userId: string | undefined) {
  const {user: me} = useAuth()

  return useQuery({
    queryKey: ['users', userId],
    queryFn: async () => {
      console.log("fetching user with ID: %o", userId)
      if (!userId || userId === me?.ID) {
        console.log("returning me")
        return me
      }
      return await GetUser(userId!)
    },
    //enabled: !!userId,
    staleTime: 5 * 60 * 1000,  // Consider data fresh for 5 minutes
  })
}