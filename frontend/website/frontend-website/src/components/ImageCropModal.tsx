import React, { useState, useRef, useEffect } from 'react';
import styles from '@/styles/ImageCropModal.module.css';

interface ImageCropModalProps {
  imageFile: File;
  onSave: (blob: Blob) => void;
  onCancel: () => void;
  title?: string;
  cropShape?: 'circle' | 'square' | 'rectangle';
  aspectRatio?: number; // ширина / высота (например, 2.25 для 360/160)
  finalSize?: number; // для square и circle - размер стороны, для rectangle - ширина
}

const ImageCropModal: React.FC<ImageCropModalProps> = ({ 
  imageFile, 
  onSave, 
  onCancel,
  title = 'Настройте изображение',
  cropShape = 'circle',
  aspectRatio = 1,
  finalSize = 300
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [image, setImage] = useState<HTMLImageElement | null>(null);
  const [crop, setCrop] = useState({ x: 0, y: 0, width: 200, height: 200 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });

  useEffect(() => {
    if (imageFile) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const img = new Image();
        img.onload = () => {
          setImage(img);
          
          // Вычисляем начальный размер crop
          let cropWidth, cropHeight;
          
          if (cropShape === 'rectangle') {
            // Для прямоугольника используем aspect ratio
            const maxWidth = Math.min(img.width, 400);
            const maxHeight = Math.min(img.height, 400);
            
            // Определяем размеры с учетом aspect ratio
            if (maxWidth / aspectRatio <= maxHeight) {
              cropWidth = maxWidth;
              cropHeight = maxWidth / aspectRatio;
            } else {
              cropHeight = maxHeight;
              cropWidth = maxHeight * aspectRatio;
            }
          } else {
            // Для круга и квадрата
            const size = Math.min(img.width, img.height, 400);
            cropWidth = size;
            cropHeight = size;
          }
          
          // Центрируем crop
          setCrop({
            x: (img.width - cropWidth) / 2,
            y: (img.height - cropHeight) / 2,
            width: cropWidth,
            height: cropHeight
          });
        };
        img.src = e.target?.result as string;
      };
      reader.readAsDataURL(imageFile);
    }
  }, [imageFile, cropShape, aspectRatio]);

  useEffect(() => {
    if (image && canvasRef.current) {
      drawCanvas();
    }
  }, [image, crop]);

  const drawCanvas = () => {
    const canvas = canvasRef.current;
    if (!canvas || !image) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    // Устанавливаем размер canvas
    const displayWidth = 400;
    const displayHeight = 400;
    canvas.width = displayWidth;
    canvas.height = displayHeight;

    // Очищаем canvas
    ctx.clearRect(0, 0, displayWidth, displayHeight);

    // Масштабируем изображение для отображения
    const scale = Math.min(displayWidth / image.width, displayHeight / image.height);
    const scaledWidth = image.width * scale;
    const scaledHeight = image.height * scale;
    const offsetX = (displayWidth - scaledWidth) / 2;
    const offsetY = (displayHeight - scaledHeight) / 2;

    // Рисуем изображение
    ctx.drawImage(image, offsetX, offsetY, scaledWidth, scaledHeight);

    // Затемняем всё кроме области обрезки
    ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
    ctx.fillRect(0, 0, displayWidth, displayHeight);

    // Очищаем область обрезки
    const cropX = crop.x * scale + offsetX;
    const cropY = crop.y * scale + offsetY;
    const cropWidth = crop.width * scale;
    const cropHeight = crop.height * scale;

    ctx.save();
    ctx.globalCompositeOperation = 'destination-out';
    ctx.beginPath();
    
    if (cropShape === 'circle') {
      ctx.arc(cropX + cropWidth / 2, cropY + cropHeight / 2, cropWidth / 2, 0, Math.PI * 2);
    } else {
      ctx.rect(cropX, cropY, cropWidth, cropHeight);
    }
    
    ctx.fill();
    ctx.restore();

    // Рисуем границу
    ctx.strokeStyle = '#37A2E6';
    ctx.lineWidth = 3;
    ctx.beginPath();
    
    if (cropShape === 'circle') {
      ctx.arc(cropX + cropWidth / 2, cropY + cropHeight / 2, cropWidth / 2, 0, Math.PI * 2);
    } else {
      ctx.rect(cropX, cropY, cropWidth, cropHeight);
    }
    
    ctx.stroke();
  };

  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas || !image) return;

    const rect = canvas.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    const scale = Math.min(400 / image.width, 400 / image.height);
    const scaledWidth = image.width * scale;
    const scaledHeight = image.height * scale;
    const offsetX = (400 - scaledWidth) / 2;
    const offsetY = (400 - scaledHeight) / 2;

    const cropX = crop.x * scale + offsetX;
    const cropY = crop.y * scale + offsetY;
    const cropWidth = crop.width * scale;
    const cropHeight = crop.height * scale;

    // Проверяем, попали ли мы в область обрезки
    let isInside = false;
    
    if (cropShape === 'circle') {
      const centerX = cropX + cropWidth / 2;
      const centerY = cropY + cropHeight / 2;
      const distance = Math.sqrt(Math.pow(x - centerX, 2) + Math.pow(y - centerY, 2));
      isInside = distance <= cropWidth / 2;
    } else {
      isInside = x >= cropX && x <= cropX + cropWidth && y >= cropY && y <= cropY + cropHeight;
    }

    if (isInside) {
      setIsDragging(true);
      setDragStart({ x: x - cropX, y: y - cropY });
    }
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDragging || !image) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    const scale = Math.min(400 / image.width, 400 / image.height);
    const scaledWidth = image.width * scale;
    const scaledHeight = image.height * scale;
    const offsetX = (400 - scaledWidth) / 2;
    const offsetY = (400 - scaledHeight) / 2;

    let newX = (x - dragStart.x - offsetX) / scale;
    let newY = (y - dragStart.y - offsetY) / scale;

    // Ограничиваем перемещение границами изображения
    newX = Math.max(0, Math.min(newX, image.width - crop.width));
    newY = Math.max(0, Math.min(newY, image.height - crop.height));

    setCrop({ ...crop, x: newX, y: newY });
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleZoom = (delta: number) => {
    if (!image) return;

    let newWidth, newHeight;
    
    if (cropShape === 'rectangle') {
      // Для прямоугольника сохраняем aspect ratio
      newWidth = Math.max(100, Math.min(image.width, crop.width + delta));
      newHeight = newWidth / aspectRatio;
      
      // Проверяем, не выходит ли за границы по высоте
      if (newHeight > image.height) {
        newHeight = image.height;
        newWidth = newHeight * aspectRatio;
      }
    } else {
      // Для круга и квадрата
      const newSize = Math.max(100, Math.min(Math.min(image.width, image.height), crop.width + delta));
      newWidth = newSize;
      newHeight = newSize;
    }
    
    const centerX = crop.x + crop.width / 2;
    const centerY = crop.y + crop.height / 2;
    
    let newX = centerX - newWidth / 2;
    let newY = centerY - newHeight / 2;

    // Корректируем позицию при изменении размера
    newX = Math.max(0, Math.min(newX, image.width - newWidth));
    newY = Math.max(0, Math.min(newY, image.height - newHeight));

    setCrop({ x: newX, y: newY, width: newWidth, height: newHeight });
  };

  const handleSave = () => {
    if (!image) return;

    // Вычисляем финальные размеры
    let finalWidth, finalHeight;
    
    if (cropShape === 'rectangle') {
      finalWidth = finalSize;
      finalHeight = Math.round(finalSize / aspectRatio);
    } else {
      finalWidth = finalSize;
      finalHeight = finalSize;
    }

    // Создаём canvas для финального изображения
    const finalCanvas = document.createElement('canvas');
    finalCanvas.width = finalWidth;
    finalCanvas.height = finalHeight;
    const ctx = finalCanvas.getContext('2d');
    
    if (!ctx) return;

    if (cropShape === 'circle') {
      // Рисуем круглое изображение
      ctx.beginPath();
      ctx.arc(finalWidth / 2, finalHeight / 2, finalWidth / 2, 0, Math.PI * 2);
      ctx.closePath();
      ctx.clip();
    }

    // Рисуем выбранную область
    ctx.drawImage(
      image,
      crop.x, crop.y, crop.width, crop.height,
      0, 0, finalWidth, finalHeight
    );

    // Конвертируем в blob
    finalCanvas.toBlob((blob) => {
      if (blob) {
        onSave(blob);
      }
    }, 'image/jpeg', 0.95);
  };

  if (!image) return null;

  const getShapeName = () => {
    if (cropShape === 'circle') return 'круг';
    if (cropShape === 'rectangle') return 'прямоугольник';
    return 'квадрат';
  };

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        <h3 className={styles.title}>{title}</h3>
        
        <canvas
          ref={canvasRef}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          className={`${styles.canvas} ${isDragging ? styles.canvasDragging : ''}`}
        />

        <div className={styles.zoomControls}>
          <button
            onClick={() => handleZoom(-20)}
            className={styles.zoomButton}
            type="button"
          >
            −
          </button>
          <button
            onClick={() => handleZoom(20)}
            className={styles.zoomButton}
            type="button"
          >
            +
          </button>
        </div>

        <div className={styles.buttonGroup}>
          <button
            onClick={handleSave}
            className={styles.saveButton}
            type="button"
          >
            Сохранить
          </button>
          <button
            onClick={onCancel}
            className={styles.cancelButton}
            type="button"
          >
            Отмена
          </button>
        </div>

        <p className={styles.hint}>
          Перетаскивайте {getShapeName()} для выбора области, используйте + и − для масштабирования
        </p>
      </div>
    </div>
  );
};

export default ImageCropModal;