import { createBrowserRouter } from "react-router-dom";
import RootLayout from "../layouts/RootLayout";
import { requireAuthLoader, requireGuestLoader } from "@/auth/requireAuth";

import HomePage from "../pages/HomePage";
import NotFoundPage from "../pages/NotFoundPage";

import PlayPage from "../pages/PlayPage";
import LoginPage from "../pages/LoginPage";
import RegisterPage from "../pages/RegisterPage";
import ProfilePage from "../pages/ProfilePage";
import TestPage from "../pages/TestPage";

export const router = createBrowserRouter([
    {
        element: <RootLayout />,
        children: [
            { path: "/", element: <HomePage /> },
            {
                path: "/play",
                element: <PlayPage />,
                loader: requireAuthLoader,
            },
            {
                path: "/profile",
                element: <ProfilePage />,
                loader: requireAuthLoader,
            },
            {
                path: "/test",
                element: <TestPage />,
            },
            {
                path: "/login",
                element: <LoginPage />,
                loader: requireGuestLoader,
            },
            {
                path: "/register",
                element: <RegisterPage />,
                loader: requireGuestLoader,
            },
        ],
    },
    {
        path: "*",
        element: <NotFoundPage />,
    },
]);
