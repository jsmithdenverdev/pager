import { useAuth0 } from "@auth0/auth0-react";
import { useEffect, useState } from "react";

export default function Home() {
  const [token, setToken] = useState<string>("");
  const { user, getAccessTokenSilently } = useAuth0();

  useEffect(() => {
    getAccessTokenSilently().then((token) => {
      setToken(token);
    });
  }, [getAccessTokenSilently]);

  return (
    <div>
      <h1 className="text-lg">Home · {user?.name}</h1>
      <div>
        <button
          className="bg-blue-500 text-white rounded p-3"
          onClick={() => navigator.clipboard.writeText(token)}
        >
          Copy token to clipboard
        </button>
      </div>
    </div>
  );
}
