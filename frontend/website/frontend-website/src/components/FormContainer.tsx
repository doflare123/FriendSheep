import React from 'react';

interface FormContainerProps {
    title: string;
    children: React.ReactNode;
    onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
    className?: string;
}

const FormContainer: React.FC<FormContainerProps> = ({ 
    title, 
    children, 
    onSubmit,
    className = "page-wrapper" 
}) => {
    return (
        <div className={className}>
            <main className="main-center">
                <div className="form-container">
                    <h1 className="heading-title">{title}</h1>
                    <form className="space-y-4" onSubmit={onSubmit}>
                        {children}
                    </form>
                </div>
            </main>
        </div>
    );
};

export default FormContainer;