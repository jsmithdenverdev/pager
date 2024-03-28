import { useAuth0 } from "@auth0/auth0-react";
import { Navigate, Outlet } from "react-router-dom";

export default function AuthGuard() {
  const { isLoading, isAuthenticated } = useAuth0();

  if (isLoading) {
    return;
  }

  return isAuthenticated ? <Outlet /> : <Navigate to="/login" />;
}
