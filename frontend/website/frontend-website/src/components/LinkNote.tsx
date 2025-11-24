import React from 'react';

interface LinkNoteProps {
  children: React.ReactNode;
  className?: string;
  leftLink?: React.ReactNode;
}

const LinkNote: React.FC<LinkNoteProps> = ({ 
  children, 
  className = "link-small-note",
  leftLink 
}) => {
  // Если передан leftLink, используем flex layout для двух элементов
  if (leftLink) {
    return (
      <div className="flex justify-between items-start text-sm">
        <div className="-mt-1">{leftLink}</div>
        <div className={className}>{children}</div>
      </div>
    );
  }
  
  // Иначе работаем как обычно (обратная совместимость)
  return (
    <div className={className}>
      {children}
    </div>
  );
};

export default LinkNote;