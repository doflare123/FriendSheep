'use client';

import React from 'react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import section3Styles from '../../styles/profile/ProfileSection3.module.css';

interface SubscriptionItemProps {
  id: number;
  name: string;
  small_description: string;
  image: string;
}

export default function SubscriptionItem({ id, name, small_description, image }: SubscriptionItemProps) {
  const router = useRouter();

  const handleClick = () => {
    router.push(`/groups/profile/${id}`);
  };

  return (
    <div className={section3Styles.groupItem} onClick={handleClick}>
      <div className={section3Styles.groupContent}>
        <div className={section3Styles.groupImageWrapper}>
          <Image
            src={image}
            alt={name}
            width={50}
            height={50}
            className={section3Styles.groupImage}
          />
        </div>
        <div className={section3Styles.groupInfo}>
          <span className={section3Styles.groupName}>{name}</span>
          <span className={section3Styles.memberCount}>{small_description}</span>
        </div>
      </div>
    </div>
  );
}
