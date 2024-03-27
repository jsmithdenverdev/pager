import { useRouteError } from "react-router-dom";

export default function ErrorPage() {
  const error = useRouteError() as { statusText?: string; message?: string };
  console.log(error);

  return (
    <div>
      <h1>Oops!</h1>
      <p>Sorry, an unexpected error has occured.</p>
      <p>
        <i>{error.statusText || error.message}</i>
      </p>
    </div>
  );
}
