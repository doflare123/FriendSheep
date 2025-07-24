"use client";

import {SectionData} from "../../types/Events"
import EventCard from "./EventCard"
import '../../styles/MainPage.css';
import React, { useRef, useState, useEffect } from 'react';
import CategoryLabel from "./CategoryLabel"; // добавим импорт

const CategorySection: React.FC<{ section: SectionData, title: string }> = ({ section, title }) => {
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

    return (
        <div className='categorySection'>
            <div className='categoryHeader'>
                <CategoryLabel title={title} patternUrl={section.pattern} />
            </div>
            <div className='cardsContainer'>
                {canScrollLeft && (
                    <button 
                        className={`scrollButton scrollLeft`}
                        onClick={() => scroll('left')}
                    >
                        ←
                    </button>
                )}
                <div className='cardsScroll' ref={scrollRef}>
                    {section.categories.map((event) => (
                        <EventCard key={event.id} {...event} />
                    ))}
                </div>
                {canScrollRight && (
                    <button 
                        className={`scrollButton scrollRight`}
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