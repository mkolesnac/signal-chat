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
import { Alert, GlobalStyles } from '@mui/joy'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import {SignIn} from "../../wailsjs/go/main/Auth";
import { useEffect, useState } from 'react'
import ReportIcon from '@mui/icons-material/Report';
import { useAuth } from '../contexts/AuthContext'

interface FormElements extends HTMLFormControlsCollection {
  email: HTMLInputElement
  password: HTMLInputElement
}
interface SignInFormElement extends HTMLFormElement {
  readonly elements: FormElements
}

export default function SignInPage() {
  const {setUser} = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      const user = await SignIn("bob@gmail.com", "test1234");
      setUser(user);
      navigate('/', { replace: true });
    })()
  })

  const handleSubmit = async (event: React.FormEvent<SignInFormElement>) => {
    event.preventDefault()

    const formElements = event.currentTarget.elements

    try {
      const user = await SignIn(formElements.email.value, formElements.password.value);
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
          width: { xs: '100%', md: '50vw' },
          transition: 'width var(--Transition-duration)',
          transitionDelay: 'calc(var(--Transition-duration) + 0.1s)',
          position: 'relative',
          zIndex: 1,
          display: 'flex',
          justifyContent: 'flex-end',
          backdropFilter: 'blur(12px)',
          backgroundColor: 'rgba(255 255 255 / 0.2)',
          [theme.getColorSchemeSelector('dark')]: {
            backgroundColor: 'rgba(19 19 24 / 0.4)',
          },
        })}
      >
        <GlobalStyles
          styles={{
            ':root': {
              '--Form-maxWidth': '800px',
              '--Transition-duration': '0.4s', // set to `none` to disable transition
            },
          }}
        />
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            minHeight: '100dvh',
            width: '100%',
            px: 2,
          }}
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
                Sign in
              </Typography>
              <Typography level="body-sm">
                New to company?{' '}
                <Link component={RouterLink} to="/signup" level="title-sm">
                  Sign up!
                </Link>
              </Typography>
            </Stack>
            <Stack sx={{ gap: 4, mt: 2 }}>
              {error && (
                <Alert color="danger" variant="soft" startDecorator={<ReportIcon />}>
                  {error}
                </Alert>
              )}
              <form onSubmit={handleSubmit}>
                <FormControl required>
                  <FormLabel>Email</FormLabel>
                  <Input type="email" name="email" />
                </FormControl>
                <FormControl required>
                  <FormLabel>Password</FormLabel>
                  <Input type="password" name="password" />
                </FormControl>
                <Stack sx={{ gap: 4, mt: 2 }}>
                  <Box>
                    <Link level="title-sm" href="#replace-with-a-link">
                      Forgot your password?
                    </Link>
                  </Box>
                  <Button type="submit" fullWidth>
                    Sign in
                  </Button>
                </Stack>
              </form>
            </Stack>
          </Box>
        </Box>
      </Box>
      <Box
        sx={(theme) => ({
          height: '100%',
          position: 'fixed',
          right: 0,
          top: 0,
          bottom: 0,
          left: { xs: 0, md: '50vw' },
          transition:
            'background-image var(--Transition-duration), left var(--Transition-duration) !important',
          transitionDelay: 'calc(var(--Transition-duration) + 0.1s)',
          backgroundColor: 'background.level1',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          backgroundRepeat: 'no-repeat',
          backgroundImage:
            'url(https://images.unsplash.com/photo-1527181152855-fc03fc7949c8?auto=format&w=1000&dpr=2)',
          [theme.getColorSchemeSelector('dark')]: {
            backgroundImage:
              'url(https://images.unsplash.com/photo-1572072393749-3ca9c8ea0831?auto=format&w=1000&dpr=2)',
          },
        })}
      />
    </Box>
  )
}
