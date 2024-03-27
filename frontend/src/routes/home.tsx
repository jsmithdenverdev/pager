import { useAuth0 } from "@auth0/auth0-react";
import { useEffect, useState } from "react";

export default function Home() {
  const { user, getAccessTokenSilently } = useAuth0();
  const [token, setToken] = useState<string>("");
  useEffect(() => {
    getAccessTokenSilently().then((token) => {
      setToken(token);
    });
  }, [getAccessTokenSilently]);

  return (
    <div>
      <h1 className="text-lg">Home</h1>
      <p>Signed in as {user?.name}</p>
      <div>
        <h3>Token</h3>
        <pre>{token}</pre>
      </div>
    </div>
  );
}
