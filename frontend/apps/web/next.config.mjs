/**
 * @type {import('next').NextConfig}
 */
const nextConfig = {
  transpilePackages: ["@workspace/ui"],
  

  allowedDevOrigins: [
    'localhost',
    'localhost:3000',
    '*.app.local',
    '*.app.local:3000',
    '*.pulzifi.local',
    '*.local',
  ],
}

export default nextConfig
