import React from 'react';
import Image from 'next/image';
import styles from '../../styles/profile/ProfileSection2.module.css';

interface StatisticsTileProps {
  title: string;
  value: string | number;
  icon: string;
  subtitle?: string;
}

const StatisticsTile: React.FC<StatisticsTileProps> = ({
  title,
  value,
  icon,
  subtitle
}) => {
  return (
    <div className={styles.statisticsTile}>
      <div className={styles.tileHeader}>
        <div className={styles.tileTitle}>{title}</div>
      </div>
      <div className={styles.tileIcon}>
        <Image src={icon} alt={title} width={100} height={100} />
      </div>
      <div className={styles.tileValue}>{value}</div>
      {subtitle && (
        <div className={styles.tileSubtitle}>{subtitle}</div>
      )}
    </div>
  );
};

export default StatisticsTile;