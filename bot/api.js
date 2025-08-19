const axios = require('axios');
const config = require('./config');

const client = axios.create({
  baseURL: config.BACKEND_API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': config.TELEGRAM_BOT_API_KEY
  }
});

/**
 * Подписка: посылаем { Us, TelegramID }
 * Возвращает { ok: true } или выбрасывает ошибку
 */
async function subscribeUser({ Us, TelegramID }) {
  const url = '/api/bot/subscriptions'; // согласно роуту в Go
  const body = { Us, telegram_id: String(TelegramID) };
  const resp = await client.post(url, body);
  return resp.data;
}

async function unsubscribeByTelegramID({ TelegramID }) {
  const url = '/api/bot/unsubscriptions';
  const body = { telegram_id: String(TelegramID) };
  const resp = await client.post(url, body);
  return resp.data;
}

module.exports = {
  subscribeUser,
  unsubscribeByTelegramID
};
