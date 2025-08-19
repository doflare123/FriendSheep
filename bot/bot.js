const TelegramBot = require('node-telegram-bot-api');
const config = require('./config');
const api = require('./api');
const pLimit = require('p-limit').default;

const bot = new TelegramBot(config.TELEGRAM_TOKEN, { polling: true });

// вспомогательная клавиатура
const mainKeyboard = {
  reply_markup: {
    inline_keyboard: [
      [{ text: 'ℹ️ Информация', callback_data: 'info' }],
      [{ text: '🔔 Подписаться', callback_data: 'subscribe' }]
    ]
  }
};

const infoKeyboard = {
  reply_markup: {
    inline_keyboard: [
      [{ text: 'Что такое Us', callback_data: 'info_us' }],
      [{ text: 'Что такое подписка', callback_data: 'info_sub' }],
      [{ text: 'Что такое FriendShip', callback_data: 'info_fs' }],
      [{ text: 'Назад', callback_data: 'back' }]
    ]
  }
};

// на /start
bot.onText(/\/start/, async (msg) => {
  const chatId = msg.chat.id;
  const text = `Привет! Я — бот уведомлений FriendShip.\n\nЯ могу:
- Напоминать о мероприятиях, на которые ты записан.
- Принимать подписки по 'Us'.
Нажми кнопку ниже, чтобы продолжить.`;
  await bot.sendMessage(chatId, text, mainKeyboard);
});

// обработка inline-кнопок
bot.on('callback_query', async (callbackQuery) => {
  const data = callbackQuery.data;
  const chatId = callbackQuery.message.chat.id;
  try {
    if (data === 'info') {
      await bot.answerCallbackQuery(callbackQuery.id);
      await bot.sendMessage(chatId, 'Информация — выбирай пункт:', infoKeyboard);
      return;
    }
    if (data === 'back') {
      await bot.answerCallbackQuery(callbackQuery.id);
      await bot.sendMessage(chatId, 'Главное меню:', mainKeyboard);
      return;
    }
    if (data === 'info_us' || data === 'info_sub' || data === 'info_fs') {
      await bot.answerCallbackQuery(callbackQuery.id);
      const map = {
        info_us: '*Что такое Us*\n\n`Us` — уникальный идентификатор пользователя в системе FriendShip. Его нужно ввести, чтобы связать аккаунт с Telegram.',
        info_sub: '*Что такое подписка*\n\nПодписка — это привязка TelegramID к вашему Us, чтобы бот мог отправлять уведомления о мероприятиях.',
        info_fs: '*Что такое FriendShip*\n\nFriendShip — это система/сервис для организации мероприятий и управления регистрацией.'
      };
      await bot.sendMessage(chatId, map[data], { parse_mode: 'Markdown', reply_markup: { inline_keyboard: [[{ text: 'Назад', callback_data: 'info' }]] }});
      return;
    }
    if (data === 'subscribe') {
      // запрашиваем Us у пользователя — переход в ожидание
      await bot.answerCallbackQuery(callbackQuery.id);
      await bot.sendMessage(chatId, 'Введи, пожалуйста, свой `Us` в системе FriendShip (просто отправь как сообщение).', { parse_mode: 'Markdown' });
      // отмечаем пользователя, что он в режиме ввода Us — простая импровизация через Map
      pendingUsRequests.set(chatId, true);
      return;
    }
    await bot.answerCallbackQuery(callbackQuery.id);
  } catch (err) {
    console.error('callback_query error', err);
  }
});

// простая Map для отслеживания ожидания ввода Us
const pendingUsRequests = new Map();

