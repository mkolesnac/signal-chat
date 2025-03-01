import * as React from 'react'
import { Outlet } from 'react-router-dom'
import Box from '@mui/joy/Box'
import Sheet from '@mui/joy/Sheet'
import ConversationsPane from '../components/ConversationsPane'

export default function ChatLayout() {
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
        <ConversationsPane/>
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

