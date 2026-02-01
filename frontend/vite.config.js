import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";

// https://vite.dev/config/
export default defineConfig({
  server: {
    host: true,
    port: 5173,
    open: true,
    allowedHosts: true,
  },
  plugins: [tailwindcss(), vue()],
});
