import { useAuth0 } from "@auth0/auth0-react";

export default function Login() {
  const { loginWithRedirect } = useAuth0();

  return (
    <div>
      <p>Please login</p>
      <button onClick={() => loginWithRedirect()}>Login</button>
    </div>
  );
}
