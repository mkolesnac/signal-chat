import * as React from 'react'
import { useEffect, useState } from 'react'
import Stack from '@mui/joy/Stack'
import Sheet from '@mui/joy/Sheet'
import Typography from '@mui/joy/Typography'
import { Alert, Box, IconButton } from '@mui/joy'
import List from '@mui/joy/List'
import ConversationItem from './ConversationItem'
import { useAuth } from '../contexts/AuthContext'
import Button from '@mui/joy/Button'
import NewConversationDialog from './NewConversationDialog'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CreateConversation, ListConversations } from '../../wailsjs/go/main/ConversationService'
import { EventsOff, EventsOn } from '../../wailsjs/runtime'
import { models } from '../../wailsjs/go/models'
import { Add } from '@mui/icons-material'
import UserProfileButton from './UserProfileButton'
import Conversation = models.Conversation
import User = models.User
import { toggleMessagesPane } from '../utils'
import CloseRoundedIcon from '@mui/icons-material/CloseRounded';

function sortConversations(conversations: Conversation[]) {
  return conversations.sort(
    (a, b) => a.LastMessageTimestamp - b.LastMessageTimestamp,
  )
}

export default function Sidebar() {
  const { user: me } = useAuth()
  const [newDialogOpen, setNewDialogOpen] = useState(false)
  const queryClient = useQueryClient()

  const {
    data: conversations,
    error,
  } = useQuery({
    queryKey: ['conversations'],
    queryFn: ListConversations,
    staleTime: Infinity,
  })

  const mutation = useMutation({
    mutationFn: async (variables: { recipientIds: string[] }) =>
      CreateConversation(variables.recipientIds),
    onSuccess: (newConversation) => {
      queryClient.setQueryData(
        ['conversations'],
        (old: Conversation[] | undefined) => {
          if (!old) return [newConversation]
          return sortConversations([...old, newConversation])
        },
      )
    },
  })

  useEffect(() => {
    EventsOn('conversation_added', (value: Conversation) => {
      queryClient.setQueryData(
        ['conversations'],
        (old: Conversation[] | undefined) => {
          return old ? [...old, value] : [value]
        },
      )
    })

    EventsOn('conversation_updated', (value: Conversation) => {
      queryClient.setQueryData(
        ['conversations'],
        (old: Conversation[] | undefined) => {
          return old!.map(conv => conv.ID === value.ID ? value : conv)
        },
      )
    })

    return () => {
      EventsOff('conversation_added')
    }
  })

  const handleCreateDialogAccept = (recipients: User[]) => {
    mutation.mutate({ recipientIds: recipients.map((u) => u.ID) })
    setNewDialogOpen(false)
  }

  return (
    <Sheet
      color='neutral'
      sx={{
        px: 2,
        py: 3,
        display: 'flex',
        flexDirection: 'column',
        gap: 3,
        borderRight: '1px solid',
        borderColor: 'divider',
        height: '100dvh',
        backgroundColor: 'background.level1',
      }}
    >
      <Stack
        direction="row"
        spacing={1}
        sx={{
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Box flex={"1 1 auto"}>
          <UserProfileButton/>
        </Box>
        <IconButton
          variant="plain"
          aria-label="edit"
          color="neutral"
          size="sm"
          onClick={() => {
            toggleMessagesPane();
          }}
          sx={{ display: { sm: 'none' } }}
        >
          <CloseRoundedIcon />
        </IconButton>
      </Stack>
      {!!error && (
        <Alert variant="soft" color="danger" size="lg">
          {(error as Error).message}
        </Alert>
      )}
      <Box flexGrow={1}>
        <Typography
          level="body-xs"
          sx={{ textTransform: 'uppercase', fontWeight: 'lg', mb: 1}}
        >
          Conversations
        </Typography>
        {conversations && (
          <List
            size='sm'
            sx={{
              '--ListItem-paddingY': '0.75rem',
              '--ListItem-paddingX': '1rem',
              overflowY: 'auto',
            }}
          >
            {conversations.map((conversation) => (
              <ConversationItem
                key={conversation.ID}
                conversation={conversation}
              />
            ))}
          </List>
        )}
      </Box>
      <Box px={2}>
        <Button fullWidth size='md' color='primary' startDecorator={<Add />} onClick={() => setNewDialogOpen(true)}>
          New conversation
        </Button>
      </Box>
      <NewConversationDialog
        open={newDialogOpen}
        onClose={() => setNewDialogOpen(false)}
        onAccept={handleCreateDialogAccept}
      />
    </Sheet>
  )
}
