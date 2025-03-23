import * as React from 'react'
import Avatar from '@mui/joy/Avatar'
import Box from '@mui/joy/Box'
import Stack, { StackProps } from '@mui/joy/Stack'
import Sheet from '@mui/joy/Sheet'
import Typography from '@mui/joy/Typography'
import InsertDriveFileRoundedIcon from '@mui/icons-material/InsertDriveFileRounded'
import { MessageProps } from '../types'
import { useAuth } from '../contexts/AuthContext'
import { useUser } from '../hooks/useUser'
import { format } from 'date-fns'
import Timestamp from './Timestamp'
import UserAvatar from './UserAvatar'
import { Card, CardContent, Skeleton, Theme } from '@mui/joy'
import { models } from '../../wailsjs/go/models'
import Message = models.Message
import { SxProps } from '@mui/joy/styles/types'
import Divider from '@mui/joy/Divider'

type ChatBubbleProps = {
  message: Message
  sx?: SxProps;
}

export default function ChatMessage(props: ChatBubbleProps) {
  const { message, sx} = props
  const {user: me} = useAuth()
  const fromMe = !message.SenderID
  const { data: sender, isLoading, error } = useUser(message.SenderID)

  if (isLoading) {
    return (
      <Stack direction='row' spacing={2}>
        <Skeleton animation="wave" variant="circular" width={32} height={32} />
        <div>
          <Skeleton animation="wave" variant="text" sx={{ width: 120 }} />
          <Skeleton
            animation="wave"
            variant="text"
            level="body-sm"
            sx={{ width: 200 }}
          />
        </div>
      </Stack>
    )
  }

  return (
    <Stack
      direction="row"
      spacing={2}
      sx={[
        {
          flexDirection: fromMe ? 'row-reverse' : 'row'
        },
        ...(Array.isArray(sx) ? sx : [sx]),
      ]}
    >
      {!fromMe && (
        <UserAvatar id={sender!.ID}/>
      )}
      <Box sx={{ maxWidth: '60%', minWidth: 'auto' }}>
        <Stack
          direction="row"
          spacing={2}
          sx={{ justifyContent: 'space-between', mb: 0.25 }}
        >
          <Typography level="body-xs">
            {sender?.Username}
          </Typography>
          <Timestamp value={message.Timestamp}/>
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
        <Card
          variant="soft"
          sx={[
            {
              borderRadius: 'lg',
            },
            fromMe
              ? {
                borderTopRightRadius: 0,
              }
              : {
                borderTopRightRadius: 'lg',
              },
            fromMe
              ? {
                borderTopLeftRadius: 'lg',
              }
              : {
                borderTopLeftRadius: 0,
              },
            // fromMe
            //   ? {
            //     backgroundColor: 'var(--joy-palette-primary-solidBg)',
            //   }
            //   : {
            //     //backgroundColor: 'background.body',
            //     backgroundColor: 'background.level2',
            //   },
          ]}
        >
          <CardContent sx={{rowGap: 1}}>
            <Typography textColor="inherit">{message.Text}</Typography>
            <Divider/>
            <Box sx={{ display: 'grid', gridTemplateColumns: '120px minmax(0, 1fr)', rowGap: 0.5 }}>
              <Typography level="body-xs" textColor="inherit">Ciphertext:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace" component="p" sx={{wordWrap: "break-word"}}>{message.Ciphertext}</Typography>

              <Typography level="body-xs" textColor="inherit">Version:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace">{message.Envelope?.Version }</Typography>
              <Typography level="body-xs" textColor="inherit">Message type:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace">{message.Envelope?.MessageType === 2 ? "Whisper" : "PreKey" }</Typography>
              <Typography level="body-xs" textColor="inherit">Counter:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace">{message.Envelope?.Counter }</Typography>
              <Typography level="body-xs" textColor="inherit" >PreviousCounter:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace">{message.Envelope?.PreviousCounter }</Typography>
              <Typography level="body-xs" textColor="inherit">RatchetKey:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace" component="p" sx={{wordWrap: "break-word"}}>{message.Envelope?.RatchetKey}</Typography>
              <Typography level="body-xs" textColor="inherit">SenderRatchetKey:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace" component="p" sx={{wordWrap: "break-word"}}>{message.Envelope?.SenderRatchetKey}</Typography>
              <Typography level="body-xs" textColor="inherit">Mac:</Typography>
              <Typography level="body-xs" textColor="inherit" fontFamily="monospace" component="p" sx={{wordWrap: "break-word"}}>{message.Envelope?.Mac}</Typography>
              {message.Envelope?.MessageType !== 2 && (
                <>
                  <Typography level="body-xs" textColor="inherit">RegistrationID:</Typography>
                  <Typography level="body-xs" textColor="inherit" fontFamily="monospace">{message.Envelope?.RegistrationID}</Typography>
                  <Typography level="body-xs" textColor="inherit" >SignedPreKeyID:</Typography>
                  <Typography level="body-xs" textColor="inherit" fontFamily="monospace">{message.Envelope?.SignedPreKeyID}</Typography>
                  <Typography level="body-xs" textColor="inherit" >PreKeyID:</Typography>
                  <Typography level="body-xs" textColor="inherit" fontFamily="monospace">{message.Envelope?.PreKeyID}</Typography>
                  <Typography level="body-xs" textColor="inherit" >IdentityKey:</Typography>
                  <Typography level="body-xs" textColor="inherit" fontFamily="monospace" component="p" sx={{wordWrap: "break-word"}}>{message.Envelope?.IdentityKey}</Typography>
                  <Typography level="body-xs" textColor="inherit" >BaseKey:</Typography>
                  <Typography level="body-xs" textColor="inherit" fontFamily="monospace" component="p" sx={{wordWrap: "break-word"}}>{message.Envelope?.BaseKey}</Typography>
                </>
              )}
            </Box>

          </CardContent>
        </Card>
        {/*<Sheet*/}
        {/*  color={fromMe ? 'primary' : 'neutral'}*/}
        {/*  variant={fromMe ? 'solid' : 'soft'}*/}
        {/*  sx={[*/}
        {/*    {*/}
        {/*      borderRadius: 'lg',*/}
        {/*    },*/}
        {/*    fromMe*/}
        {/*      ? {*/}
        {/*        borderTopRightRadius: 0,*/}
        {/*      }*/}
        {/*      : {*/}
        {/*        borderTopRightRadius: 'lg',*/}
        {/*      },*/}
        {/*    fromMe*/}
        {/*      ? {*/}
        {/*        borderTopLeftRadius: 'lg',*/}
        {/*      }*/}
        {/*      : {*/}
        {/*        borderTopLeftRadius: 0,*/}
        {/*      },*/}
        {/*    fromMe*/}
        {/*      ? {*/}
        {/*        backgroundColor: 'var(--joy-palette-primary-solidBg)',*/}
        {/*      }*/}
        {/*      : {*/}
        {/*        //backgroundColor: 'background.body',*/}
        {/*        backgroundColor: 'background.level2',*/}
        {/*      },*/}
        {/*  ]}*/}
        {/*>*/}
        {/*  <Box sx={{p: 1.25}}>*/}
        {/*    <Typography*/}
        {/*      level="body-sm"*/}
        {/*      sx={[*/}
        {/*        fromMe*/}
        {/*          ? {*/}
        {/*            color: 'var(--joy-palette-common-white)',*/}
        {/*          }*/}
        {/*          : {*/}
        {/*            color: 'var(--joy-palette-text-primary)',*/}
        {/*          },*/}
        {/*      ]}*/}
        {/*    >*/}
        {/*      {message.Text}*/}
        {/*    </Typography>*/}
        {/*  </Box>*/}
        {/*</Sheet>*/}
      </Box>
    </Stack>
  )
}
