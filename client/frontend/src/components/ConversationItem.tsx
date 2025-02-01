import * as React from 'react';
import Box from '@mui/joy/Box';
import ListDivider from '@mui/joy/ListDivider';
import ListItem from '@mui/joy/ListItem';
import ListItemButton, { ListItemButtonProps } from '@mui/joy/ListItemButton';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import CircleIcon from '@mui/icons-material/Circle';
import AvatarWithStatus from './AvatarWithStatus';
import { ChatProps, MessageProps, UserProps } from '../types';
import { toggleMessagesPane } from '../utils';
import { main } from '../../wailsjs/go/models'
import Conversation = main.Conversation
import { useNavigate } from 'react-router-dom'

type ConversationItemProps = ListItemButtonProps & {
  conversation: Conversation
  selectedChatId?: string;
};

export default function ConversationItem(props: ConversationItemProps) {
  const { conversation, selectedChatId } = props;
  const navigate = useNavigate();
  const selected = selectedChatId === conversation.ID;

  return (
    <React.Fragment>
      <ListItem>
        <ListItemButton
          onClick={() => {
            navigate(`/${conversation.ID}`)
          }}
          selected={selected}
          color="neutral"
          sx={{ flexDirection: 'column', alignItems: 'initial', gap: 1 }}
        >
          <Stack direction="row" spacing={1.5}>
            {/*<AvatarWithStatus online={sender.online} src={sender.avatar} />*/}
            <Box sx={{ flex: 1 }}>
              <Typography level="title-sm">sender.name</Typography>
              <Typography level="body-sm">sender.username</Typography>
            </Box>
            <Box sx={{ lineHeight: 1.5, textAlign: 'right' }}>
              {/*{unread && (*/}
              {/*  <CircleIcon sx={{ fontSize: 12 }} color="primary" />*/}
              {/*)}*/}
              <Typography
                level="body-xs"
                noWrap
                sx={{ display: { xs: 'none', md: 'block' } }}
              >
                5 mins ago
              </Typography>
            </Box>
          </Stack>
          <Typography
            level="body-sm"
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
  );
}