import React from 'react';

interface FormInputProps {
    id: string;
    label: string;
    type?: 'text' | 'password' | 'email' | 'number' | 'tel';
    placeholder?: string;
    value: string;
    onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
    required?: boolean;
    className?: string;
}

const FormInput: React.FC<FormInputProps> = ({ 
    id, 
    label, 
    type = "text", 
    placeholder, 
    value, 
    onChange, 
    required = false,
    className = ""
}) => {
    return (
        <div className={className}>
            <label className="label-style" htmlFor={id}>
                {label}
            </label>
            <input
                id={id}
                className="input-style"
                type={type}
                placeholder={placeholder}
                value={value}
                onChange={onChange}
                required={required}
            />
        </div>
    );
};

export default FormInput;