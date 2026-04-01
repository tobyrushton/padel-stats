// @ts-check
import { defineConfig } from 'astro/config';

import react from '@astrojs/react';

import tailwindcss from '@tailwindcss/vite';

import aws from "astro-sst"

// https://astro.build/config
export default defineConfig({
  integrations: [react()],
  output: "server",
  adapter: aws(),

  vite: {
    plugins: [tailwindcss()]
  }
});