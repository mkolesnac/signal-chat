import * as React from 'react'
import Stack from '@mui/joy/Stack'
import Sheet from '@mui/joy/Sheet'
import Typography from '@mui/joy/Typography'
import { Box, Chip, IconButton, Input } from '@mui/joy'
import List from '@mui/joy/List'
import EditNoteRoundedIcon from '@mui/icons-material/EditNoteRounded'
import SearchRoundedIcon from '@mui/icons-material/SearchRounded'
import CloseRoundedIcon from '@mui/icons-material/CloseRounded'
import ConversationItem from './ConversationItem'
import { ChatProps } from '../types'
import { toggleMessagesPane } from '../utils'
import Divider from '@mui/joy/Divider'
import Avatar from '@mui/joy/Avatar'
import LogoutRoundedIcon from '@mui/icons-material/LogoutRounded'
import { useAuth } from '../contexts/AuthContext'
import { useConversations } from '../hooks/useConversations'

type ChatsPaneProps = {
  chats: ChatProps[]
  setSelectedChat: (chat: ChatProps) => void
  selectedChatId: string
}

export default function ChatsPane(props: ChatsPaneProps) {
  const { user } = useAuth()
  const { data: conversations, isLoading, error } = useConversations()
  const { chats, setSelectedChat, selectedChatId } = props



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
        <Typography
          component="h1"
          sx={{
            fontSize: { xs: 'md', md: 'lg' },
            fontWeight: 'lg',
            mr: 'auto',
          }}
        >
          Messages
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <Avatar
            variant="outlined"
            size="sm"
            src="https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&w=286"
          />
          <Box sx={{ minWidth: 0, flex: 1 }}>
            {/*@ts-ignore*/}
            <Typography level="title-sm">{user.Username}</Typography>
          </Box>
          {/*<IconButton size="sm" variant="plain" color="neutral">*/}
          {/*  <LogoutRoundedIcon />*/}
          {/*</IconButton>*/}
        </Box>

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
      <Box sx={{ px: 2, py: 1.5 }}>
        <Input
          size="sm"
          startDecorator={<SearchRoundedIcon />}
          placeholder="Search"
          aria-label="Search"
        />
      </Box>
      {isLoading && (
        <Typography>Loading</Typography>
      )}
      {!!error && (
        <Typography>Error</Typography>
      )}
      {conversations && (
        <List
          sx={{
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
    </Sheet>
  )
}
