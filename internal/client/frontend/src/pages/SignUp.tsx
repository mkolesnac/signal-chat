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
import { Link as RouterLink } from 'react-router-dom'

interface FormElements extends HTMLFormControlsCollection {
  email: HTMLInputElement
  password: HTMLInputElement
  persistent: HTMLInputElement
}
interface SignInFormElement extends HTMLFormElement {
  readonly elements: FormElements
}

export default function SignUp() {
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
            <form
              onSubmit={(event: React.FormEvent<SignInFormElement>) => {
                event.preventDefault()
                const formElements = event.currentTarget.elements
                const data = {
                  email: formElements.email.value,
                  password: formElements.password.value,
                  persistent: formElements.persistent.checked,
                }
                alert(JSON.stringify(data, null, 2))
              }}
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
