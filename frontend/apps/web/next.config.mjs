/** @type {import('next').NextConfig} */
const nextConfig = {
  transpilePackages: ["@workspace/ui"],
  
  // Allow multi-tenant subdomains in development
  experimental: {
    allowedDevOrigins: [
      'http://localhost',
      'http://localhost:3000',
      'http://*.app.local',
      'http://*.app.local:3000',
    ],
  },
}

export default nextConfig
