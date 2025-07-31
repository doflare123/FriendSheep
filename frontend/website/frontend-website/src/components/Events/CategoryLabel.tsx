// components/CategoryLabel.tsx
import React from 'react';
import '../../styles/CategoryLabel.css';

interface CategoryLabelProps {
    title: string;
    patternUrl: string; // путь к паттерну (например, /patterns/pattern1.png)
}

const CategoryLabel: React.FC<CategoryLabelProps> = ({ title, patternUrl }) => {
    return (
        <div
            className="categoryLabel"
            style={{
                backgroundImage: `url(${patternUrl}), linear-gradient(#316BC2, #316BC2)`,
            }}
        >
            <span className="categoryLabelText">{title}</span>
        </div>
    );
};

export default CategoryLabel;
