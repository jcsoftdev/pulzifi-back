import { withPayload } from '@payloadcms/next/withPayload'

const apiPort = process.env.HTTP_PORT ?? '3000'
const localstackPort = process.env.LOCALSTACK_PORT ?? '4566'

/** @type {import('next').NextConfig} */
const nextConfig = {
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

  allowedDevOrigins: [
    'localhost',
    `localhost:${apiPort}`,
    '*.localhost',
    `*.localhost:${apiPort}`,
    'app.local',
    `app.local:${apiPort}`,
    '*.app.local',
    `*.app.local:${apiPort}`,
    '*.pulzifi.local',
    `*.pulzifi.local:${apiPort}`,
    '*.local',
    `*.local:${apiPort}`,
    'pulzifi.com',
    '*.pulzifi.com',
  ],
}

export default withPayload(nextConfig)
