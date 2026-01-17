import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  trailingSlash: true,
  basePath: "/ui",
  assetPrefix: "/ui",
  images: {
    unoptimized: true,
  },
  // Disable server components for static export
  experimental: {},
};

export default nextConfig;
