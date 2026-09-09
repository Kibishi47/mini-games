import type { Config } from 'tailwindcss'

export default {
  content: [
    './components/**/*.{js,vue,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './plugins/**/*.{js,ts}',
    './app.vue',
    './error.vue',
  ],
  theme: {
    extend: {
      colors: {
        'board-cream': '#FAF5E8', // Fond de page principal (papier chaud mat)
        'ink-black': '#121212',   // Tous les traits, contours, textes et ombres
        'game-yellow': '#FFD300', // Jaune canari : tuiles mal placées, actions primaires, alertes
        'game-blue': '#1D4ED8',   // Bleu roi : accents structurels, rôle Master, boutons secondaires
        'game-red': '#E63228',    // Rouge vermillon : actions destructrices (kick), chronos < 10s
        'game-green': '#158A44',  // Vert émeraude : tuiles correctes, succès, validations
        'game-pink': '#F27A9B',   // Rose bonbon : badges de modération, accents secondaires
        'board-white': '#FFFFFF', // Cartes de contenu, fond des tuiles neutres, faces de dés
      },
      fontFamily: {
        display: ['"Space Grotesk"', 'sans-serif'], // Titres percutants, majuscules massives
        condensed: ['"Archivo Black"', 'sans-serif'], // Chronomètres, scores, numéros de tours
        body: ['"Inter"', '"Plus Jakarta Sans"', 'sans-serif'], // Chat, textes courants, options
      },
      borderWidth: {
        '3': '3px',
        '4': '4px',
      },
      boxShadow: {
        'pop-xs': '2px 2px 0px #121212',
        'pop-sm': '3px 3px 0px #121212',
        'pop-md': '4px 4px 0px #121212',
        'pop-lg': '6px 6px 0px #121212',
      },
    },
  },
  plugins: [],
} satisfies Config
