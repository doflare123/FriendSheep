import React from 'react';
import '../../styles/CategoryLabel.css';

interface CategoryLabelProps {
    title: string;
    patternUrl: string; // путь к паттерну (например, /patterns/pattern1.png)
    clickable?: boolean; // новый пропс для управления кликабельностью
    onClick?: () => void; // функция обработки клика
}

const CategoryLabel: React.FC<CategoryLabelProps> = ({ 
    title, 
    patternUrl, 
    clickable = true, // по умолчанию кликабельный
    onClick 
}) => {
    const handleClick = () => {
        if (clickable && onClick) {
            onClick();
        }
    };

    return (
        <div
            className={`categoryLabel ${clickable ? 'categoryLabel--clickable' : 'categoryLabel--static'}`}
            style={{
                backgroundImage: `url(${patternUrl}), linear-gradient(var(--color-primary-blue-dark), var(--color-primary-blue-dark))`,
            }}
            onClick={handleClick}
        >
            <span className="categoryLabelText">{title}</span>
        </div>
    );
};

export default CategoryLabel;