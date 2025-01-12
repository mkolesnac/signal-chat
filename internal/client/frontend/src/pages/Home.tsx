import * as React from 'react';
import { Navigate, Outlet } from 'react-router-dom'
import Sidebar from '../components/Sidebar'
import Header from '../components/Header'
import Box from '@mui/joy/Box'
import MyMessages from '../components/MyMessages';

export default function Home() {
  return (
    <Box sx={{ display: 'flex', minHeight: '100dvh' }}>
      <Sidebar />
      <Header />
      <Box component="main" className="MainContent" sx={{ flex: 1 }}>
        <MyMessages />
      </Box>
    </Box>
  );
}

