import React from 'react';

type TextSize = 'xs' | 'sm' | 'md' | 'lg';

interface FormTextProps {
  children: React.ReactNode;
  className?: string;
  size?: TextSize;
}

const FormText: React.FC<FormTextProps> = ({ 
  children, 
  className = "text-left text-sm text-gray-600 mt-2",
  size = "sm"
}) => {
  const sizeClasses: Record<TextSize, string> = {
    xs: "text-xs",
    sm: "text-sm", 
    md: "text-base",
    lg: "text-lg"
  };

  const finalClassName = className.replace(/text-\w+/, sizeClasses[size]);

  return (
    <div className={finalClassName}>
      {children}
    </div>
  );
};

export default FormText;