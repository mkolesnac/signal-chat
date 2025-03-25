import React, { useState } from 'react'
import {
  Button,
  DialogContent,
  DialogTitle,
  FormControl,
  FormLabel,
  IconButton,
  Modal,
  ModalDialog,
  Stack,
  Typography,
} from '@mui/joy'
import UserAutocomplete from './UserAutocomplete'
import Close from '@mui/icons-material/Close'
import List from '@mui/joy/List'
import { models } from '../../../wailsjs/go/models'
import ListItem from '@mui/joy/ListItem'
import UserAvatar from '../UserAvatar'
import User = models.User

type NewConversationDialogProps = {
  open: boolean
  onClose: () => void
  onAccept: (recipients: User[]) => void
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
    onAccept(selectedUsers)
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
