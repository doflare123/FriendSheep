require('dotenv').config();

const config = {
  TELEGRAM_TOKEN: process.env.TELEGRAM_TOKEN,
  BACKEND_API_BASE_URL: process.env.BACKEND_API_BASE_URL,
  TELEGRAM_BOT_API_KEY: process.env.TELEGRAM_BOT_API_KEY,
  BROADCAST_RATE_MS: process.env.BROADCAST_RATE_MS ? parseInt(process.env.BROADCAST_RATE_MS) : 100,
  PORT: process.env.PORT ? parseInt(process.env.PORT) : 3000,
};

if (!config.TELEGRAM_TOKEN) {
  console.error('TELEGRAM_TOKEN is required');
  process.exit(1);
}
if (!config.BACKEND_API_BASE_URL) {
  console.error('BACKEND_API_BASE_URL is required');
  process.exit(1);
}
if (!config.TELEGRAM_BOT_API_KEY) {
  console.error('TELEGRAM_BOT_API_KEY is required');
  process.exit(1);
}

module.exports = config;
