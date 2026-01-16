import fs from 'fs';
import path from 'path';
import sharp from 'sharp';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const sourceDir = path.join(__dirname, 'public', 'images', 'teams');
const files = fs.readdirSync(sourceDir).filter(file => file.endsWith('.png'));

console.log('🚀 Начинаем конвертацию изображений в WebP с помощью Sharp...\n');

let totalSaved = 0;
let converted = 0;

async function convertImages() {
  for (const file of files) {
    const inputPath = path.join(sourceDir, file);
    const outputPath = path.join(sourceDir, file.replace('.png', '.webp'));

    try {
      // Получаем размер оригинального файла
      const originalSize = fs.statSync(inputPath).size;

      // Конвертируем с помощью sharp
      await sharp(inputPath)
        .webp({
          quality: 85,
          effort: 6 // Максимальное качество сжатия
        })
        .toFile(outputPath);

      const webpSize = fs.statSync(outputPath).size;
      const saved = originalSize - webpSize;
      const percentage = ((saved / originalSize) * 100).toFixed(1);

      console.log(`✅ ${file} -> ${file.replace('.png', '.webp')}`);
      console.log(`   ${(originalSize / 1024).toFixed(1)}KB -> ${(webpSize / 1024).toFixed(1)}KB (${percentage}% сжатие)\n`);

      totalSaved += saved;
      converted++;

    } catch (error) {
      console.log(`❌ Ошибка конвертации ${file}:`, error.message);
    }
  }

  console.log(`🎉 Конвертация завершена!`);
  console.log(`📊 Конвертировано: ${converted}/${files.length} изображений`);
  console.log(`💾 Общий объем сжатия: ${(totalSaved / 1024).toFixed(1)}KB`);
}

convertImages().catch(console.error);