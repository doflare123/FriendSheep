import React from 'react';

interface FormButtonProps {
    children: React.ReactNode;
    type?: 'button' | 'submit' | 'reset';
    onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
    className?: string;
    disabled?: boolean;
}

const FormButton: React.FC<FormButtonProps> = ({ 
    children, 
    type = "button", 
    onClick,
    className = "button-primary",
    disabled = false 
}) => {
    return (
        <button
            className={className}
            type={type}
            onClick={onClick}
            disabled={disabled}
            >
            {children}
        </button>
    );
};

export default FormButton;