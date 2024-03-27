import React from "react";
import ReactDOM from "react-dom/client";
import { Auth0Provider, useAuth0 } from "@auth0/auth0-react";
import { RouterProvider, createBrowserRouter } from "react-router-dom";
import ErrorPage from "./error-page";
import Root from "./routes/root";
import Login from "./routes/login";
import AuthCallback from "./routes/auth-callback";
import Home from "./routes/home";
import "./index.css";

function App() {
  const {} = useAuth0();
  return (
    <RouterProvider
      router={createBrowserRouter([
        {
          path: "/",
          element: <Root />,
          errorElement: <ErrorPage />,
          children: [
            {
              index: true,
              element: <Home />,
            },
          ],
        },
        { path: "callback", element: <AuthCallback /> },
        { path: "login", element: <Login /> },
      ])}
    />
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Auth0Provider
      domain={import.meta.env.VITE_AUTH0_DOMAIN}
      clientId={import.meta.env.VITE_AUTH0_CLIENT_ID}
      authorizationParams={{
        redirect_uri: window.location.origin,
      }}
    >
      <App />
    </Auth0Provider>
  </React.StrictMode>
);
