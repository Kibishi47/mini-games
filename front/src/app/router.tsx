import { createBrowserRouter } from "react-router-dom";
import RootLayout from "../layouts/RootLayout";
import AuthLayout from "../layouts/AuthLayout";

import HomePage from "../pages/HomePage";
import NotFoundPage from "../pages/NotFoundPage";

export const router = createBrowserRouter([
    {
        element: <RootLayout />,
        children: [{ path: "/", element: <HomePage /> }],
    },
    {
        element: <AuthLayout />,
        children: [],
    },
    {
        path: "*",
        element: <NotFoundPage />,
    },
]);
