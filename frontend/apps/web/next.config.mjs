/**
 * @type {import('next').NextConfig}
 */

function getBackendOrigin() {
  const apiBase = process.env.SERVER_API_URL
  if (!apiBase) throw new Error('SERVER_API_URL is not configured')
  return new URL(apiBase).origin
}

const nextConfig = {
  transpilePackages: ['@workspace/ui', '@workspace/services', '@workspace/shared-http', '@workspace/notix'],

  async rewrites() {
    const backend = getBackendOrigin()
    return [
      {
        source: '/api/v1/:path*',
        destination: `${backend}/api/v1/:path*`,
      },
    ]
  },

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
