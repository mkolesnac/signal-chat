import * as React from 'react'
import { useEffect, useState } from 'react'
import Stack from '@mui/joy/Stack'
import Sheet from '@mui/joy/Sheet'
import Typography from '@mui/joy/Typography'
import { Alert, Box, IconButton } from '@mui/joy'
import List from '@mui/joy/List'
import CloseRoundedIcon from '@mui/icons-material/CloseRounded'
import ConversationItem from './ConversationItem'
import { toggleMessagesPane } from '../utils'
import Avatar from '@mui/joy/Avatar'
import { useAuth } from '../contexts/AuthContext'
import Button from '@mui/joy/Button'
import NewConversationDialog from './NewConversationDialog'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CreateConversation, ListConversations } from '../../wailsjs/go/main/ConversationService'
import { EventsOff, EventsOn } from '../../wailsjs/runtime'
import { models } from '../../wailsjs/go/models'
import Conversation = models.Conversation
import User = models.User
import UserAvatar from './UserAvatar'
import AddCommentIcon from '@mui/icons-material/AddComment';
import AddCircleIcon from '@mui/icons-material/AddCircle';
import AddCircleOutlinedIcon from '@mui/icons-material/AddCircleOutlined';
import AddBoxIcon from '@mui/icons-material/AddBox';
import { Add } from '@mui/icons-material'
import UserProfileButton from './UserProfileButton'

function sortConversations(conversations: Conversation[]) {
  return conversations.sort(
    (a, b) => a.LastMessageTimestamp - b.LastMessageTimestamp,
  )
}

export default function ConversationsPane() {
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
    mutationFn: async (variables: { name: string; participantIds: string[] }) =>
      CreateConversation(variables.name, variables.participantIds),
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

  const handleCreateDialogAccept = (name: string, user: User[]) => {
    mutation.mutate({ name, participantIds: user.map((u) => u.ID) })
    setNewDialogOpen(false)
  }

  return (
    <Sheet
      sx={{
        borderRight: '1px solid',
        borderColor: 'divider',
        height: '100dvh',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: 'background.level1',
      }}
    >
      <Stack
        direction="row"
        spacing={1}
        sx={{
          alignItems: 'center',
          justifyContent: 'space-between',
          p: 2,
          pb: 1.5,
        }}
      >
        <Box flex={"1 1 auto"}>
          <UserProfileButton size={"md"}/>
        </Box>
        <Button size='sm' color='primary' startDecorator={<Add />} onClick={() => setNewDialogOpen(true)}>
          New
        </Button>

        {/*<IconButton*/}
        {/*  variant="plain"*/}
        {/*  aria-label="edit"*/}
        {/*  color="neutral"*/}
        {/*  size="sm"*/}
        {/*  sx={{ display: { xs: 'none', sm: 'unset' } }}*/}
        {/*>*/}
        {/*  <EditNoteRoundedIcon />*/}
        {/*</IconButton>*/}
        <IconButton
          variant="plain"
          aria-label="edit"
          color="neutral"
          size="sm"
          onClick={() => {
            toggleMessagesPane()
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
      {conversations && (
        <List
          sx={{
            mt: 0.5,
            py: 0,
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
      <NewConversationDialog
        open={newDialogOpen}
        onClose={() => setNewDialogOpen(false)}
        onAccept={handleCreateDialogAccept}
      />
    </Sheet>
  )
}
