import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import styles from '../../styles/Groups/SocialContactsModal.module.css';

interface SocialContactsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (contacts: { name: string; link: string }[]) => void;
  initialContacts: { name: string; link: string }[] | any;
}

interface SocialNetwork {
  id: string;
  name: string;
  icon: string;
  isCustom?: boolean;
}

const SocialContactsModal: React.FC<SocialContactsModalProps> = ({ 
  isOpen, 
  onClose, 
  onSave, 
  initialContacts 
}) => {
  const baseSocialNetworks: SocialNetwork[] = [
    { id: 'discord', name: 'Discord', icon: '/social/ds.png' },
    { id: 'telegram', name: 'Telegram', icon: '/social/tg.png' },
    { id: 'vkontakte', name: 'VKontakte', icon: '/social/vk.png' },
    { id: 'whatsapp', name: 'WhatsApp', icon: '/social/wa.png' },
    { id: 'snapchat', name: 'Snapchat', icon: '/social/snap.png' }
  ];

  const [socialNetworks, setSocialNetworks] = useState<SocialNetwork[]>(baseSocialNetworks);
  const [contacts, setContacts] = useState<Record<string, { name: string; link: string }>>({});

  // Инициализация контактов при открытии модального окна
  useEffect(() => {
    if (isOpen) {
      const contactsMap: Record<string, { name: string; link: string }> = {};
      
      // Проверяем, что initialContacts является массивом
      const contactsArray = Array.isArray(initialContacts) ? initialContacts : [];
      
      // Инициализируем базовые социальные сети
      baseSocialNetworks.forEach(social => {
        const existingContact = contactsArray.find(contact => contact?.name === social.name);
        contactsMap[social.id] = {
          name: social.name,
          link: existingContact?.link || ''
        };
      });

      // Добавляем кастомные социальные сети из initialContacts
      contactsArray.forEach(contact => {
        if (contact && contact.name && contact.link) {
          const isBaseSocial = baseSocialNetworks.some(social => social.name === contact.name);
          if (!isBaseSocial) {
            const customId = `custom_${Date.now()}_${Math.random()}`;
            const customNetwork: SocialNetwork = {
              id: customId,
              name: contact.name,
              icon: '/default/soc_net.png',
              isCustom: true
            };
            setSocialNetworks(prev => [...prev, customNetwork]);
            contactsMap[customId] = {
              name: contact.name,
              link: contact.link
            };
          }
        }
      });

      setContacts(contactsMap);
    }
  }, [isOpen, initialContacts]);

  const handleContactChange = (socialId: string, field: string, value: string) => {
    setContacts(prev => ({
      ...prev,
      [socialId]: {
        ...prev[socialId],
        [field]: value
      }
    }));
  };

  const addCustomSocialNetwork = () => {
    const customId = `custom_${Date.now()}`;
    const newCustomNetwork: SocialNetwork = {
      id: customId,
      name: '',
      icon: '/default/soc_net.png',
      isCustom: true
    };
    
    setSocialNetworks(prev => [...prev, newCustomNetwork]);
    
    setContacts(prev => ({
      ...prev,
      [customId]: {
        name: '',
        link: ''
      }
    }));
  };

  const convertToArray = (contactsData: Record<string, { name: string; link: string }>) => {
    const result: { name: string; link: string }[] = [];
    
    socialNetworks.forEach(social => {
      const contact = contactsData[social.id];
      if (contact && contact.link.trim()) {
        result.push({
          name: contact.name.trim(),
          link: contact.link.trim()
        });
      }
    });
    
    return result;
  };

  const removeEmptyCustomNetworks = () => {
    const networksToKeep = socialNetworks.filter(social => {
      if (!social.isCustom) return true;
      const contact = contacts[social.id];
      return contact && (contact.name.trim() || contact.link.trim());
    });
    setSocialNetworks(networksToKeep);
  };

  const handleSave = () => {
    removeEmptyCustomNetworks();
    const contactsArray = convertToArray(contacts);
    onSave(contactsArray);
    onClose();
  };

  const handleClose = () => {
    removeEmptyCustomNetworks();
    const contactsArray = convertToArray(contacts);
    onSave(contactsArray);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className={styles.modalOverlay} onClick={handleClose}>
      <div className={styles.socialModalContent} onClick={(e) => e.stopPropagation()}>
        <div className={styles.socialModalHeader}>
          <h3>Выберите и заполните контакты</h3>
          <button className={styles.closeButton} onClick={handleClose}>×</button>
        </div>
        <div className={styles.socialModalBody}>
          {socialNetworks.map((social) => (
            <div key={social.id} className={styles.socialContactRow}>
              <div className={styles.socialIconContainer}>
                <Image
                  src={social.icon}
                  alt={social.name}
                  width={80}
                  height={80}
                />
              </div>
              <div className={styles.socialContactInputs}>
                <div className={styles.inputGroup}>
                  <span className={styles.inputLabel}>Имя:</span>
                  {social.isCustom ? (
                    <input
                      type="text"
                      className={styles.socialContactInput}
                      value={contacts[social.id]?.name || ''}
                      onChange={(e) => handleContactChange(social.id, 'name', e.target.value)}
                      placeholder="Введите название"
                    />
                  ) : (
                    <span className={styles.fixedSocialName}>{social.name}</span>
                  )}
                </div>
                <div className={styles.inputGroup}>
                  <span className={styles.inputLabel}>Ссылка:</span>
                  <input
                    type="text"
                    className={styles.socialContactInput}
                    value={contacts[social.id]?.link || ''}
                    onChange={(e) => handleContactChange(social.id, 'link', e.target.value)}
                    placeholder="Введите ссылку"
                  />
                </div>
              </div>
            </div>
          ))}
          
          <div className={styles.addCustomSocialRow}>
            <div
              className={styles.addCustomSocialButton}
              onClick={addCustomSocialNetwork}
            >
              <Image
                src="/add_button.png"
                alt="Добавить кастомную соц. сеть"
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
};

export default SocialContactsModal;