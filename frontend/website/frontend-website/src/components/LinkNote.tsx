import React from 'react';

interface LinkNoteProps {
  children: React.ReactNode;
  className?: string;
}

const LinkNote: React.FC<LinkNoteProps> = ({ 
  children, 
  className = "link-small-note" 
}) => {
  return (
    <div className={className}>
      {children}
    </div>
  );
};

export default LinkNote;