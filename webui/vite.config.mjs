import path from "node:path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

function packageChunkName(id) {
  const normalized = id.split("\\").join("/");
  const marker = "/node_modules/";
  const index = normalized.lastIndexOf(marker);
  if (index === -1) {
    return undefined;
  }

  const packagePath = normalized.slice(index + marker.length);
  const segments = packagePath.split("/");
  if (segments[0]?.startsWith("@") && segments.length > 1) {
    return `${segments[0].slice(1)}-${segments[1]}`;
  }
  return segments[0];
}

export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  plugins: [react(), tailwindcss()],
  optimizeDeps: {
    exclude: ["react-resizable-panels"],
  },
  server: {
    proxy: {
      "/api": {
        target: process.env.VITE_API_URL || "http://localhost:3009",
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes("node_modules")) {
            return undefined;
          }

          const packageName = packageChunkName(id);
          if (!packageName) {
            return "vendor";
          }

          if (["react", "react-dom", "scheduler", "react-router", "react-router-dom"].includes(packageName) || packageName.startsWith("remix-run")) {
            return "vendor-react";
          }

          if (["react-dnd", "dnd-core", "react-dnd-html5-backend"].includes(packageName)) {
            return "vendor-dnd";
          }

          if (packageName === "framer-motion" || packageName === "motion-dom" || packageName === "motion-utils") {
            return "vendor-motion";
          }

          if (packageName.startsWith("radix-ui-")) {
            return "vendor-radix";
          }

          if (["react-hook-form", "hookform-resolvers", "zod"].includes(packageName)) {
            return "vendor-forms";
          }

          if (["react-markdown", "remark-parse", "remark-rehype", "micromark", "unified"].includes(packageName) || packageName.startsWith("micromark-") || packageName.startsWith("mdast-") || packageName.startsWith("hast-") || packageName.startsWith("unist-") || packageName.startsWith("vfile")) {
            return "vendor-markdown";
          }

          if (["leaflet", "react-leaflet", "react-leaflet-core"].includes(packageName)) {
            return "vendor-maps";
          }

          if (["axios", "zustand", "sonner"].includes(packageName)) {
            return "vendor-app";
          }

          return undefined;
        },
      },
    },
  },
});
