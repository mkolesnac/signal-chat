import * as React from 'react'
import Avatar, { AvatarProps } from '@mui/joy/Avatar'

type UserAvatarProps = AvatarProps & {
  username: string;
};

export default function UserAvatar(props: UserAvatarProps) {
  const { username, ...other } = props;

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(part => part[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  // Function to generate consistent color based on username
  const generateColor = (str: string) => {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }

    // Define a set of preset colors
    const colors = [
      '#2196F3', // blue
      '#4CAF50', // green
      '#FFC107', // yellow
      '#F44336', // red
      '#9C27B0', // purple
      '#E91E63', // pink
      '#3F51B5', // indigo
      '#009688'  // teal
    ];

    // Use the hash to select a color
    const colorIndex = Math.abs(hash) % colors.length;
    return colors[colorIndex];
  };

  const initials = getInitials(username);
  const backgroundColor = generateColor(username);

  return (
    <Avatar
      size='sm'
      sx={{
        backgroundColor,
        color: 'white',
      }}
      {...other}
    >
      {initials}
    </Avatar>
  );
}