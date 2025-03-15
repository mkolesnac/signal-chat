import * as React from 'react'
import Box from '@mui/joy/Box'
import Sheet from '@mui/joy/Sheet'
import Stack from '@mui/joy/Stack'
import ChatBubble from '../components/ChatBubble'
import MessageInput from '../components/MessageInput'
import { useParams } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { Alert, Skeleton } from '@mui/joy'
import { ListMessages, SendMessage } from '../../wailsjs/go/main/ConversationService'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { EventsOff, EventsOn } from '../../wailsjs/runtime'
import { models } from '../../wailsjs/go/models'
import Message = models.Message
import { useDebounce } from 'use-debounce'

function sortMessages(messages: Message[]) {
  return messages.sort((a, b) => a.Timestamp - b.Timestamp)
}

export default function MessagesPane() {
  const { conversationId } = useParams()
  const queryClient = useQueryClient()
  const {user: me} = useAuth()
  const [pendingMessage, setPendingMessage] = useState<Message | undefined>()

  const { data: messages, isLoading, isError, error } = useQuery({
    queryKey: ['messages', conversationId],
    queryFn: async ({queryKey}) => {
      const messages = await ListMessages(queryKey[1]!)
      console.log("fetched messages for conversation '%o': %o", queryKey[1], messages.length)

      return sortMessages(messages)
    },
    staleTime: Infinity,
    enabled: !!conversationId,
  })

  const mutation = useMutation({
    mutationFn: async (text: string) => SendMessage(conversationId!, text!),
    onSuccess: (newMessage) => {
      queryClient.setQueryData(['messages', conversationId], (old: Message[] | undefined) => {
        if (!old) return [newMessage]
        return sortMessages([...old, newMessage])
      })
    }
  })

  const [debouncedIsLoading] = useDebounce(isLoading, 300);

  useEffect(() => {
    EventsOn('message_added', (value: Message) => {
      queryClient.setQueryData(['messages', conversationId], (old: Message[] | undefined) => {
        if (!old) return [value]
        return sortMessages([...old, value])
      })
    })

    return () => {
      EventsOff('message_added')
    }
  })

  const handleSubmit = (text: string) => {
    const msg = new Message({
      ID: "temp",
      Text: text,
      SenderID: me?.ID,
      Timestamp: new Date().toISOString()
    })
    setPendingMessage(msg)
    mutation.mutate(text)
  }

  console.log("loading: %o, debouncedIsLoading: %o", isLoading, debouncedIsLoading)

  return (
    <Sheet
      sx={{
        height: '100dvh',
        display: 'flex',
        flexDirection: 'column'
      }}
    >
      {/*<MessagesPaneHeader sender={chat.sender} />*/}
      <Box
        sx={{
          display: 'flex',
          flex: 1,
          minHeight: 0,
          px: 2,
          py: 3,
          overflowY: 'scroll',
          flexDirection: 'column-reverse',
        }}
      >
        <Stack spacing={2} sx={{ justifyContent: 'flex-end' }}>
          {debouncedIsLoading && (
            <>
              <Skeleton variant="rectangular" width={0.4} height="3em" />
              <Skeleton variant="rectangular" width={0.4} height="3em" sx={{alignSelf: 'end'}}/>
            </>
          )}
          {!!error && (
            <Alert
              variant="soft"
              color="danger"
              size="lg"
            >
              {(error as Error).message}
            </Alert>
          )}
          {messages?.map(message => (
            <ChatBubble key={message.ID} message={message}/>
          ))}
          {mutation.isLoading && (
            <ChatBubble message={pendingMessage!} sx={{opacity: 0.5}}/>
          )}
        </Stack>
      </Box>
      <MessageInput
        onSubmit={handleSubmit}
      />
    </Sheet>
  );
}