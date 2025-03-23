import * as React from 'react';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Chip from '@mui/joy/Chip';
import IconButton from '@mui/joy/IconButton';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import CircleIcon from '@mui/icons-material/Circle';
import ArrowBackIosNewRoundedIcon from '@mui/icons-material/ArrowBackIosNewRounded';
import PhoneInTalkRoundedIcon from '@mui/icons-material/PhoneInTalkRounded';
import MoreVertRoundedIcon from '@mui/icons-material/MoreVertRounded';
import { UserProps } from '../types';
import { toggleMessagesPane } from '../utils';
import { useNavigate, useParams } from 'react-router-dom'
import { useQueries, useQuery, useQueryClient } from '@tanstack/react-query'
import { models } from '../../wailsjs/go/models'
import Conversation = models.Conversation
import UserAvatar from './UserAvatar'
import { AvatarGroup } from '@mui/joy'
import Box from '@mui/joy/Box'
import User = models.User
import { GetUser } from '../../wailsjs/go/main/UserService'

type MessagesPaneHeaderProps = {
};


export default function MessagesPaneHeader({  }: MessagesPaneHeaderProps) {
  const { conversationId } = useParams()
  const queryClient = useQueryClient()
  const { data: conversation } = useQuery({
    queryKey: ['conversation', conversationId],
    queryFn: () => null, // No fetch needed, we'll select from cache
    select: () => {
      const conversations = queryClient.getQueryData<Conversation[]>(['conversations'])
      return conversations?.find(conversation => conversation.ID === conversationId)
    },
    enabled: !!conversationId && !!queryClient.getQueryData(['conversations']), // Only execute if the parent query succeeded
    staleTime: Infinity,
  })
  const participantQueries = useQueries({
    queries: (conversation?.OtherParticipantIDs || []).map(id => ({
      queryKey: ['users', id],
      queryFn: async () => GetUser(id!),
      initialData: () => queryClient.getQueryData<User>(['users', id]),
      staleTime: 5 * 60 * 1000,  // Consider data fresh for 5 minutes
      enabled: !!conversation,
    })),
  });

  const getParticipantNames = () => {
    const names = participantQueries.map(query => query.data?.Username).join(', ');
    return `You and ${names}`
  }

  if (!conversation) {
    return <div>Conversation not found</div>;
  }

  return (
    <Stack
      direction="row"
      sx={{
        justifyContent: 'space-between',
        py: { xs: 2, md: 2 },
        px: { xs: 1, md: 2 },
        borderBottom: '1px solid',
        borderColor: 'divider',
        backgroundColor: 'background.body',
      }}
    >
      <Stack
        direction="row"
        spacing={{ xs: 1, md: 2 }}
        sx={{ alignItems: 'center' }}
      >
        <IconButton
          variant="plain"
          color="neutral"
          size="sm"
          sx={{ display: { xs: 'inline-flex', sm: 'none' } }}
          onClick={() => toggleMessagesPane()}
        >
          <ArrowBackIosNewRoundedIcon />
        </IconButton>
        <AvatarGroup>
          {conversation.OtherParticipantIDs.map((id, i) => (
            <UserAvatar key={`${id}-${i}`} id={id} size={'md'} />
          ))}
          {conversation.OtherParticipantIDs.length > 2 && (
            <Avatar size={'md'}>
              +{conversation.OtherParticipantIDs.length - 2}
            </Avatar>
          )}
        </AvatarGroup>
        <Box>
            <Typography
              component="h2"
              noWrap
              sx={{ fontWeight: 'lg', fontSize: 'lg' }}
            >
              {conversation.Name}
            </Typography>
          <Typography level="body-sm">{getParticipantNames()}</Typography>
        </Box>
        {/*<div>*/}
        {/*  <Typography*/}
        {/*    component="h2"*/}
        {/*    noWrap*/}
        {/*    endDecorator={*/}
        {/*      sender.online ? (*/}
        {/*        <Chip*/}
        {/*          variant="outlined"*/}
        {/*          size="sm"*/}
        {/*          color="neutral"*/}
        {/*          sx={{ borderRadius: 'sm' }}*/}
        {/*          startDecorator={*/}
        {/*            <CircleIcon sx={{ fontSize: 8 }} color="success" />*/}
        {/*          }*/}
        {/*          slotProps={{ root: { component: 'span' } }}*/}
        {/*        >*/}
        {/*          Online*/}
        {/*        </Chip>*/}
        {/*      ) : undefined*/}
        {/*    }*/}
        {/*    sx={{ fontWeight: 'lg', fontSize: 'lg' }}*/}
        {/*  >*/}
        {/*    {sender.name}*/}
        {/*  </Typography>*/}
        {/*  <Typography level="body-sm">{sender.username}</Typography>*/}
        {/*</div>*/}
      </Stack>
      <Stack spacing={1} direction="row" sx={{ alignItems: 'center' }}>
        <IconButton size="sm" variant="plain" color="neutral">
          <MoreVertRoundedIcon />
        </IconButton>
      </Stack>
    </Stack>
  );
}