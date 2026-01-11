import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import Image from 'next/image';
import styles from '../../styles/Groups/SocialContactsModal.module.css';
import { getSocialIcon } from '../../Constants';

interface SocialContactsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (contacts: { name: string; link: string }[]) => void;
  initialContacts: { name: string; link: string }[] | any;
}

interface Contact {
  id: string;
  name: string;
  link: string;
}

const SocialContactsModal: React.FC<SocialContactsModalProps> = ({ 
  isOpen, 
  onClose, 
  onSave, 
  initialContacts 
}) => {
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (isOpen) {
      const contactsArray = Array.isArray(initialContacts) ? initialContacts : [];
      
      if (contactsArray.length > 0) {
        const mappedContacts = contactsArray.map((contact, index) => ({
          id: `contact_${Date.now()}_${index}`,
          name: contact?.name || '',
          link: contact?.link || ''
        }));
        setContacts(mappedContacts);
      } else {
        setContacts([{
          id: `contact_${Date.now()}`,
          name: '',
          link: ''
        }]);
      }
    }
  }, [isOpen, initialContacts]);

  const handleContactChange = (contactId: string, field: 'name' | 'link', value: string) => {
    setContacts(prev => prev.map(contact => 
      contact.id === contactId 
        ? { ...contact, [field]: value }
        : contact
    ));
  };

  const addContact = () => {
    const newContact: Contact = {
      id: `contact_${Date.now()}`,
      name: '',
      link: ''
    };
    setContacts(prev => [...prev, newContact]);
  };

  const removeContact = (contactId: string) => {
    if (contacts.length === 1) {
      setContacts([{
        id: `contact_${Date.now()}`,
        name: '',
        link: ''
      }]);
    } else {
      setContacts(prev => prev.filter(contact => contact.id !== contactId));
    }
  };

  const handleSave = () => {
    const validContacts = contacts
      .filter(contact => contact.link.trim() && contact.name.trim())
      .map(contact => ({
        name: contact.name.trim(),
        link: contact.link.trim()
      }));
    
    onSave(validContacts);
    onClose();
  };

  const handleClose = () => {
    const validContacts = contacts
      .filter(contact => contact.link.trim() && contact.name.trim())
      .map(contact => ({
        name: contact.name.trim(),
        link: contact.link.trim()
      }));
    
    onSave(validContacts);
    onClose();
  };

  if (!isOpen || !mounted) return null;

  const modalContent = (
    <div 
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'var(--overlay-medium)',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        zIndex: 9999
      }}
      onClick={handleClose}
    >
      <div className={styles.socialModalContent} onClick={(e) => e.stopPropagation()}>
        <div className={styles.socialModalHeader}>
          <h3>Контакты группы</h3>
          <button className={styles.closeButton} onClick={handleClose}>×</button>
        </div>
        
        <div className={styles.socialModalBody}>
          {contacts.map((contact) => {
            const iconSrc = getSocialIcon(contact.link);
            const isDiscord = contact.link.toLowerCase().includes('discord') || 
                             contact.name.toLowerCase().includes('discord') || 
                             contact.name.toLowerCase().includes('ds');
            
            return (
              <div key={contact.id} className={styles.socialContactRow}>
                <div 
                  className={styles.socialIconContainer}
                  title={isDiscord ? "*Деятельность организации запрещена на территории РФ" : undefined}
                >
                  <Image
                    src={iconSrc}
                    alt={contact.name || 'Иконка'}
                    width={80}
                    height={80}
                  />
                </div>
                
                <div className={styles.socialContactInputs}>
                  <div className={styles.inputGroup}>
                    <span className={styles.inputLabel}>Имя:</span>
                    <input
                      type="text"
                      className={styles.socialContactInput}
                      value={contact.name}
                      onChange={(e) => handleContactChange(contact.id, 'name', e.target.value)}
                      placeholder="Например: Max"
                    />
                  </div>
                  
                  <div className={styles.inputGroup}>
                    <span className={styles.inputLabel}>Ссылка:</span>
                    <input
                      type="text"
                      className={styles.socialContactInput}
                      value={contact.link}
                      onChange={(e) => handleContactChange(contact.id, 'link', e.target.value)}
                      placeholder="https://..."
                    />
                  </div>
                </div>
                
                <button
                  className={styles.removeContactButton}
                  onClick={() => removeContact(contact.id)}
                  title="Удалить контакт"
                >
                  ×
                </button>
              </div>
            );
          })}
          
          <div className={styles.addCustomSocialRow}>
            <div
              className={styles.addCustomSocialButton}
              onClick={addContact}
            >
              <Image
                src="/add_button.png"
                alt="Добавить контакт"
                width={40}
                height={40}
              />
            </div>
          </div>
        </div>
        
        <button className={styles.saveSocialButton} onClick={handleSave}>
          Сохранить
        </button>
      </div>
    </div>
  );

  return createPortal(modalContent, document.body);
};

export default SocialContactsModal;