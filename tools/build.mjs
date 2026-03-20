import { mkdir, copyFile } from "node:fs/promises";
import { build } from "esbuild";

await mkdir("web/static/dist", { recursive: true });

await build({
  bundle: true,
  entryPoints: ["web/static/src/app.js"],
  format: "iife",
  outfile: "web/static/dist/app.js",
  sourcemap: false,
  minify: false,
  target: "es2022",
});

await copyFile("web/static/src/app.css", "web/static/dist/app.css");
