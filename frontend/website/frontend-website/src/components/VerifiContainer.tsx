import React from 'react';

interface VerifiContainerProps {
	title: string;
	children: React.ReactNode;
	onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
	className?: string;
	containerClassName?: string;
}

const VerifiContainer: React.FC<VerifiContainerProps> = ({ 
	title,
	children,
	onSubmit,
	className = "page-wrapper",
}) => {
	return (
		<div className={className}>
			<main className="main-center">
                <div className="form-container max-w-xl">
                    <h1 className="heading-title">{title}</h1>

                    <div className="dots-animation">
                        <div className="dot dot1" />
                        <div className="dot dot2" />
                        <div className="dot dot3" />
                    </div>

                    <form className="space-y-4 mt-12" onSubmit={onSubmit}>
                        {children}
                    </form>
				</div>
			</main>
		</div>
	);
};

export default VerifiContainer;