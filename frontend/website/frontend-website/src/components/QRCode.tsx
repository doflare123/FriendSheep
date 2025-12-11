import Image from 'next/image';
import { MOBILE_APP_URL } from '@/Constants';
import styles from '@/styles/QRCodeStyles.module.css';

interface QRCodeProps {
  size?: number;
  showText?: boolean;
}

export default function QRCode({ size = 200, showText = true }: QRCodeProps) {
  return (
    <div className={styles.qrCodeContainer}>
      <a 
        href={MOBILE_APP_URL}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.qrCodeLink}
      >
        <Image 
          src="/social/qr_mob.png" 
          alt="QR код для скачивания приложения" 
          width={size} 
          height={size}
          className={styles.qrCode}
          style={{ borderRadius: '16px' }}
        />
      </a>
      {showText && (
        <p className={styles.qrCodeText}>Отсканируйте QR-код для скачивания</p>
      )}
    </div>
  );
}