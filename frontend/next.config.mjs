/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Security headers applied to every response. The CSP stays strictly 'self'
  // because next/font self-hosts the fonts at build time (no external hosts),
  // and the BFF proxies all API traffic — the browser never calls the Go API
  // cross-origin, so no connect-src exceptions are needed.
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          {
            key: 'Permissions-Policy',
            value: 'camera=(), microphone=(), geolocation=()',
          },
        ],
      },
    ];
  },
};

export default nextConfig;
