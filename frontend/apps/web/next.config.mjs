import { withPayload } from '@payloadcms/next/withPayload'

/** @type {import('next').NextConfig} */
const nextConfig = {
  transpilePackages: ['@workspace/ui', '@workspace/services', '@workspace/shared-http', '@workspace/notix'],

  images: {
    unoptimized: process.env.NODE_ENV === 'development',
    remotePatterns: [
      // Local MinIO (dev)
      { protocol: 'http', hostname: 'localhost', port: '4566', pathname: '/**' },
      // Payload local media API (dev)
      { protocol: 'http', hostname: 'localhost', port: '3000', pathname: '/api/media/**' },
      // Production pulzifi.com assets
      { protocol: 'https', hostname: 'pulzifi.com', pathname: '/**' },
      { protocol: 'https', hostname: '*.pulzifi.com', pathname: '/**' },
    ],
  },

  experimental: {
    turbopackServerFastRefresh: true,
  },

  allowedDevOrigins: [
    'localhost',
    'localhost:3000',
    '*.localhost',
    '*.localhost:3000',
    'app.local',
    'app.local:3000',
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

export default withPayload(nextConfig)
