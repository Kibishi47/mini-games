import { redirect } from "react-router-dom";
import { authApi } from "./auth.api";

export const requireAuthLoader = async () => {
    try {
        await authApi.refresh();
        return null;
    } catch {
        return redirect("/login");
    }
};

export const requireGuestLoader = async () => {
    try {
        await authApi.refresh();
        return redirect("/account");
    } catch {
        return null;
    }
};
