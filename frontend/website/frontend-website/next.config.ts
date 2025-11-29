import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  
  // ВАЖНО: Игнорируем ESLint и TypeScript ошибки при сборке
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
  
  // Отключаем статическую оптимизацию для страниц с динамическими параметрами
  experimental: {
    
  },
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '8080',
        pathname: '/uploads/**',
      },
      {
        protocol: 'https',
        hostname: 'cdn-icons-png.flaticon.com',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'kinopoiskapiunofficial.tech',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: '**.selstorage.ru',
        port: '',
        pathname: '/**',
      }
    ],
  },
  reactStrictMode: false,
};

export default nextConfig;