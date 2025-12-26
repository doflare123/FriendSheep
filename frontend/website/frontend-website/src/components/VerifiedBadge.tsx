import React, { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import styles from '@/styles/VerifiedBadge.module.css';

interface VerifiedBadgeProps {
  isVerified: boolean;
  size?: number;
}

const VerifiedBadge: React.FC<VerifiedBadgeProps> = ({ isVerified, size = 16 }) => {
  const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });
  const badgeRef = useRef<HTMLDivElement>(null);

  const handleMouseEnter = (e: React.MouseEvent) => {
    if (badgeRef.current) {
      const rect = badgeRef.current.getBoundingClientRect();
      setTooltipPosition({
        x: rect.left + rect.width / 2,
        y: rect.top - 8
      });
    }
  };

  if (!isVerified) return null;

  return (
    <div 
      className={styles.verifiedBadge} 
      ref={badgeRef}
      onMouseEnter={handleMouseEnter}
    >
      <Image 
        src="/profile/mark.png" 
        alt="verified" 
        width={size} 
        height={size}
        className={styles.verifiedIcon}
      />
      <span 
        className={styles.verifiedTooltip}
        style={{
          top: `${tooltipPosition.y}px`,
          left: `${tooltipPosition.x}px`,
          transform: 'translate(-50%, -100%)'
        }}
      >
        Этот пользователь надёжен
      </span>
    </div>
  );
};

export default VerifiedBadge;