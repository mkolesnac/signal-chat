import * as React from 'react'
import Avatar from '@mui/joy/Avatar'
import Box from '@mui/joy/Box'
import Stack from '@mui/joy/Stack'
import Sheet from '@mui/joy/Sheet'
import Typography from '@mui/joy/Typography'
import InsertDriveFileRoundedIcon from '@mui/icons-material/InsertDriveFileRounded'
import { MessageProps } from '../types'
import { main } from '../../wailsjs/go/models'
import Message = main.Message

type ChatBubbleProps = {
  message: Message
  variant: 'sent' | 'received'
}

export default function ChatBubble(props: ChatBubbleProps) {
  const { message, variant } = props
  const isSent = variant === 'sent'

  return (
    <Box sx={{ maxWidth: '60%', minWidth: 'auto' }}>
      <Stack
        direction="row"
        spacing={2}
        sx={{ justifyContent: 'space-between', mb: 0.25 }}
      >
        <Typography level="body-xs">
          {message.SenderID === 'You' ? 'sender' : 'sender.name'}
        </Typography>
        <Typography level="body-xs">timestamp</Typography>
      </Stack>
      {/*{attachment ? (*/}
      {/*  <Sheet*/}
      {/*    variant="outlined"*/}
      {/*    sx={[*/}
      {/*      {*/}
      {/*        px: 1.75,*/}
      {/*        py: 1.25,*/}
      {/*        borderRadius: 'lg',*/}
      {/*      },*/}
      {/*      isSent*/}
      {/*        ? { borderTopRightRadius: 0 }*/}
      {/*        : { borderTopRightRadius: 'lg' },*/}
      {/*      isSent ? { borderTopLeftRadius: 'lg' } : { borderTopLeftRadius: 0 },*/}
      {/*    ]}*/}
      {/*  >*/}
      {/*    <Stack direction="row" spacing={1.5} sx={{ alignItems: 'center' }}>*/}
      {/*      <Avatar color="primary" size="lg">*/}
      {/*        <InsertDriveFileRoundedIcon />*/}
      {/*      </Avatar>*/}
      {/*      <div>*/}
      {/*        <Typography sx={{ fontSize: 'sm' }}>*/}
      {/*          {attachment.fileName}*/}
      {/*        </Typography>*/}
      {/*        <Typography level="body-sm">{attachment.size}</Typography>*/}
      {/*      </div>*/}
      {/*    </Stack>*/}
      {/*  </Sheet>*/}
      {/*) : (*/}
        <Sheet
          color={isSent ? 'primary' : 'neutral'}
          variant={isSent ? 'solid' : 'soft'}
          sx={[
            {
              p: 1.25,
              borderRadius: 'lg',
            },
            isSent
              ? {
                  borderTopRightRadius: 0,
                }
              : {
                  borderTopRightRadius: 'lg',
                },
            isSent
              ? {
                  borderTopLeftRadius: 'lg',
                }
              : {
                  borderTopLeftRadius: 0,
                },
            isSent
              ? {
                  backgroundColor: 'var(--joy-palette-primary-solidBg)',
                }
              : {
                  //backgroundColor: 'background.body',
                  backgroundColor: 'background.level2',
                },
          ]}
        >
          <Typography
            level="body-sm"
            sx={[
              isSent
                ? {
                    color: 'var(--joy-palette-common-white)',
                  }
                : {
                    color: 'var(--joy-palette-text-primary)',
                  },
            ]}
          >
            {message.Text}
          </Typography>
        </Sheet>
      {/*)}*/}
    </Box>
  )
}
