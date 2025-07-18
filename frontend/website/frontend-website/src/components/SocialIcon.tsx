import React from 'react';
import Image from 'next/image';

// Компонент для иконок социальных сетей
interface SocialIconProps {
    href: string;
    iconName: string;
    alt: string;
    size: number;
}

const SocialIcon: React.FC<SocialIconProps> = ({ href, iconName, alt, size }) => {
    return (
        <a 
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className={`social-icon`}
        >
            <Image src={`/social/${iconName}.png`} alt={alt} width={size} height={size} />
        </a>
    );
};

export default SocialIcon;