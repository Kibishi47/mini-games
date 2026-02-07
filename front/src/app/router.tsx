import { createBrowserRouter } from "react-router-dom";
import RootLayout from "../layouts/RootLayout";

import HomePage from "../pages/HomePage";
import NotFoundPage from "../pages/NotFoundPage";

import PlayPage from "../pages/PlayPage";
import LoginPage from "../pages/LoginPage";
import RegisterPage from "../pages/RegisterPage";

export const router = createBrowserRouter([
    {
        element: <RootLayout />,
        children: [
            { path: "/", element: <HomePage /> },
            { path: "/play", element: <PlayPage /> },
            { path: "/login", element: <LoginPage /> },
            { path: "/register", element: <RegisterPage /> },
        ],
    },
    {
        path: "*",
        element: <NotFoundPage />,
    },
]);
