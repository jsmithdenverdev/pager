import { useAuth0 } from "@auth0/auth0-react";

export default function Login() {
  const { loginWithRedirect } = useAuth0();

  return (
    <div className="flex items-center justify-center h-screen">
      <div className="flex flex-col align-center space-y-4">
        <h1 className="text-3xl">Login to Pager</h1>
        <button
          className="rounded-sm bg-blue-800 px-4 py-2 font-bold uppercase text-gray-200 shadow hover:bg-gray-300"
          onClick={() => loginWithRedirect()}
        >
          Login
        </button>
      </div>
    </div>
  );
}
