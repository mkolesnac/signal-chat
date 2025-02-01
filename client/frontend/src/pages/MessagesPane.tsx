import * as React from 'react';
import Box from '@mui/joy/Box';
import Sheet from '@mui/joy/Sheet';
import Stack from '@mui/joy/Stack';
import AvatarWithStatus from '../components/AvatarWithStatus';
import ChatBubble from '../components/ChatBubble';
import MessageInput from '../components/MessageInput';
import MessagesPaneHeader from '../components/MessagesPaneHeader';
import { ChatProps, MessageProps } from '../types';
import { useParams } from 'react-router-dom'
import { useConversations } from '../hooks/useConversations'
import { useMessages } from '../hooks/useMessages'
import Typography from '@mui/joy/Typography'
import { main } from '../../wailsjs/go/models'
import Message = main.Message

type MessagesPaneProps = {
};

export default function MessagesPane(props: MessagesPaneProps) {
  const { conversationId } = useParams()
  const { data: messages, isLoading, error } = useMessages(conversationId)
  const [textAreaValue, setTextAreaValue] = React.useState('');

  // React.useEffect(() => {
  //   setChatMessages(chat.messages);
  // }, [chat.messages]);

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
          {isLoading && (
            <Typography>Loading</Typography>
          )}
          {!!error && (
            <Typography>Error</Typography>
          )}
          {!!messages && messages.map((message: Message, index: number) => {
            const isYou = message.SenderID === 'You';
            return (
              <Stack
                key={index}
                direction="row"
                spacing={2}
                sx={{ flexDirection: isYou ? 'row-reverse' : 'row' }}
              >
                {/*{message.SenderID !== 'You' && (*/}
                {/*  <AvatarWithStatus*/}
                {/*    online={message.sender.online}*/}
                {/*    src={message.sender.avatar}*/}
                {/*  />*/}
                {/*)}*/}
                <ChatBubble variant={isYou ? 'sent' : 'received'} message={message}/>
              </Stack>
            );
          })}
        </Stack>
      </Box>
      {/*<MessageInput*/}
      {/*  textAreaValue={textAreaValue}*/}
      {/*  setTextAreaValue={setTextAreaValue}*/}
      {/*  onSubmit={() => {*/}
      {/*    const newId = chatMessages.length + 1;*/}
      {/*    const newIdString = newId.toString();*/}
      {/*    setChatMessages([*/}
      {/*      ...chatMessages,*/}
      {/*      {*/}
      {/*        id: newIdString,*/}
      {/*        sender: 'You',*/}
      {/*        content: textAreaValue,*/}
      {/*        timestamp: 'Just now',*/}
      {/*      },*/}
      {/*    ]);*/}
      {/*  }}*/}
      {/*/>*/}
    </Sheet>
  );
}