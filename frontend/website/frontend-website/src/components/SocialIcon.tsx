'use client';

import React from 'react';
import Image from 'next/image';
import { getSocialIcon } from '@/Constants';

// Компонент для иконок социальных сетей
interface SocialIconProps {
    href: string;
    alt: string;
    size: number;
}

const SocialIcon: React.FC<SocialIconProps> = ({ href, alt, size }) => {
    const isDiscord = href.toLowerCase().includes('discord') || 
                     alt.toLowerCase().includes('discord') || 
                     alt.toLowerCase().includes('ds');
    
    const title = isDiscord 
        ? `${alt}\n*Деятельность организации запрещена на территории РФ`
        : alt;
    
    return (
        <a 
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className="social-icon"
            title={title}
        >
            <Image 
                src={getSocialIcon(href)} 
                alt={alt} 
                width={size} 
                height={size} 
            />
        </a>
    );
};

export default SocialIcon;