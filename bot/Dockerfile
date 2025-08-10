# Указываем базовый образ
FROM node:20-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем package.json и package-lock.json для установки зависимостей
COPY package*.json ./

# Устанавливаем зависимости
RUN npm install --only=production

# Копируем весь проект
COPY . .

# Задаём переменные окружения по умолчанию (можно переопределить через docker run -e ...)
ENV NODE_ENV=production

# Запускаем бота
CMD ["npm", "start"]
