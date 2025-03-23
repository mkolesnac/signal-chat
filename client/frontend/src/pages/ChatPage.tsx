import Box from '@mui/joy/Box'
import MessagesPane from '../components/MessagesPane'
export default function ChatPage() {
  return (
    <Box
      sx={{
        width: '100%',
        height: '100%'
      }}
    >
      <MessagesPane/>
    </Box>
  )
}