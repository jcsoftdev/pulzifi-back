/**
 * @type {import('next').NextConfig}
 */

function getBackendOrigin() {
  const apiBase = process.env.SERVER_API_URL
  if (!apiBase) {
    // Warn at build time — the rewrite will be a no-op.
    // In production, SERVER_API_URL must be set as a build variable in Railway.
    console.warn('[next.config] WARNING: SERVER_API_URL is not set — /api/v1/* rewrite will not be configured')
    return null
  }
  return new URL(apiBase).origin
}

const nextConfig = {
  transpilePackages: ['@workspace/ui', '@workspace/services', '@workspace/shared-http', '@workspace/notix'],

  async rewrites() {
    const backend = getBackendOrigin()
    if (!backend) return []
    console.log(`[next.config] Rewrite /api/v1/* → ${backend}/api/v1/*`)
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
