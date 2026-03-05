import { useRouteError, isRouteErrorResponse } from "react-router-dom";
import Error404Page from "./Error404Page";
import Error500Page from "./Error500Page";
import GenericErrorPage from "./GenericErrorPage";

const ErrorBoundary = () => {
    const error = useRouteError();

    if (isRouteErrorResponse(error)) {
        if (error.status === 404) {
            return <Error404Page />;
        }

        if (error.status >= 500 && error.status < 600) {
            return <Error500Page />;
        }
    }

    // Generic error fallback for other response errors or JS errors
    console.error("Application Error:", error);
    return <GenericErrorPage error={error} />;
};

export default ErrorBoundary;
