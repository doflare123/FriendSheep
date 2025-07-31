import React, { useState, useRef } from 'react';
import Image from 'next/image';
import SocialContactsModal from './SocialContactsModal';
import '../../styles/Groups/CreateGroupModal.css';

interface CreateGroupFormProps {
  onSubmit: (groupData: any) => void;
}

const CreateGroupForm: React.FC<CreateGroupFormProps> = ({ onSubmit }) => {
  const [formData, setFormData] = useState({
    name: '',
    shortDescription: '',
    description: '',
    city: '',
    isPrivate: false,
    categories: [] as string[],
    socialContacts: {}
  });

  const [selectedImage, setSelectedImage] = useState<string>('/default/group.jpg');
  const [isSocialModalOpen, setIsSocialModalOpen] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const categories = [
    { id: 'movies', name: 'Фильмы', icon: '/events/movies.png' },
    { id: 'games', name: 'Игры', icon: '/events/games.png' },
    { id: 'board', name: 'Настолки', icon: '/events/board.png' },
    { id: 'other', name: 'Другое', icon: '/events/other.png' }
  ];

  const socialNetworks = [
    { id: 'ds', name: 'Discord', icon: '/social/ds.png' },
    { id: 'tg', name: 'Telegram', icon: '/social/tg.png' },
    { id: 'vk', name: 'VKontakte', icon: '/social/vk.png' },
    { id: 'wa', name: 'WhatsApp', icon: '/social/wa.png' },
    { id: 'snap', name: 'Snapchat', icon: '/social/snap.png' }
  ];

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const toggleCategory = (categoryId: string) => {
    setFormData(prev => ({
      ...prev,
      categories: prev.categories.includes(categoryId)
        ? prev.categories.filter(id => id !== categoryId)
        : [...prev.categories, categoryId]
    }));
  };

  const handleImageUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setSelectedImage(e.target?.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleImageClick = () => {
    fileInputRef.current?.click();
  };

  const handleSocialContactsSave = (contacts: any) => {
    setFormData(prev => ({
      ...prev,
      socialContacts: contacts
    }));
  };

  const handleSocialIconClick = (socialId: string) => {
    const contact = formData.socialContacts[socialId];
    if (contact?.link) {
      window.open(contact.link, '_blank');
    }
  };

  const getActiveSocialNetworks = () => {
    const activeSocials = [];
    
    // Добавляем стандартные соц сети
    for (const social of socialNetworks) {
      if (formData.socialContacts[social.id]?.link) {
        activeSocials.push(social);
      }
    }
    
    // Добавляем кастомные соц сети
    for (const [key, contact] of Object.entries(formData.socialContacts)) {
      if (key.startsWith('custom_') && (contact as any)?.link) {
        activeSocials.push({
          id: key,
          name: 'Custom',
          icon: '/default/soc_net.png'
        });
      }
    }
    
    return activeSocials;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      ...formData,
      image: selectedImage
    });
  };

  return (
    <>
      <form onSubmit={handleSubmit} className="createGroupForm">
        <div className="formRow">
          <div className="leftColumn">
            {/* Название */}
            <div className="formGroup">
              <input
                type="text"
                className="nameInput"
                placeholder="Название"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                required
              />
            </div>

            {/* Краткое описание */}
            <div className="formGroup">
              <label className="formLabel">Краткое описание (опционально)</label>
              <input
                type="text"
                className="formInput"
                value={formData.shortDescription}
                onChange={(e) => handleInputChange('shortDescription', e.target.value)}
              />
            </div>

            {/* Описание */}
            <div className="formGroup">
              <label className="formLabel">Описание</label>
              <textarea
                className="formTextarea"
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                rows={4}
              />
            </div>

            {/* Контакты */}
            <div className="formGroup">
              <label className="formLabel">Контакты</label>
              <div className="socialIcons">
                {getActiveSocialNetworks().map((social) => (
                  <div
                    key={social.id}
                    className="socialIcon"
                    title={formData.socialContacts[social.id]?.description || social.name}
                    onClick={() => handleSocialIconClick(social.id)}
                  >
                    <Image
                      src={social.icon}
                      alt={social.name}
                      width={50}
                      height={50}
                    />
                  </div>
                ))}
                <div
                  className="addSocialIcon"
                  onClick={() => setIsSocialModalOpen(true)}
                >
                  <Image
                    src="/add_button.png"
                    alt="Add social"
                    width={50}
                    height={50}
                  />
                </div>
              </div>
            </div>
          </div>

          <div className="rightColumn">
            {/* Изображение группы */}
            <div className="groupImageSection">
              <div className="groupImagePreview" onClick={handleImageClick}>
                <Image
                  src={selectedImage}
                  alt="Group preview"
                  width={180}
                  height={180}
                  className="previewImage"
                />
                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleImageUpload}
                  accept="image/*"
                  style={{ display: 'none' }}
                />
              </div>
            </div>

            {/* Город */}
            <div className="formGroup">
              <input
                type="text"
                className="cityInput"
                placeholder="Город"
                value={formData.city}
                onChange={(e) => handleInputChange('city', e.target.value)}
              />
            </div>

            {/* Приватная группа */}
            <div className="formGroup">
              <label className="checkboxLabel">
                <input
                  type="checkbox"
                  className="privateCheckbox"
                  checked={formData.isPrivate}
                  onChange={(e) => handleInputChange('isPrivate', e.target.checked)}
                />
                Эта группа приватная?
              </label>
            </div>

            {/* Категории */}
            <div className="formGroup">
              <label className="formLabel">Категории:</label>
              <div className="categoriesRow">
                {categories.map((category) => (
                  <div
                    key={category.id}
                    className={`categoryItem ${formData.categories.includes(category.id) ? 'selected' : ''}`}
                    onClick={() => toggleCategory(category.id)}
                    title={category.name}
                  >
                    <Image
                      src={category.icon}
                      alt={category.name}
                      width={20}
                      height={20}
                    />
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Кнопка создания */}
        <button type="submit" className="createButton">
          Создать группу
        </button>
      </form>

      <SocialContactsModal
        isOpen={isSocialModalOpen}
        onClose={() => setIsSocialModalOpen(false)}
        onSave={handleSocialContactsSave}
        initialContacts={formData.socialContacts}
      />
    </>
  );
};

export default CreateGroupForm;