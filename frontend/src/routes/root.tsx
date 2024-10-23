import { useAuth0 } from "@auth0/auth0-react";
import { Outlet } from "react-router-dom";
import DebugUtilities from "../components/debug-utilities";

export default function Root() {
  const { isLoading, isAuthenticated, loginWithRedirect, logout } = useAuth0();

  return (
    <>
      <header className="flex items-center justify-between bg-blue-950 p-4">
        <div className="font-sans text-lg font-bold text-gray-200">
          Alpine Rescue Team
        </div>
        <div className="space-x-4">
          <button className="rounded-sm bg-gray-200 px-4 py-2 font-bold uppercase text-gray-800 shadow hover:bg-gray-100">
            Send Page
          </button>
          <button
            disabled={isLoading}
            className="rounded-sm bg-gray-200 px-4 py-2 font-bold uppercase text-gray-800 shadow hover:bg-gray-100"
            onClick={() => {
              if (isAuthenticated) {
                logout();
              } else {
                loginWithRedirect();
              }
            }}
          >
            {isAuthenticated ? "Logout" : "Login"}
          </button>
        </div>
      </header>

      <section className="p-4">
        {isLoading ? (
          <></>
        ) : (
          <div className="space-y-4">
            {import.meta.env.DEV && <DebugUtilities />}
            <Outlet />
          </div>
        )}
      </section>
    </>
  );
}
