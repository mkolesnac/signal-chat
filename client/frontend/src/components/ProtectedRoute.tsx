import React, { useEffect } from 'react'
import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from '../contexts/AuthContext'

const ProtectedRoute = () => {
  const { user } = useAuth();

  return user ? <Outlet /> : <Navigate to="/signin" replace />;
};

export default ProtectedRoute;