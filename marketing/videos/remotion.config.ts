import { Config } from "@remotion/cli/config";
import path from "path";

Config.setConcurrency(4);
Config.setCodec("h264");
Config.setOutputLocation("out");
Config.setOverwriteOutput(true);

// Resolve @/* path alias in Remotion's webpack bundler
Config.overrideWebpackConfig((config) => ({
  ...config,
  resolve: {
    ...config.resolve,
    alias: {
      ...(config.resolve?.alias ?? {}),
      "@": path.join(__dirname, "src"),
    },
  },
}));
