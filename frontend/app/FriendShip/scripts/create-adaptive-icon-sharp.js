const sharp = require('sharp');
const path = require('path');

async function createAdaptiveIcon() {
  const inputPath = path.join(__dirname, '../assets/images/logo.png');
  const outputPath = path.join(__dirname, '../assets/images/adaptive-icon.png');
  
  try {
    console.log('📂 Загрузка изображения:', inputPath);

    const scaledSize = Math.floor(1024 * 0.65);
    const padding = Math.floor((1024 - scaledSize) / 2);
    
    console.log('📏 Создание адаптивной иконки...');
    console.log(`   Размер логотипа: ${scaledSize}x${scaledSize}`);
    console.log(`   Отступы: ${padding}px`);
    
    await sharp(inputPath)
      .resize(scaledSize, scaledSize, {
        fit: 'contain',
        background: { r: 0, g: 0, b: 0, alpha: 0 }
      })
      .extend({
        top: padding,
        bottom: padding,
        left: padding,
        right: padding,
        background: { r: 0, g: 0, b: 0, alpha: 0 }
      })
      .png()
      .toFile(outputPath);
    
    console.log('✅ Adaptive icon успешно создан:', outputPath);
    console.log('📱 Следующий шаг: обновите app.json (если ещё не сделали)');
    console.log('🔨 Затем пересоберите: eas build --platform android');
  } catch (error) {
    console.error('❌ Ошибка:', error.message);
    console.error(error);
    process.exit(1);
  }
}

createAdaptiveIcon();