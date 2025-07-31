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
  color = "#37A2E6",
  onClick 
}) => {
  const linkStyle = color !== "#37A2E6" ? { color } : {};
  
  return (
    <a 
      href={href} 
      className={`${className} ${color === "#37A2E6" ? "text-[#37A2E6]" : ""}`}
      style={color !== "#37A2E6" ? linkStyle : {}}
      onClick={onClick}
    >
      {children}
    </a>
  );
};

export default FormLink;