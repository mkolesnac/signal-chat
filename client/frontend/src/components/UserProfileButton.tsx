import { Dropdown, IconButton, Menu, MenuButton, MenuItem } from '@mui/joy'
import UserAvatar from './UserAvatar'
import { useAuth } from '../contexts/AuthContext'
import { SignOut } from '../../wailsjs/go/main/Auth'
import { AvatarProps } from '@mui/joy/Avatar'
import { useNavigate } from 'react-router-dom'
import Typography from '@mui/joy/Typography'
import Stack from '@mui/joy/Stack'

type UserProfileButtonProps = AvatarProps & {

}

export default function UserProfileButton(props: UserProfileButtonProps) {
  const {user: me} = useAuth()
  const navigate = useNavigate();

  const handleSignOut = async () => {
    await SignOut()
    navigate(`/signin`)
  }

  return (
    <Dropdown>
      <MenuButton
        slots={{ root: IconButton }}
        slotProps={{ root: { size: 'lg' } }}
        sx={{ p: 0.5 }}
      >
        <Stack direction='row' spacing={1} alignItems='center'>
          <UserAvatar id={me?.ID!} size='sm'/>
          <Typography level="body-md">{me?.Username}</Typography>
        </Stack>
      </MenuButton>
      <Menu>
        <MenuItem onClick={handleSignOut}>Sign out</MenuItem>
      </Menu>
    </Dropdown>
  );
}