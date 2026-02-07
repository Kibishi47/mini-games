import { NavLink } from "react-router-dom";
import Button from "./Button";
import "@/styles/components/header.css";

const Header = () => {
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

          <Button to="/login" variant="outline">
            Se connecter
          </Button>
        </div>
      </nav>
    </header>
  );
};

export default Header;