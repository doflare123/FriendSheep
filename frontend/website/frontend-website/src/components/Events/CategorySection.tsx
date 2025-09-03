"use client";

import {SectionData} from "../../types/Events"
import { useRouter } from 'next/navigation';
import EventCard from "./EventCard"
import styles from '../../styles/MainPage.module.css';
import React, { useRef, useState, useEffect } from 'react';
import CategoryLabel from "./CategoryLabel";

interface CategorySectionProps {
    section: SectionData;
    title: string;
    showCategoryLabel?: boolean;
    clickable?: boolean;
}

const CategorySection: React.FC<CategorySectionProps> = ({ 
    section, 
    title, 
    showCategoryLabel = true,
    clickable = true
}) => {
    const router = useRouter();
    const scrollRef = useRef<HTMLDivElement>(null);
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(true);

    const checkScrollButtons = () => {
        if (scrollRef.current) {
            const { scrollLeft, scrollWidth, clientWidth } = scrollRef.current;
            setCanScrollLeft(scrollLeft > 0);
            setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 1);
        }
    };

    useEffect(() => {
        const scrollElement = scrollRef.current;
        if (scrollElement) {
            checkScrollButtons();
            scrollElement.addEventListener('scroll', checkScrollButtons);
            
            // Проверяем при изменении размера окна
            const resizeObserver = new ResizeObserver(checkScrollButtons);
            resizeObserver.observe(scrollElement);

            return () => {
                scrollElement.removeEventListener('scroll', checkScrollButtons);
                resizeObserver.disconnect();
            };
        }
    }, [section.categories]);

    const scroll = (direction: 'left' | 'right') => {
        if (scrollRef.current) {
            const scrollAmount = 580; // ширина карточки + отступы
            scrollRef.current.scrollBy({
                left: direction === 'left' ? -scrollAmount : scrollAmount,
                behavior: 'smooth'
            });
        }
    };

    const onClick = () => {
        if (title && clickable) {
            router.push(`/search/events?category=${title}&pattern=${section.pattern}`);
        }
    };

    return (
        <div className={styles.categorySection}>
            {showCategoryLabel && (
                <div className={styles.categoryHeader}>
                    <CategoryLabel title={title} patternUrl={section.pattern} onClick={onClick} clickable={clickable} />
                </div>
            )}
            <div className={styles.cardsContainer}>
                {canScrollLeft && (
                    <button 
                        className={`${styles.scrollButton} ${styles.scrollLeft}`}
                        onClick={() => scroll('left')}
                    >
                        ←
                    </button>
                )}
                <div className={styles.cardsScroll} ref={scrollRef}>
                    {section.categories.map((event) => (
                        <EventCard key={event.id} {...event} />
                    ))}
                </div>
                {canScrollRight && (
                    <button 
                        className={`${styles.scrollButton} ${styles.scrollRight}`}
                        onClick={() => scroll('right')}
                    >
                        →
                    </button>
                )}
            </div>
        </div>
    );
};

export default CategorySection;