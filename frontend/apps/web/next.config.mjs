/**
 * @type {import('next').NextConfig}
 */

function getBackendOrigin() {
  const apiBase =
    process.env.SERVER_API_URL ??
    process.env.API_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    'http://localhost:9090'
  try {
    return new URL(apiBase).origin
  } catch {
    return 'http://localhost:9090'
  }
}

const nextConfig = {
  transpilePackages: ["@workspace/ui"],

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
  ],
}

export default nextConfig
