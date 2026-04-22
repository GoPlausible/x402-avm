import { defineConfig } from "tsup";

const baseConfig = {
  entry: {
    index: "src/index.ts",
  },
  dts: {
    resolve: true,
  },
  sourcemap: true,
  target: "node16" as const,
};

export default defineConfig([
  {
    ...baseConfig,
    format: "esm" as const,
    outDir: "dist/esm",
    clean: true,
  },
  {
    ...baseConfig,
    format: "cjs" as const,
    outDir: "dist/cjs",
    clean: false,
  },
]);
