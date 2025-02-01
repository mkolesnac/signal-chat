import * as React from 'react';
import { Navigate, Outlet } from 'react-router-dom'
import Sidebar from '../components/Sidebar'
import Header from '../components/Header'
import Box from '@mui/joy/Box'
import { ChatProps } from '../types'
import { chats } from '../data'
import Sheet from '@mui/joy/Sheet'
import ChatsPane from '../components/ChatsPane'
import MessagesPane from './MessagesPane'

export default function ChatLayout() {
  const [selectedChat, setSelectedChat] = React.useState<ChatProps>(chats[0]);
  return (
    <Sheet
      sx={{
        width: '100%',
        height: '100dvh',
        display: 'grid',
        gridTemplateColumns: {
          xs: '1fr',
          sm: 'minmax(min-content, min(30%, 400px)) 1fr',
        },
      }}
    >
      <Sheet
        sx={{
          position: { xs: 'fixed', sm: 'sticky' },
          transform: {
            xs: 'translateX(calc(100% * (var(--MessagesPane-slideIn, 0) - 1)))',
            sm: 'none',
          },
          transition: 'transform 0.4s, width 0.4s',
          zIndex: 100,
          width: '100%'
        }}
      >
        <ChatsPane
          chats={chats}
          selectedChatId={selectedChat.id}
          setSelectedChat={setSelectedChat}
        />
      </Sheet>
      <Box component="main" className="MainContent">
        <Outlet/>
        {/*<MessagesPane chat={selectedChat} />*/}
      </Box>
    </Sheet>
  );

  // return (
  //   <Box sx={{ display: 'flex', minHeight: '100dvh' }}>
  //     <Box component="main" className="MainContent" sx={{ flex: 1 }}>
  //       <MyMessages />
  //     </Box>
  //   </Box>
  // );
}

