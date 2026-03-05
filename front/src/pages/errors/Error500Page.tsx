import Button from "../../components/common/Button";

const Error500Page = () => {
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
                500
            </h1>
            <h2
                style={{
                    fontSize: "2rem",
                    margin: "1rem 0 2rem",
                    color: "var(--color-text)",
                }}
            >
                Erreur Interne du Serveur
            </h2>
            <p
                style={{
                    maxWidth: "500px",
                    color: "rgba(255, 255, 255, 0.6)",
                    marginBottom: "3rem",
                    fontSize: "1.1rem",
                }}
            >
                Désolé, nos serveurs rencontrent une difficulté passagère. Nos techniciens
                ont été alertés et travaillent sur le problème.
            </p>
            <Button onClick={() => window.location.reload()} variant="primary">
                Actualiser la page
            </Button>
            <div style={{ marginTop: "1rem" }}>
                <Button to="/" variant="outline">
                    Retour à l'accueil
                </Button>
            </div>
        </div>
    );
};

export default Error500Page;
