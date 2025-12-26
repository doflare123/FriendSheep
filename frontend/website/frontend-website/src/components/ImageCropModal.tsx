import React, { useState, useRef } from 'react';
import ReactCrop, { Crop, PixelCrop } from 'react-image-crop';
import 'react-image-crop/dist/ReactCrop.css';
import styles from '@/styles/ImageCropModal.module.css';

interface ImageCropModalProps {
  imageFile: File;
  onSave: (blob: Blob) => void;
  onCancel: () => void;
  title?: string;
  cropShape?: 'circle' | 'square' | 'rectangle';
  aspectRatio?: number;
  finalSize?: number;
}

const ImageCropModal: React.FC<ImageCropModalProps> = ({
  imageFile,
  onSave,
  onCancel,
  title = 'Настройте изображение',
  cropShape = 'circle',
  aspectRatio = 1,
  finalSize = 300,
}) => {
  const [imageSrc, setImageSrc] = useState<string>('');
  const [crop, setCrop] = useState<Crop>();
  const [completedCrop, setCompletedCrop] = useState<PixelCrop | null>(null);
  const imgRef = useRef<HTMLImageElement | null>(null);

  React.useEffect(() => {
    if (imageFile) {
      const reader = new FileReader();
      reader.addEventListener('load', () => {
        setImageSrc(reader.result as string);
      });
      reader.readAsDataURL(imageFile);
    }
  }, [imageFile]);

  const onImageLoad = (e: React.SyntheticEvent<HTMLImageElement>) => {
    const img = e.currentTarget;
    imgRef.current = img;
    
    const { naturalWidth, naturalHeight } = img;
    const centerX = naturalWidth / 2;
    const centerY = naturalHeight / 2;
    
    let cropWidth, cropHeight;

    if (cropShape === 'rectangle') {
      const maxSize = Math.min(naturalWidth, naturalHeight) * 0.8;
      if (aspectRatio >= 1) {
        cropWidth = maxSize * aspectRatio;
        cropHeight = maxSize;
        if (cropWidth > naturalWidth * 0.9) {
          cropWidth = naturalWidth * 0.9;
          cropHeight = cropWidth / aspectRatio;
        }
      } else {
        cropHeight = maxSize / aspectRatio;
        cropWidth = maxSize;
        if (cropHeight > naturalHeight * 0.9) {
          cropHeight = naturalHeight * 0.9;
          cropWidth = cropHeight * aspectRatio;
        }
      }
    } else {
      const size = Math.min(naturalWidth, naturalHeight) * 0.8;
      cropWidth = size;
      cropHeight = size;
    }

    const scale = img.width / naturalWidth;

    setCrop({
      unit: 'px',
      x: (centerX - cropWidth / 2) * scale,
      y: (centerY - cropHeight / 2) * scale,
      width: cropWidth * scale,
      height: cropHeight * scale,
    });
  };

  const getCroppedImg = async (
    image: HTMLImageElement,
    crop: PixelCrop,
    cropShape: 'circle' | 'square' | 'rectangle',
    finalSize: number,
    aspectRatio: number
  ): Promise<Blob> => {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');

    if (!ctx) {
      throw new Error('No 2d context');
    }

    let finalWidth, finalHeight;

    if (cropShape === 'rectangle') {
      finalWidth = finalSize;
      finalHeight = Math.round(finalSize / aspectRatio);
    } else {
      finalWidth = finalSize;
      finalHeight = finalSize;
    }

    canvas.width = finalWidth;
    canvas.height = finalHeight;

    const scaleX = image.naturalWidth / image.width;
    const scaleY = image.naturalHeight / image.height;

    if (cropShape === 'circle') {
      ctx.beginPath();
      ctx.arc(finalWidth / 2, finalHeight / 2, finalWidth / 2, 0, Math.PI * 2);
      ctx.closePath();
      ctx.clip();
    }

    ctx.drawImage(
      image,
      crop.x * scaleX,
      crop.y * scaleY,
      crop.width * scaleX,
      crop.height * scaleY,
      0,
      0,
      finalWidth,
      finalHeight
    );

    return new Promise((resolve, reject) => {
      canvas.toBlob(
        (blob) => {
          if (blob) {
            resolve(blob);
          } else {
            reject(new Error('Canvas is empty'));
          }
        },
        'image/jpeg',
        0.95
      );
    });
  };

  const handleSave = async () => {
    if (!completedCrop || !imgRef.current) return;

    try {
      const croppedImage = await getCroppedImg(
        imgRef.current,
        completedCrop,
        cropShape,
        finalSize,
        aspectRatio
      );
      onSave(croppedImage);
    } catch (e) {
      console.error(e);
    }
  };

  const getShapeName = () => {
    if (cropShape === 'circle') return 'круг';
    if (cropShape === 'rectangle') return 'прямоугольник';
    return 'квадрат';
  };

  if (!imageSrc) return null;

  const aspect = cropShape === 'circle' || cropShape === 'square' ? 1 : aspectRatio;

  return (
    <div className={styles.overlay} onClick={(e) => e.target === e.currentTarget && onCancel()}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <h3 className={styles.title}>{title}</h3>

        <div
          style={{
            width: '100%',
            height: '400px',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            marginBottom: '15px',
            overflow: 'hidden',
            backgroundColor: '#000',
            borderRadius: '8px',
          }}
        >
          <ReactCrop
            crop={crop}
            onChange={(c) => setCrop(c)}
            onComplete={(c) => setCompletedCrop(c)}
            aspect={aspect}
            circularCrop={cropShape === 'circle'}
            minHeight={50}
            minWidth={50}
            keepSelection
            style={{
              maxHeight: '400px',
              maxWidth: '100%',
            }}
          >
            <img
              src={imageSrc}
              onLoad={onImageLoad}
              style={{
                maxHeight: '400px',
                maxWidth: '100%',
                display: 'block',
              }}
              alt="Crop"
            />
          </ReactCrop>
        </div>

        <div className={styles.buttonGroup}>
          <button onClick={handleSave} className={styles.saveButton} type="button">
            Сохранить
          </button>
          <button onClick={onCancel} className={styles.cancelButton} type="button">
            Отмена
          </button>
        </div>

        <p className={styles.hint}>
          Перетаскивайте {getShapeName()} и тяните за углы для изменения размера
        </p>
      </div>
    </div>
  );
};

export default ImageCropModal;