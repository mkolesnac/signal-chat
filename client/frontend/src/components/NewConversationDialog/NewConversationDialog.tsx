import React, { useState, useEffect } from 'react'
import {
  Button,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  FormLabel,
  Autocomplete,
  Chip,
  Stack,
  Typography,
  CircularProgress,
  Input,
  Modal,
  ModalDialog,
  IconButton,
} from '@mui/joy'
import { SxProps } from '@mui/joy/styles/types'
import UserAutocomplete from './UserAutocomplete'
import Close from '@mui/icons-material/Close'
import List from '@mui/joy/List'
import { models } from '../../../wailsjs/go/models'
import User = models.User
import ListItem from '@mui/joy/ListItem'
import UserAvatar from '../UserAvatar'
import Box from '@mui/joy/Box'
import { SignIn } from '../../../wailsjs/go/main/Auth'
import Conversation = models.Conversation

type NewConversationDialogProps = {
  open: boolean
  onClose: () => void
  onAccept: (name: string, participants: User[]) => void
}

interface FormElements extends HTMLFormControlsCollection {
  name: HTMLInputElement
}
interface NewConversationForm extends HTMLFormElement {
  readonly elements: FormElements
}

const NewConversationDialog = ({
  open,
  onClose,
  onAccept,
}: NewConversationDialogProps) => {
  const [selectedUsers, setSelectedUsers] = useState<User[]>([])

  const handleUserSelected = (user: User) => {
    console.log('handleUserSelected: user: %o', user)
    if (!selectedUsers.find((u) => u.ID === user.ID)) {
      setSelectedUsers((prev) => [...prev, user])
    }
  }

  const handleUserRemoved = (user: User) => {
    setSelectedUsers((prev) => prev.filter((u) => u.ID !== user.ID))
  }

  const handleSubmit = async (event: React.FormEvent<NewConversationForm>) => {
    event.preventDefault()

    const name = event.currentTarget.elements.name.value
    onAccept(name, selectedUsers)
  }

  return (
    <Modal open={open} onClose={onClose}>
      <ModalDialog minWidth={400}>
        <DialogTitle>Create new conversation</DialogTitle>
        <DialogContent>
          Fill in the information of the conversation.
        </DialogContent>
        <form onSubmit={handleSubmit}>
          <Stack spacing={2}>
            <FormControl>
              <FormLabel>Name</FormLabel>
              <Input
                autoFocus
                required
                name="name"
                slotProps={{
                  input: { autoComplete: 'new-password' },
                }}
              />
            </FormControl>
            <FormControl>
              <FormLabel>Add participants</FormLabel>
              <UserAutocomplete onUserClick={handleUserSelected} />
            </FormControl>
            <List sx={{ p: 0 }}>
              {selectedUsers.map((usr) => (
                <ListItem key={usr.ID}>
                  <UserAvatar id={usr.ID} />
                  <Typography level="body-sm" sx={{ flex: '1 1 auto' }}>
                    {usr.Username}
                  </Typography>
                  <IconButton onClick={() => handleUserRemoved(usr)}>
                    <Close />
                  </IconButton>
                </ListItem>
              ))}
            </List>
            <Button type="submit">Submit</Button>
          </Stack>
        </form>
      </ModalDialog>
    </Modal>
  )
}

export default NewConversationDialog
