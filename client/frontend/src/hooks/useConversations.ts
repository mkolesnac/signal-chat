import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ListConversations } from '../../wailsjs/go/main/ConversationService'
import { useEffect } from 'react'
import { EventsOff, EventsOn } from '../../wailsjs/runtime'
import { models } from '../../wailsjs/go/models'
import Conversation = models.Conversation

export function useConversations() {
  const queryClient = useQueryClient()

  useEffect(() => {
    EventsOn('conversation-added', (value: Conversation) => {
      queryClient.setQueryData(['conversations'], (old: Conversation[] | undefined) => {
        return old ? [...old, value] : [value]
      })
    })

    return () => {
      EventsOff('conversation-added')
    }
  })

  return useQuery({
    queryKey: ['conversations'],
    queryFn: ListConversations,
    staleTime: Infinity
  })
}