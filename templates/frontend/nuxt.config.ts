// https://nuxt.com/docs/api/configuration/nuxt-config
const isProduction = process.env.NODE_ENV === 'production';

export default defineNuxtConfig({
  runtimeConfig: {
    public: {
      apiBase: '/api',
    },
  },
  devtools: { enabled: !isProduction },
  sourcemap: {
    client: !isProduction,
    server: !isProduction,
  },
  nitro: {
    compressPublicAssets: true,
  },
  app: {
    head: {
      htmlAttrs: { lang: 'en' },
      title: '{{UI_PACKAGE}}',
      meta: [
        { name: 'application-name', content: '{{UI_PACKAGE}}' },
        { name: 'color-scheme', content: 'light dark' },
      ],
    },
  },
  modules: ['@nuxt/ui', '@nuxt/eslint'],
  css: ['~/assets/css/main.css'],
  compatibilityDate: '{{NUXT_COMPATIBILITY_DATE}}',
});
