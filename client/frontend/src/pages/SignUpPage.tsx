import * as React from 'react'
import Box from '@mui/joy/Box'
import Button from '@mui/joy/Button'
import FormControl from '@mui/joy/FormControl'
import FormLabel from '@mui/joy/FormLabel'
import IconButton from '@mui/joy/IconButton'
import Link from '@mui/joy/Link'
import Input from '@mui/joy/Input'
import Typography from '@mui/joy/Typography'
import Stack from '@mui/joy/Stack'
import BadgeRoundedIcon from '@mui/icons-material/BadgeRounded'
import ColorSchemeToggle from '../components/ColorSchemeToggle'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { SignUp } from '../../wailsjs/go/main/Auth'
import { Alert } from '@mui/joy'
import ReportIcon from '@mui/icons-material/Report'
import { useAuth } from '../contexts/AuthContext'

interface FormElements extends HTMLFormControlsCollection {
  username: HTMLInputElement
  password: HTMLInputElement
}
interface SignUpFormElement extends HTMLFormElement {
  readonly elements: FormElements
}

export default function SignUpPage() {
  const {setUser} = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (event: React.FormEvent<SignUpFormElement>) => {
    event.preventDefault()

    const formElements = event.currentTarget.elements

    try {
      const user = await SignUp(formElements.username.value, formElements.password.value);
      console.log('usere: %o', user)
      setUser(user);
      navigate('/', { replace: true });
    }  catch (error) {
      setError(String(error))
    }
  }

  // @ts-ignore
  return (
    <Box>
      <Box
        sx={(theme) => ({
          backgroundColor: 'rgba(255 255 255 / 0.2)',
          [theme.getColorSchemeSelector('dark')]: {
            backgroundColor: 'rgba(19 19 24 / 0.4)',
          },
          display: 'flex',
          flexDirection: 'column',
          minHeight: '100dvh',
          width: '100%',
          px: 2,
        })}
      >
        <Box
          component="header"
          sx={{ py: 3, display: 'flex', justifyContent: 'space-between' }}
        >
          <Box sx={{ gap: 2, display: 'flex', alignItems: 'center' }}>
            <IconButton variant="soft" color="primary" size="sm">
              <BadgeRoundedIcon />
            </IconButton>
            <Typography level="title-lg">Company logo</Typography>
          </Box>
          <ColorSchemeToggle />
        </Box>
        <Box
          component="main"
          sx={{
            my: 'auto',
            py: 2,
            pb: 5,
            display: 'flex',
            flexDirection: 'column',
            gap: 2,
            width: 400,
            maxWidth: '100%',
            mx: 'auto',
            borderRadius: 'sm',
            '& form': {
              display: 'flex',
              flexDirection: 'column',
              gap: 2,
            },
            [`& .MuiFormLabel-asterisk`]: {
              visibility: 'hidden',
            },
          }}
        >
          <Stack sx={{ gap: 1 }}>
            <Typography component="h1" level="h3">
              Sign up
            </Typography>
            <Typography level="body-sm">
              Already have an account?{' '}
              <Link component={RouterLink} to="/signin" level="title-sm">
                Sign in!
              </Link>
            </Typography>
          </Stack>
          <Stack sx={{ gap: 4, mt: 2 }}>
            {error && (
              <Alert color="danger" variant="soft" startDecorator={<ReportIcon />}>
                {error}
              </Alert>
            )}
            <form
              onSubmit={handleSubmit}
            >
              <FormControl required>
                <FormLabel>Email</FormLabel>
                <Input type="email" name="email" />
              </FormControl>
              <FormControl required>
                <FormLabel>Password</FormLabel>
                <Input type="password" name="password" />
              </FormControl>
              <Button type="submit" fullWidth sx={{ mt: 2 }}>
                Sign up
              </Button>
            </form>
          </Stack>
        </Box>
      </Box>
    </Box>
  )
}
