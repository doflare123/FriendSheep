import React from 'react';

interface FormLinkProps {
  href?: string;
  children: React.ReactNode;
  className?: string;
  color?: string;
  onClick?: (e: React.MouseEvent<HTMLAnchorElement>) => void;
}

const FormLink: React.FC<FormLinkProps> = ({ 
  href = "#", 
  children, 
  className = "hover:underline",
  color = "var(--color-primary-blue)",
  onClick 
}) => {
  const linkStyle = color !== "var(--color-primary-blue)" ? { color } : {};
  
  return (
    <a 
      href={href} 
      className={`${className} ${color === "var(--color-primary-blue)" ? "text-[var(--color-primary-blue)]" : ""}`}
      style={color !== "var(--color-primary-blue)" ? linkStyle : {}}
      onClick={onClick}
    >
      {children}
    </a>
  );
};

export default FormLink;