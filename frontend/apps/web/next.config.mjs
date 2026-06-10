import { withPayload } from '@payloadcms/next/withPayload'

const apiPort = process.env.HTTP_PORT ?? '3000'
const localstackPort = process.env.LOCALSTACK_PORT ?? '4566'

/** @type {import('next').NextConfig} */
const nextConfig = {
  poweredByHeader: false,

  transpilePackages: ['@workspace/ui', '@workspace/services', '@workspace/shared-http', '@workspace/notix'],

  images: {
    unoptimized: process.env.NODE_ENV === 'development',
    remotePatterns: [
      // Local MinIO (dev)
      { protocol: 'http', hostname: 'localhost', port: localstackPort, pathname: '/**' },
      // Payload local media API (dev)
      { protocol: 'http', hostname: 'localhost', port: apiPort, pathname: '/api/media/**' },
      // Production pulzifi.com assets
      { protocol: 'https', hostname: 'pulzifi.com', pathname: '/**' },
      { protocol: 'https', hostname: '*.pulzifi.com', pathname: '/**' },
    ],
  },

  experimental: {
    turbopackServerFastRefresh: true,
  },

  async redirects() {
    return [
      // Slug normalized to lowercase kebab-case (2026-06-09); old URL may be indexed
      {
        source: '/blog/What-is-competitive-intelligence',
        destination: '/blog/what-is-competitive-intelligence',
        permanent: true,
      },
    ]
  },

  async headers() {
    const securityHeaders = [
      { key: 'X-Content-Type-Options', value: 'nosniff' },
      { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
      { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
    ]

    const privateRoutePrefixes = [
      '/dashboard/:path*',
      '/workspaces/:path*',
      '/settings/:path*',
      '/team/:path*',
      '/admin/:path*',
      '/onboarding/:path*',
    ]

    return [
      {
        source: '/:path*',
        headers: securityHeaders,
      },
      ...privateRoutePrefixes.map((source) => ({
        source,
        headers: [{ key: 'X-Robots-Tag', value: 'noindex, nofollow' }],
      })),
    ]
  },

  allowedDevOrigins: [
    'localhost',
    `localhost:${apiPort}`,
    '*.localhost',
    `*.localhost:${apiPort}`,
    'lvh.me',
    `lvh.me:${apiPort}`,
    '*.lvh.me',
    `*.lvh.me:${apiPort}`,
    'pulzifi.com',
    '*.pulzifi.com',
  ],
}

export default withPayload(nextConfig)
