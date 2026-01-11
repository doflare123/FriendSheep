'use client';

import React, { useState } from 'react';
import Image from 'next/image';
import { getSocialIcon } from '@/Constants';
import ConfirmModal from './ConfirmModal';

interface SocialIconProps {
    href: string;
    alt: string;
    size: number;
    onClick?: (e: React.MouseEvent) => void;
    isClickable?: boolean;
}

const SocialIcon: React.FC<SocialIconProps> = ({ href, alt, size, onClick, isClickable = false }) => {
    const [showModal, setShowModal] = useState(false);
    
    const isDiscord = href.toLowerCase().includes('discord') || 
                     alt.toLowerCase().includes('discord') || 
                     alt.toLowerCase().includes('ds');
    
    const title = isDiscord 
        ? `${alt}\n*Деятельность организации запрещена на территории РФ`
        : alt;
    
    const handleClick = (e: React.MouseEvent) => {
        e.preventDefault();
        
        if (isClickable && onClick) {
            onClick(e);
        } else {
            setShowModal(true);
        }
    };
    
    const handleConfirm = () => {
        setShowModal(false);
        window.open(href, '_blank', 'noopener,noreferrer');
    };
    
    const handleCancel = () => {
        setShowModal(false);
    };
    
    if (isClickable) {
        return (
            <div 
                onClick={handleClick}
                className="social-icon"
                title={title}
                style={{ cursor: 'pointer' }}
            >
                <Image 
                    src={getSocialIcon(href)} 
                    alt={alt} 
                    width={size} 
                    height={size} 
                />
            </div>
        );
    }
    
    return (
        <>
            <div 
                onClick={handleClick}
                className="social-icon"
                title={title}
                style={{ cursor: 'pointer' }}
            >
                <Image 
                    src={getSocialIcon(href)} 
                    alt={alt} 
                    width={size} 
                    height={size} 
                />
            </div>
            
            <ConfirmModal 
                isOpen={showModal}
                onConfirm={handleConfirm}
                onCancel={handleCancel}
                url={href}
            />
        </>
    );
};

export default SocialIcon;