import * as React from 'react'
import Box from '@mui/joy/Box'
import { Skeleton } from '@mui/joy'

type ChatBubbleProps = {
  variant: 'sent' | 'received'
}

export default function ChaSkeleton() {
  return (
    <Box sx={{ maxWidth: '60%', minWidth: 'auto' }}>
      <Skeleton />{' '}
    </Box>
  )
}
