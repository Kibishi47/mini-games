import { NavLink } from "react-router-dom";
import Button from "./Button";
import { useAuth } from "../../auth/AuthContext";
import "@/styles/components/header.css";

const Header = () => {
  const { user, isLoading } = useAuth();

  return (
    <header className="header">
      <nav className="nav">
        {/* Logo */}
        <NavLink to="/" className="logo">
          Mini<span>Games</span>
        </NavLink>

        {/* Actions */}
        <div className="nav-actions">
          <Button to="/play" variant="primary">
            Jouer
          </Button>

          {isLoading ? (
            <div style={{ minWidth: "140px" }} />
          ) : user ? (
            <div style={{ display: "flex", alignItems: "center", gap: "1rem" }}>
              <Button to="/profile" variant="outline">
                {user.displayName}
              </Button>
            </div>
          ) : (
            <Button to="/login" variant="outline">
              Se connecter
            </Button>
          )}
        </div>
      </nav>
    </header>
  );
};

export default Header;