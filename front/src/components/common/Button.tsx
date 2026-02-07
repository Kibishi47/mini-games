import React from "react";
import { Link } from "react-router-dom";
import "@/styles/components/button.css";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: "primary" | "outline";
    to?: string;
    className?: string;
}

const Button: React.FC<ButtonProps> = ({
    children,
    variant = "primary",
    to,
    className = "",
    ...props
}) => {
    const baseClass = `btn btn-${variant} ${className}`;

    if (to) {
        return (
            <Link to={to} className={baseClass}>
                {children}
            </Link>
        );
    }

    return (
        <button className={baseClass} {...props}>
            {children}
        </button>
    );
};

export default Button;
