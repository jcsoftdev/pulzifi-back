/**
 * @type {import('next').NextConfig}
 */

const nextConfig = {
  transpilePackages: ['@workspace/ui', '@workspace/services', '@workspace/shared-http', '@workspace/notix'],

  allowedDevOrigins: [
    'localhost',
    'localhost:3000',
    '*.localhost',
    '*.localhost:3000',
    '*.app.local',
    '*.app.local:3000',
    '*.pulzifi.local',
    '*.pulzifi.local:3000',
    '*.local',
    '*.local:3000',
    'pulzifi.com',
    '*.pulzifi.com',
  ],
}

export default nextConfig