// обработка обычных сообщений
bot.on('message', async (msg) => {
  const chatId = msg.chat.id;
  const text = msg.text ? msg.text.trim() : '';
  // если сообщение — команда /unsubscriptions
  if (text === '/unsubscriptions') {
    await handleUnsubscribeCommand(chatId);
    return;
  }

  // если ожидаем Us
  if (pendingUsRequests.get(chatId)) {
    pendingUsRequests.delete(chatId);
    const Us = text;
    // basic validation
    if (!Us || Us.length > 200) {
      await bot.sendMessage(chatId, 'Некорректный Us. Попробуй ещё раз.');
      return;
    }
    // выполняем запрос подписки
    try {
      const telegramID = msg.from && msg.from.id;
      await bot.sendMessage(chatId, 'Запрашиваю подписку на стороне FriendShip...');
      await api.subscribeUser({ Us, TelegramID: telegramID });
      await bot.sendMessage(chatId, 'Успешно подписан ✅. Ты будешь получать уведомления.');
    } catch (err) {
      console.error('subscribeUser error', err?.response?.data || err.message || err);
      const details = err?.response?.data?.error || err?.message || 'внутренняя ошибка';
      await bot.sendMessage(chatId, `Не удалось подписаться: ${details}`);
    }
    return;
  }

  // Отвечаем на любые другие ненужные сообщения кратко
  if (text && !text.startsWith('/')) {
    await bot.sendMessage(chatId, 'Если хочешь подписаться — нажми «Подписаться» или напиши /unsubscriptions для отписки.', mainKeyboard);
  }
});

async function handleUnsubscribeCommand(chatId) {
  try {
    const telegramID = chatId;
    await bot.sendMessage(chatId, 'Запрашиваю отписку...');
    await api.unsubscribeByTelegramID({ TelegramID: telegramID });
    await bot.sendMessage(chatId, 'Успешно отписан ✅');
  } catch (err) {
    console.error('unsubscribe error', err?.response?.data || err.message || err);
    const details = err?.response?.data?.error || err?.message || 'внутренняя ошибка';
    await bot.sendMessage(chatId, `Не удалось отписаться: ${details}`);
  }
}

// экспорт для server.js (нужен для рассылки)
module.exports = {
  bot,
  sendBroadcast: async function sendBroadcast(items = []) {
    // items: [{ telegramIds: [id,...], imageUrl?, title?, text? }, ...]
    if (!Array.isArray(items)) throw new Error('items must be array');
    const limit = pLimit(5); // параллельно обрабатываем до 5 задач (блоков)
    for (const block of items) {
      // для каждого получателя создаём очередь отправки с задержкой между сообщениями
      const telegramIds = Array.isArray(block.telegramIds) ? block.telegramIds : [];
      for (const tid of telegramIds) {
        // каждый отправляем с задержкой, чтобы снизить риск rate limit
        await new Promise(resolve => setTimeout(resolve, config.BROADCAST_RATE_MS));
        // отправка в отдельной функции, чтобы ловить исключения
        limit(() => (async () => {
          try {
            // если есть картинка — сначала фото с подписью
            const captionParts = [];
            if (block.title) captionParts.push(`*${escapeMarkdown(block.title)}*`);
            if (block.text) captionParts.push(escapeMarkdown(block.text));
            const caption = captionParts.join('\n\n');
            if (block.imageUrl) {
              await bot.sendPhoto(tid, block.imageUrl, caption ? { caption, parse_mode: 'Markdown' } : {});
            } else if (caption) {
              await bot.sendMessage(tid, caption, { parse_mode: 'Markdown' });
            } else {
              // пустое сообщение — пропускаем
            }
          } catch (err) {
            // логируем ошибки; если чат закрыт — помечаем
            console.error(`send to ${tid} failed:`, err?.response?.body || err?.response || err?.message || err);
          }
        })()).catch(e => console.error('limit wrapper error', e));
      }
      // дождаться завершения всех limit-задач текущего блока
      await limit(() => Promise.resolve());
    }
  }
};

// простая защита для Markdown (экранирование), минимально необходимая
function escapeMarkdown(text = '') {
  // экранируем символы MarkdownV2 — но мы используем Markdown (not V2) - простое экранирование нескольких символов:
  return String(text).replace(/([_*[\]()~`>#+\-=|{}.!])/g, '\\$1');
}
