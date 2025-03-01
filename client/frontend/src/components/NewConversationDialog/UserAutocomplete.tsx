import {
  Autocomplete,
  Chip,
  CircularProgress,
  Dropdown,
  IconButton,
  Input,
  Menu,
  MenuButton,
  MenuItem,
  MenuList,
  styled,
  TextField,
} from '@mui/joy'
import { ClickAwayListener } from '@mui/base/ClickAwayListener'
import React, { useRef, useState } from 'react'
import { useUser } from '../../hooks/useUser'
import Typography from '@mui/joy/Typography'
import { models } from '../../../wailsjs/go/models'
import User = models.User
import ListItemContent from '@mui/joy/ListItemContent'
import UserAvatar from '../UserAvatar'
import ListItem from '@mui/joy/ListItem'
import Box from '@mui/joy/Box'
import Stack from '@mui/joy/Stack'
import Close from '@mui/icons-material/Close'
import { useDebounce } from 'use-debounce'
import { Popper } from '@mui/base/Popper'

const Popup = styled(Popper)({
  zIndex: 1000,
})

type UserAutocompleteProps = {
  onUserClick: (user: User) => void
}

export default function UserAutocomplete(props: UserAutocompleteProps) {
  const [inputValue, setInputValue] = React.useState('')
  const [menuOpen, setMenuOpen] = useState(false)
  const [debouncedSearch] = useDebounce(inputValue, 500)
  const inputRef = useRef<HTMLDivElement>(null)
  const { data: user, isFetching, error } = useUser(debouncedSearch.length === 36 ? debouncedSearch : undefined)

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = event.target.value.trim()
    setInputValue(newValue)
    setMenuOpen(newValue.length > 0)
  }

  const handleItemClick = () => {
    setInputValue("")
    setMenuOpen(false)
    props.onUserClick(user!)
  }

  const renderMenuItems = () => {
    if (error) {
      return (
        <MenuItem disabled sx={{ justifyContent: 'center' }}>
          <Typography level="body-sm" color="warning" sx={{ p: 1 }}>
            No user found with the given ID
          </Typography>
        </MenuItem>
      )
    }

    if (user) {
      return (
        <MenuItem onClick={handleItemClick} sx={{ gap: 1, alignItems: 'center' }}>
          <UserAvatar id={user.ID}/>
          <Typography level="body-sm">
            {user.Username}
          </Typography>
        </MenuItem>
      )
    }

    if (inputValue.length === 36) {
      return (
        <MenuItem disabled sx={{ justifyContent: 'center' }}>
          <Stack direction='row' alignItems='center' sx={{ p: 1 }}>
            <CircularProgress size="sm" sx={{ mr: 1 }} />
            <Typography level="body-sm">Searching for user...</Typography>
          </Stack>
        </MenuItem>
      )
    }

    return (
      <MenuItem disabled sx={{ justifyContent: 'center' }}>
        <Typography level="body-sm" sx={{ p: 1 }}>
          Use 36 character ID to search for a user
        </Typography>
      </MenuItem>
    )
  }

  return (
    <Stack>
      <Input
        ref={inputRef}
        placeholder="Search user by ID..."
        value={inputValue}
        onChange={handleInputChange}
        slotProps={{
          input: {autoComplete: 'new-password'}
        }}
      />
      <Popup
        role={undefined}
        id="user-menu"
        open={menuOpen}
        anchorEl={inputRef.current}
        placement={'bottom-start'}
        disablePortal
        sx={{width: 1}}
      >
        <ClickAwayListener
          onClickAway={(event) => {
            if (event.target !== inputRef.current) {
              setMenuOpen(false)
            }
          }}
        >
          <MenuList variant="outlined" sx={{ boxShadow: 'md', width: 1 }}>
            {renderMenuItems()}
          </MenuList>
        </ClickAwayListener>
      </Popup>
    </Stack>
  )
}
