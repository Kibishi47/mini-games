import Button from "../components/common/Button";

const HomePage = () => {
    return (
        <div className="container" style={{ paddingTop: "80px" }}>
            {/* Hero Section */}
            <section style={{
                minHeight: "80vh",
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                alignItems: "flex-start",
                gap: "1.5rem"
            }}>
                <h1 style={{ fontSize: "5rem", maxWidth: "800px", margin: 0 }}>
                    Jouez aux meilleurs <br />
                    <span className="text-gradient">Mini-Jeux</span> en ligne.
                </h1>
                <p style={{
                    fontSize: "1.25rem",
                    color: "rgba(255,255,255,0.7)",
                    maxWidth: "600px",
                    lineHeight: "1.6"
                }}>
                    Plongez dans un univers arcade néon. Des classiques revisités aux nouveautés exclusives. Rejoignez la compétition maintenant.
                </p>
                <div style={{ display: "flex", gap: "1rem", marginTop: "1rem" }}>
                    <Button to="/play" variant="primary" style={{ padding: "1rem 2.5rem", fontSize: "1.1rem" }}>
                        Commencer à jouer
                    </Button>
                    <Button to="/about" variant="outline" style={{ padding: "1rem 2.5rem", fontSize: "1.1rem" }}>
                        En savoir plus
                    </Button>
                </div>
            </section>

            {/* Featured Games Placeholder */}
            <section style={{ padding: "4rem 0" }}>
                <h2 style={{ fontSize: "2.5rem", marginBottom: "2rem" }}>Jeux Populaires</h2>
                <div style={{
                    display: "grid",
                    gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
                    gap: "2rem"
                }}>
                    {[1, 2, 3].map((item) => (
                        <div key={item} style={{
                            background: "var(--color-secondary)",
                            borderRadius: "16px",
                            overflow: "hidden",
                            border: "1px solid rgba(255,255,255,0.05)",
                            transition: "transform 0.2s"
                        }}>
                            <div style={{
                                height: "200px",
                                background: `linear-gradient(45deg, #111, #222)`
                            }} />
                            <div style={{ padding: "1.5rem" }}>
                                <h3 style={{ margin: "0 0 0.5rem 0" }}>Jeu {item}</h3>
                                <p style={{ color: "#888", marginBottom: "1rem" }}>Description courte du jeu pour donner envie.</p>
                                <Button variant="outline" style={{ width: "100%" }}>Jouer maintenant</Button>
                            </div>
                        </div>
                    ))}
                </div>
            </section>
        </div>
    );
};

export default HomePage;
