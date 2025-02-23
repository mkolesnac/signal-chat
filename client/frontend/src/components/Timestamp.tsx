import { TypographyProps } from '@mui/joy'
import React, { useEffect, useReducer, useState } from 'react'
import { format, formatDistanceToNowStrict, isThisYear } from 'date-fns'
import Typography from '@mui/joy/Typography'

type TimestampProps = TypographyProps & {
  value: number
}

function getTimeText(timestamp: number) {
  const date = new Date(timestamp)
  const now = new Date()

  // Less than 7 days - show relative time
  if (now.getTime() - date.getTime() < 7 * 24 * 60 * 60 * 1000) {
    const distance = formatDistanceToNowStrict(date, {
      addSuffix: true,
      roundingMethod: 'floor',
    })
    // Replace "minutes" with "m", "hours" with "h", etc.
    return distance
    // return distance
    //   .replace(' minutes', 'm')
    //   .replace(' minute', 'm')
    //   .replace(' hours', 'h')
    //   .replace(' hour', 'h')
    //   .replace(' days', 'd')
    //   .replace(' day', 'd')
  }

  // Older messages - show date
  if (isThisYear(date)) {
    return format(date, 'MMM d')
  }

  return format(date, 'MMM d, yyyy')
}

export default function Timestamp(props: TimestampProps) {
  const { value, ...rest } = props
  const [timeText, setTimeText] = useState<string | undefined>()

  console.log("Timestamp, value: %o, text: %o", value, timeText)

  useEffect(() => {
    const date = new Date(value)
    const now = new Date()
    const age = now.getTime() - date.getTime()

    setTimeText(getTimeText(value))
    console.log("age: %o", getTimeText(value))
    // Only set up auto-update for recent messages
    if (age < 24 * 60 * 60 * 1000) {
      // less than 24 hours old
      const timer = setInterval(() => {
        setTimeText(getTimeText(value))
      }, 60 * 1000) // update every minute
      return () => clearInterval(timer)
    }
  }, [value])

  return (
    <Typography level="body-xs" noWrap {...rest}>
      {timeText}
    </Typography>
  )
}
