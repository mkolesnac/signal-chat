import * as React from 'react'
import Box from '@mui/joy/Box'
import ListDivider from '@mui/joy/ListDivider'
import ListItem from '@mui/joy/ListItem'
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

type ConversationItemProps = {
  conversation: Conversation
}

export default function ConversationItem({
  conversation,
}: ConversationItemProps) {
  const { conversationId: activeConversationId } = useParams()
  const selected = activeConversationId === conversation.ID
  const navigate = useNavigate()

  return (
    <React.Fragment>
      <ListItem>
        <ListItemButton
          onClick={() => {
            navigate(`/${conversation.ID}`)
          }}
          selected={selected}
          color="neutral"
          sx={{ flexDirection: 'row', alignItems: 'center', gap: 1 }}
        >
          <AvatarGroup>
            {conversation.OtherParticipantIDs.map((id, i) => (
              <UserAvatar key={`${id}-${i}`} id={id} size={'sm'} />
            ))}
            {conversation.OtherParticipantIDs.length > 2 && (
              <Avatar size={'sm'}>+{conversation.OtherParticipantIDs.length - 2}</Avatar>
            )}
          </AvatarGroup>
          <Box flexGrow={1}>
            <Typography level="title-sm" sx={{ flex: '1 1 auto' }}>
              {conversation.Name}
            </Typography>
            <Stack direction="row" spacing={1} alignItems="center" mt={1}>
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

          {/*<Stack direction='row' spacing={1.5}>*/}
          {/*  {isLoading ? (*/}
          {/*    <>*/}

          {/*    </>*/}
          {/*  ) : (*/}
          {/*    <>*/}

          {/*    </>*/}
          {/*  )}*/}
          {/*  */}
          {/*</Stack>*/}
        </ListItemButton>
      </ListItem>
      <ListDivider sx={{ margin: 0 }} />
    </React.Fragment>
  )
}
