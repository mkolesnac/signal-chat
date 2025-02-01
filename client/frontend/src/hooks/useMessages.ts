import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ListConversations, ListMessages } from '../../wailsjs/go/main/ConversationService'
import { useEffect } from 'react'
import { EventsOff, EventsOn } from '../../wailsjs/runtime'
import { main } from '../../wailsjs/go/models'
import Conversation = main.Conversation

export function useMessages(conversationId: string | undefined) {
  const queryClient = useQueryClient()

  useEffect(() => {
    EventsOn('message-added', (value: Conversation) => {
      queryClient.setQueryData(['messages', conversationId], (old: Conversation[] | undefined) => {
        return old ? [...old, value] : [value]
      })
    })

    return () => {
      EventsOff('message-added')
    }
  })

  return useQuery({
    queryKey: ['messages', conversationId],
    queryFn: ({queryKey}) => ListMessages(queryKey[1]!),
    staleTime: Infinity
  })
}