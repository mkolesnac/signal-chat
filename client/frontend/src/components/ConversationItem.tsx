import * as React from 'react'
import Box from '@mui/joy/Box'
import ListDivider from '@mui/joy/ListDivider'
import ListItem from '@mui/joy/ListItem'
import ListItemButton from '@mui/joy/ListItemButton'
import Stack from '@mui/joy/Stack'
import Typography from '@mui/joy/Typography'
import UserAvatar from './UserAvatar'
import { useNavigate, useParams } from 'react-router-dom'
import { useUser } from '../hooks/useUser'
import Timestamp from './Timestamp'
import { models } from '../../wailsjs/go/models'
import Conversation = models.Conversation
import { Skeleton } from '@mui/joy'

type ConversationItemProps = {
  conversation: Conversation
};

export default function ConversationItem(props: ConversationItemProps) {
  const { conversation } = props
  const { conversationId: activeConversationId } = useParams()
  const selected = activeConversationId === conversation.ID
  const navigate = useNavigate()
  const { data: sender, isLoading, error } = useUser(conversation.LastMessageSenderID)

  return (
    <React.Fragment>
      <ListItem>
        <ListItemButton
          onClick={() => {
            navigate(`/${conversation.ID}`)
          }}
          selected={selected}
          color='neutral'
          sx={{ flexDirection: 'column', alignItems: 'initial', gap: 1 }}
        >
          <Stack direction='row' spacing={1.5}>
            {isLoading ? (
              <>
                <Skeleton animation="wave" variant="circular" width={32} height={32} />
                <Skeleton animation="wave" variant="text" sx={{ width: 120 }} />
              </>
            ) : (
              <>
                <UserAvatar username={sender!.Username} />
                <Typography level='title-sm'>{sender!.Username}</Typography>
              </>
            )}
            <Box sx={{ lineHeight: 1.5, textAlign: 'right' }}>
              {/*{unread && (*/}
              {/*  <CircleIcon sx={{ fontSize: 12 }} color="primary" />*/}
              {/*)}*/}
              <Timestamp value={conversation.LastMessageTimestamp} sx={{ display: { xs: 'none', md: 'block' } }} />
            </Box>
          </Stack>
          <Typography
            level='body-sm'
            sx={{
              display: '-webkit-box',
              WebkitLineClamp: '2',
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
            }}
          >
            {conversation.LastMessagePreview}
          </Typography>
        </ListItemButton>
      </ListItem>
      <ListDivider sx={{ margin: 0 }} />
    </React.Fragment>
  )
}