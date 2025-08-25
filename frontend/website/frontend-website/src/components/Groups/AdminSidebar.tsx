import React from 'react';
import { AdminMenuSection } from '../../types/AdminTypes';
import styles from '../../styles/Groups/admin/AdminPage.module.css';

interface AdminSidebarProps {
  sections: AdminMenuSection[];
  activeSection: string;
  onSectionChange: (sectionId: string) => void;
}

const AdminSidebar: React.FC<AdminSidebarProps> = ({
  sections,
  activeSection,
  onSectionChange
}) => {
  return (
    <div className={styles.sidebar}>
      {sections.map((section) => (
        <div
          key={section.id}
          className={`${styles.sidebarItem} ${
            activeSection === section.id ? styles.active : ''
          }`}
          onClick={() => onSectionChange(section.id)}
        >
          {section.title}
        </div>
      ))}
    </div>
  );
};

export default AdminSidebar;