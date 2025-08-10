const express = require('express');
const bodyParser = require('body-parser');
const config = require('./config');
const { bot, sendBroadcast } = require('./bot');

const app = express();
app.use(bodyParser.json({ limit: '5mb' }));

// простой middleware для проверки X-API-Key
function botAuthMiddleware(req, res, next) {
  const key = req.header('X-API-Key') || '';
  if (!key || key !== config.TELEGRAM_BOT_API_KEY) {
    return res.status(401).json({ error: 'unauthorized' });
  }
  next();
}

/**
 * POST /internal/broadcast
 * body: { items: [ { telegramIds: [..], imageUrl?, title?, text? }, ... ] }
 * заголовок X-API-Key: должен совпадать с TELEGRAM_BOT_API_KEY
 */
app.post('/internal/broadcast', botAuthMiddleware, async (req, res) => {
  try {
    const items = req.body.items;
    if (!Array.isArray(items)) return res.status(400).json({ error: 'items must be array' });

    // Запускаем рассылку асинхронно, но ждём завершения чтобы вернуть статус
    // (можно поменять на fire-and-forget при желании)
    await sendBroadcast(items);
    return res.json({ ok: true, delivered: true });
  } catch (err) {
    console.error('broadcast error', err);
    return res.status(500).json({ error: 'internal_error' });
  }
});

app.get('/health', (req, res) => res.send('ok'));

const port = config.PORT || 3000;
app.listen(port, () => {
  console.log('Server listening on', port);
});
