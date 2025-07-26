import React, { useState } from 'react';
import Image from 'next/image';
import '../../styles/Groups/CreateGroupModal.css';

interface CreateGroupModalProps {
  onClose: () => void;
  onSubmit: (groupData: any) => void;
}

const CreateGroupModal: React.FC<CreateGroupModalProps> = ({ onClose, onSubmit }) => {
  const [formData, setFormData] = useState({
    name: '',
    shortDescription: '',
    description: '',
    city: '',
    isPrivate: false,
    categories: [] as string[],
    socialLinks: {
      ds: '',
      tg: '',
      vk: ''
    }
  });

  const [selectedImage, setSelectedImage] = useState<string>('/group-default.png');

  const categories = [
    { id: 'movies', name: 'Фильмы', icon: '/events/movies.png' },
    { id: 'games', name: 'Игры', icon: '/events/games.png' },
    { id: 'board', name: 'Настолки', icon: '/events/board.png' },
    { id: 'other', name: 'Другое', icon: '/events/other.png' }
  ];

  const socialNetworks = [
    { id: 'ds', name: 'Discord', icon: '/social/ds.png' },
    { id: 'tg', name: 'Telegram', icon: '/social/tg.png' },
    { id: 'vk', name: 'VKontakte', icon: '/social/vk.png' }
  ];

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSocialLinkChange = (platform: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      socialLinks: {
        ...prev.socialLinks,
        [platform]: value
      }
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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      ...formData,
      image: selectedImage
    });
  };

  return (
    <div className='modalOverlay' onClick={onClose}>
      <div className='modalContent' onClick={(e) => e.stopPropagation()}>
        <div className='modalHeader'>
          <h2>Основная информация</h2>
          <button className='closeButton' onClick={onClose}>×</button>
        </div>

        <form onSubmit={handleSubmit} className='modalForm'>
          <div className='formRow'>
            <div className='leftColumn'>
              {/* Название */}
              <div className='formGroup'>
                <label className='formLabel'>Название</label>
                <input
                  type='text'
                  className='formInput'
                  value={formData.name}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  required
                />
              </div>

              {/* Краткое описание */}
              <div className='formGroup'>
                <label className='formLabel'>Краткое описание (опционально)</label>
                <input
                  type='text'
                  className='formInput'
                  value={formData.shortDescription}
                  onChange={(e) => handleInputChange('shortDescription', e.target.value)}
                />
              </div>

              {/* Описание */}
              <div className='formGroup'>
                <label className='formLabel'>Описание</label>
                <textarea
                  className='formTextarea'
                  value={formData.description}
                  onChange={(e) => handleInputChange('description', e.target.value)}
                  rows={4}
                />
              </div>

              {/* Контакты */}
              <div className='formGroup'>
                <label className='formLabel'>Контакты</label>
                <div className='socialInputs'>
                  {socialNetworks.map((social) => (
                    <div key={social.id} className='socialInput'>
                      <div className='socialIconInput'>
                        <Image
                          src={social.icon}
                          alt={social.name}
                          width={24}
                          height={24}
                        />
                      </div>
                      <input
                        type='text'
                        placeholder={`Ссылка на ${social.name}`}
                        className='socialLinkInput'
                        value={formData.socialLinks[social.id as keyof typeof formData.socialLinks]}
                        onChange={(e) => handleSocialLinkChange(social.id, e.target.value)}
                      />
                    </div>
                  ))}
                  <div className='addSocialButton'>
                    <div className='addIcon'>+</div>
                  </div>
                </div>
              </div>
            </div>

            <div className='rightColumn'>
              {/* Изображение группы */}
              <div className='groupImageSection'>
                <div className='groupImagePreview'>
                  <Image
                    src={selectedImage}
                    alt='Group preview'
                    width={180}
                    height={180}
                    className='previewImage'
                  />
                </div>
              </div>

              {/* Город */}
              <div className='formGroup'>
                <label className='formLabel'>Город</label>
                <input
                  type='text'
                  className='formInput'
                  value={formData.city}
                  onChange={(e) => handleInputChange('city', e.target.value)}
                />
              </div>

              {/* Приватная группа */}
              <div className='formGroup'>
                <label className='checkboxLabel'>
                  <input
                    type='checkbox'
                    className='checkbox'
                    checked={formData.isPrivate}
                    onChange={(e) => handleInputChange('isPrivate', e.target.checked)}
                  />
                  Эта группа приватная?
                </label>
              </div>

              {/* Категории */}
              <div className='formGroup'>
                <label className='formLabel'>Категории:</label>
                <div className='categoriesGrid'>
                  {categories.map((category) => (
                    <label
                      key={category.id}
                      className='categoryLabel'
                    >
                      <input
                        type='checkbox'
                        className='categoryCheckbox'
                        checked={formData.categories.includes(category.id)}
                        onChange={() => toggleCategory(category.id)}
                      />
                      <div className='categoryContent'>
                        <Image
                          src={category.icon}
                          alt={category.name}
                          width={24}
                          height={24}
                        />
                        <span className='categoryText'>{category.name}</span>
                      </div>
                    </label>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Кнопка создания */}
          <button type='submit' className='createButton'>
            Создать группу
          </button>
        </form>
      </div>
    </div>
  );
};

export default CreateGroupModal;