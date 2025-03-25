import * as React from 'react'
import Box from '@mui/joy/Box'
import ListDivider from '@mui/joy/ListDivider'
import ListItem, { ListItemProps } from '@mui/joy/ListItem'
import ListItemButton from '@mui/joy/ListItemButton'
import Stack from '@mui/joy/Stack'
import Typography from '@mui/joy/Typography'
import { useNavigate, useParams } from 'react-router-dom'
import { useUser } from '../hooks/useUser'
import Timestamp from './Timestamp'
import { models } from '../../wailsjs/go/models'
import { AvatarGroup, Skeleton } from '@mui/joy'
import UserAvatar from './UserAvatar'
import Conversation = models.Conversation
import Avatar from '@mui/joy/Avatar'
import { useRecipients } from '../hooks/useRecipients'

type ConversationItemProps = ListItemProps & {
  conversation: Conversation
}

export default function ConversationItem({
  conversation,
  ...rest
}: ConversationItemProps) {
  const { conversationId: activeConversationId } = useParams()
  const selected = activeConversationId === conversation.ID
  const navigate = useNavigate()
  const {recipients} = useRecipients(conversation?.RecipientIDs)

  return (
    <ListItem sx={{"--ListItem-radius": "8px"}} {...rest}>
      <ListItemButton
        onClick={() => {
          navigate(`/${conversation.ID}`)
        }}
        selected={selected}
        color="neutral"
        sx={{ flexDirection: 'row', alignItems: 'flex-start', gap: 1 }}
      >
        <AvatarGroup>
          {conversation.RecipientIDs.map((id, i) => (
            <UserAvatar key={`${id}-${i}`} id={id} size={'sm'} />
          ))}
          {conversation.RecipientIDs.length > 2 && (
            <Avatar size={'sm'}>
              +{conversation.RecipientIDs.length - 2}
            </Avatar>
          )}
        </AvatarGroup>
        <Box flexGrow={1}>
          <Typography level="title-sm" sx={{ flex: '1 1 auto' }}>
            {recipients.map(r => r?.Username).join(', ')}
          </Typography>
          <Stack direction="row" spacing={1} alignItems="center" mt={0.75}>
            <Typography
              level="body-sm"
              sx={{
                flex: '1 1 auto',
                display: '-webkit-box',
                WebkitLineClamp: '2',
                WebkitBoxOrient: 'vertical',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
              }}
            >
              {conversation.LastMessagePreview}
            </Typography>
            {conversation.LastMessageTimestamp > 0 && (
              <Timestamp
                value={conversation.LastMessageTimestamp}
                sx={{ display: { xs: 'none', md: 'block' } }}
              />
            )}
          </Stack>
        </Box>
      </ListItemButton>
    </ListItem>
  )
}
