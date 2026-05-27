/** @type {import('next').NextConfig} */

// Content-Security-Policy-г middleware.ts-ээс per-request nonce-тэй өгнө.
// Энд зөвхөн CSP-аас бусад статик security header-уудыг хариуцна.

const nextConfig = {
  reactStrictMode: true,
  // Standalone output → slim production Docker image (server.js + minimal
  // node_modules) instead of shipping the whole tree. See frontend/Dockerfile.
  output: 'standalone',
  // Security headers applied to every response.
  async headers() {
    const securityHeaders = [
      { key: 'X-Frame-Options', value: 'DENY' },
      { key: 'X-Content-Type-Options', value: 'nosniff' },
      { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
      {
        key: 'Permissions-Policy',
        value: 'camera=(), microphone=(), geolocation=()',
      },
      // Content-Security-Policy — middleware.ts дотор per-request nonce-тэй
      // тохируулагдана. Энд static header болгож тавихгүй (nonce-гүй болно).
    ];

    // HSTS-ийг зөвхөн production-д илгээнэ — dev дээр http тул HSTS тохиромжгүй.
    if (process.env.NODE_ENV === 'production') {
      securityHeaders.push({
        key: 'Strict-Transport-Security',
        value: 'max-age=63072000; includeSubDomains',
      });
    }

    return [
      {
        source: '/:path*',
        headers: securityHeaders,
      },
    ];
  },
};

export default nextConfig;
