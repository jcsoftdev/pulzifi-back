import { withPayload } from '@payloadcms/next/withPayload'

/** @type {import('next').NextConfig} */
const nextConfig = {
  transpilePackages: ['@workspace/ui', '@workspace/services', '@workspace/shared-http', '@workspace/notix'],

  images: {
    remotePatterns: [
      // Local MinIO (dev)
      { protocol: 'http', hostname: 'localhost', port: '4566', pathname: '/**' },
      // Payload local media API (dev)
      { protocol: 'http', hostname: 'localhost', port: '3000', pathname: '/api/media/**' },
      // Production MinIO/S3 domain — update when known
      // { protocol: 'https', hostname: 's3.pulzifi.com', pathname: '/**' },
    ],
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
