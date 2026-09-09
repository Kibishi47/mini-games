// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: false },

  modules: [
    '@nuxtjs/tailwindcss',
    '@pinia/nuxt',
  ],

  css: ['~/assets/css/main.css'],

  runtimeConfig: {
    public: {
      apiUrl: process.env.NUXT_PUBLIC_API_URL || 'http://localhost:8080',
      wsUrl: process.env.NUXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws',
    },
  },

  app: {
    head: {
      title: 'MiniGames | Festival de Jeux de Société & Wordle Multijoueur',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Plateforme multijoueur festive et pop-moderniste de jeux de société et Wordle compétitif en temps réel.' },
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Archivo+Black&family=Inter:wght@400;500;600;700;800&family=Space+Grotesk:wght@600;700;800;900&display=swap' },
      ],
    },
  },

  typescript: {
    strict: true,
    typeCheck: false,
  },
})
