import { useAuth0 } from "@auth0/auth0-react";
import { useState, useEffect } from "react";

export default function DebugUtilities() {
  const [token, setToken] = useState<string>("");
  const { user, getAccessTokenSilently } = useAuth0();

  const [tokenButtonText, setTokenButtonText] = useState("Copy token");

  const handleCopyToken = async () => {
    try {
      await navigator.clipboard.writeText(token);
      setTokenButtonText("Copied!");
      setTimeout(() => setTokenButtonText("Copy token"), 2000); // Reset the button text after 2 seconds
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
  };

  useEffect(() => {
    getAccessTokenSilently().then((token) => {
      setToken(token);
    });
  }, [getAccessTokenSilently]);

  return (
    <section className="flex flex-col border-dashed border-2 border-sky-500 p-4 ">
      <div>
        <h1 className="text-xl uppercase">
          <span className="uppercase">Debug</span>
        </h1>
      </div>
      <div>
        <details className="p-4  rounded-sm shadow">
          <summary className="text-sm font-semibold cursor-pointer">
            User Details - {user?.email}
          </summary>
          <pre>{JSON.stringify(user, null, 2)}</pre>
        </details>
      </div>
      <div>
        <button
          className="rounded-sm bg-gray-200 px-4 py-2 font-bold uppercase text-gray-800 shadow hover:bg-gray-300 ml-auto"
          onClick={handleCopyToken}
        >
          {tokenButtonText}
        </button>
      </div>
    </section>
  );
}
