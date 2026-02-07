import Button from "../components/common/Button";

const NotFoundPage = () => {
    return (
        <div
            style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                justifyContent: "center",
                minHeight: "100vh",
                textAlign: "center",
                padding: "2rem",
            }}
        >
            <h1
                className="text-gradient"
                style={{
                    fontSize: "8rem",
                    margin: 0,
                    fontWeight: 800,
                    lineHeight: 1,
                }}
            >
                404
            </h1>
            <h2
                style={{
                    fontSize: "2rem",
                    margin: "1rem 0 2rem",
                    color: "var(--color-text)",
                }}
            >
                Page Introuvable
            </h2>
            <p
                style={{
                    maxWidth: "500px",
                    color: "rgba(255, 255, 255, 0.6)",
                    marginBottom: "3rem",
                    fontSize: "1.1rem",
                }}
            >
                Oups ! Il semblerait que vous vous soyez perdu dans le cyberespace. Cette
                page n'existe pas ou a été déplacée.
            </p>
            <Button to="/" variant="primary">
                Retour à l'accueil
            </Button>
        </div>
    );
};

export default NotFoundPage;
