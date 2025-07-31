import React, { useState } from 'react';
import Image from 'next/image';
import '../../styles/Groups/SocialContactsModal.css';

interface SocialContactsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (contacts: any) => void;
  initialContacts: any;
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
  const [contacts, setContacts] = useState(initialContacts);
  
  const baseSocialNetworks: SocialNetwork[] = [
    { id: 'ds', name: 'Discord', icon: '/social/ds.png' },
    { id: 'tg', name: 'Telegram', icon: '/social/tg.png' },
    { id: 'vk', name: 'VKontakte', icon: '/social/vk.png' },
    { id: 'wa', name: 'WhatsApp', icon: '/social/wa.png' },
    { id: 'snap', name: 'Snapchat', icon: '/social/snap.png' }
  ];

  const [socialNetworks, setSocialNetworks] = useState<SocialNetwork[]>(baseSocialNetworks);

  const handleContactChange = (platform: string, field: string, value: string) => {
    setContacts((prev: any) => ({
      ...prev,
      [platform]: {
        ...prev[platform],
        [field]: value
      }
    }));
  };

  const addCustomSocialNetwork = () => {
    const customId = `custom_${Date.now()}`;
    const newCustomNetwork: SocialNetwork = {
      id: customId,
      name: 'Кастомная соц. сеть',
      icon: '/default/soc_net.png',
      isCustom: true
    };
    
    setSocialNetworks(prev => [...prev, newCustomNetwork]);
    
    // Инициализируем пустые поля для новой кастомной соц. сети
    setContacts((prev: any) => ({
      ...prev,
      [customId]: {
        description: '',
        link: ''
      }
    }));
  };

  const filterEmptyCustomContacts = (contactsData: any) => {
    const filtered = { ...contactsData };
    
    Object.keys(filtered).forEach(key => {
      if (key.startsWith('custom_')) {
        const contact = filtered[key];
        if (!contact?.description?.trim() && !contact?.link?.trim()) {
          delete filtered[key];
          // Также удаляем из списка социальных сетей
          setSocialNetworks(prev => prev.filter(social => social.id !== key));
        }
      }
    });
    
    return filtered;
  };

  const handleSave = () => {
    const filteredContacts = filterEmptyCustomContacts(contacts);
    setContacts(filteredContacts);
    onSave(filteredContacts);
    onClose();
  };

  const handleClose = () => {
    const filteredContacts = filterEmptyCustomContacts(contacts);
    setContacts(filteredContacts);
    onSave(filteredContacts);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="modalOverlay" onClick={handleClose}>
      <div className="socialModalContent" onClick={(e) => e.stopPropagation()}>
        <div className="socialModalHeader">
          <h3>Выберите и заполните контакты</h3>
          <button className="closeButton" onClick={handleClose}>×</button>
        </div>
        <div className="socialModalBody">
          {socialNetworks.map((social) => (
            <div key={social.id} className="socialContactRow">
              <div className="socialIconContainer">
                <Image
                  src={social.icon}
                  alt={social.name}
                  width={80}
                  height={80}
                />
              </div>
              <div className="socialContactInputs">
                <div className="inputGroup">
                  <span className="inputLabel">Описание:</span>
                  <input
                    type="text"
                    className="socialContactInput"
                    value={contacts[social.id]?.description || ''}
                    onChange={(e) => handleContactChange(social.id, 'description', e.target.value)}
                  />
                </div>
                <div className="inputGroup">
                  <span className="inputLabel">Ссылка:</span>
                  <input
                    type="text"
                    className="socialContactInput"
                    value={contacts[social.id]?.link || ''}
                    onChange={(e) => handleContactChange(social.id, 'link', e.target.value)}
                  />
                </div>
              </div>
            </div>
          ))}
          
          <div className="addCustomSocialRow">
            <div
              className="addCustomSocialButton"
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
        <button className="saveSocialButton" onClick={handleSave}>
          Сохранить
        </button>
      </div>
    </div>
  );
};

export default SocialContactsModal;