import React, { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import SocialContactsModal from './SocialContactsModal';
import { GroupFormData } from '../../types/AdminTypes';
import styles from '../../styles/Groups/CreateGroupModal.module.css';

interface CreateGroupFormProps {
  onSubmit: (groupData: any) => void;
  initialData?: Partial<GroupFormData>;
  showTitle?: boolean;
  isLoading?: boolean;
}

const CreateGroupForm: React.FC<CreateGroupFormProps> = ({ 
  onSubmit, 
  initialData,
  showTitle = true,
  isLoading = false
}) => {
  const [formData, setFormData] = useState({
    name: initialData?.name || '',
    shortDescription: initialData?.shortDescription || '',
    description: initialData?.description || '',
    city: initialData?.city || '',
    isPrivate: initialData?.isPrivate || false,
    categories: initialData?.categories || [] as string[],
    socialContacts: initialData?.socialContacts || [] as { name: string; link: string }[]
  });

  const [selectedImage, setSelectedImage] = useState<string>(
    initialData?.imagePreview || '/default/group.jpg'
  );
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null);
  const [isSocialModalOpen, setIsSocialModalOpen] = useState(false);
  const [errors, setErrors] = useState({
    name: '',
    shortDescription: '',
    description: '',
    categories: ''
  });
  const fileInputRef = useRef<HTMLInputElement>(null);

  const categories = [
    { id: 'movies', name: 'Фильмы', icon: '/events/movies.png' },
    { id: 'games', name: 'Игры', icon: '/events/games.png' },
    { id: 'board', name: 'Настолки', icon: '/events/board.png' },
    { id: 'other', name: 'Другое', icon: '/events/other.png' }
  ];

  const baseSocialNetworks = [
    { name: 'Discord', icon: '/social/ds.png' },
    { name: 'Telegram', icon: '/social/tg.png' },
    { name: 'VKontakte', icon: '/social/vk.png' },
    { name: 'WhatsApp', icon: '/social/wa.png' },
    { name: 'Snapchat', icon: '/social/snap.png' }
  ];

  // Обновляем форму при изменении initialData
  useEffect(() => {
    if (initialData) {
      setFormData({
        name: initialData.name || '',
        shortDescription: initialData.shortDescription || '',
        description: initialData.description || '',
        city: initialData.city || '',
        isPrivate: initialData.isPrivate || false,
        categories: initialData.categories || [],
        socialContacts: initialData.socialContacts || []
      });
      
      if (initialData.imagePreview) {
        setSelectedImage(initialData.imagePreview);
      }
    }
  }, [initialData]);

  const validateForm = () => {
    const newErrors = {
      name: '',
      shortDescription: '',
      description: '',
      categories: ''
    };

    if (!formData.name.trim()) {
      newErrors.name = 'Название группы обязательно';
    }

    if (!formData.shortDescription.trim()) {
      newErrors.shortDescription = 'Краткое описание обязательно';
    }

    if (!formData.description.trim()) {
      newErrors.description = 'Описание группы обязательно';
    }

    if (formData.categories.length === 0) {
      newErrors.categories = 'Выберите хотя бы одну категорию';
    }

    setErrors(newErrors);
    return !newErrors.name && !newErrors.shortDescription && !newErrors.description && !newErrors.categories;
  };

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));

    // Очищаем ошибку для поля при вводе
    if (field === 'name' && errors.name) {
      setErrors(prev => ({ ...prev, name: '' }));
    }
    if (field === 'shortDescription' && errors.shortDescription) {
      setErrors(prev => ({ ...prev, shortDescription: '' }));
    }
    if (field === 'description' && errors.description) {
      setErrors(prev => ({ ...prev, description: '' }));
    }
  };

  const toggleCategory = (categoryId: string) => {
    setFormData(prev => ({
      ...prev,
      categories: prev.categories.includes(categoryId)
        ? prev.categories.filter(id => id !== categoryId)
        : [...prev.categories, categoryId]
    }));

    // Очищаем ошибку категорий при выборе
    if (errors.categories) {
      setErrors(prev => ({ ...prev, categories: '' }));
    }
  };

  const handleImageUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setSelectedImageFile(file);
      
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

  const handleSocialContactsSave = (contacts: { name: string; link: string }[]) => {
    setFormData(prev => ({
      ...prev,
      socialContacts: contacts
    }));
  };

  const handleSocialIconClick = (link: string) => {
    if (link) {
      window.open(link, '_blank');
    }
  };

  const getSocialIcon = (name: string) => {
    const baseSocial = baseSocialNetworks.find(social => social.name === name);
    return baseSocial ? baseSocial.icon : '/default/soc_net.png';
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return; // Не отправляем форму, если есть ошибки валидации
    }

    onSubmit({
      ...formData,
      image: selectedImageFile,
      imagePreview: selectedImage
    });
  };

  return (
    <>
      {showTitle && (
        <div style={{ marginBottom: '30px' }}>
          <h2 style={{ 
            color: '#000', 
            fontSize: '24px', 
            fontWeight: '600', 
            margin: 0 
          }}>
            Основная информация
          </h2>
        </div>
      )}
      
      <form onSubmit={handleSubmit} className={styles.createGroupForm}>
        <div className={styles.formRow}>
          <div className={styles.leftColumn}>
            {/* Название */}
            <div className={styles.formGroup}>
              <input
                type="text"
                className={`${styles.nameInput} ${errors.name ? styles.error : ''}`}
                placeholder="Название *"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
              />
              {errors.name && <span className={styles.errorMessage}>{errors.name}</span>}
            </div>

            {/* Краткое описание */}
            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Краткое описание *</label>
              <input
                type="text"
                className={`${styles.formInput} ${errors.shortDescription ? styles.error : ''}`}
                value={formData.shortDescription}
                onChange={(e) => handleInputChange('shortDescription', e.target.value)}
              />
              {errors.shortDescription && <span className={styles.errorMessage}>{errors.shortDescription}</span>}
            </div>

            {/* Описание */}
            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Описание *</label>
              <textarea
                className={`${styles.formTextarea} ${errors.description ? styles.error : ''}`}
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                rows={4}
              />
              {errors.description && <span className={styles.errorMessage}>{errors.description}</span>}
            </div>

            {/* Контакты */}
            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Контакты</label>
              <div className={styles.socialIcons}>
                {formData.socialContacts.map((contact, index) => (
                  <div
                    key={index}
                    className={styles.socialIcon}
                    title={`${contact.name}: ${contact.link}`}
                    onClick={() => handleSocialIconClick(contact.link)}
                  >
                    <Image
                      src={getSocialIcon(contact.name)}
                      alt={contact.name}
                      width={50}
                      height={50}
                    />
                  </div>
                ))}
                <div
                  className={styles.addSocialIcon}
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

          <div className={styles.rightColumn}>
            {/* Изображение группы */}
            <div className={styles.groupImageSection}>
              <div className={styles.groupImagePreview} onClick={handleImageClick}>
                <Image
                  src={selectedImage}
                  alt="Group preview"
                  width={180}
                  height={180}
                  className={styles.previewImage}
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
            <div className={styles.formGroup}>
              <input
                type="text"
                className={styles.cityInput}
                placeholder="Город"
                value={formData.city}
                onChange={(e) => handleInputChange('city', e.target.value)}
              />
            </div>

            {/* Приватная группа */}
            <div className={styles.formGroup}>
              <label className={styles.checkboxLabel}>
                <input
                  type="checkbox"
                  className={styles.privateCheckbox}
                  checked={formData.isPrivate}
                  onChange={(e) => handleInputChange('isPrivate', e.target.checked)}
                />
                Эта группа приватная?
              </label>
            </div>

            {/* Категории */}
            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Категории: *</label>
              <div className={`${styles.categoriesRow} ${errors.categories ? styles.error : ''}`}>
                {categories.map((category) => (
                  <div
                    key={category.id}
                    className={`${styles.categoryItem} ${formData.categories.includes(category.id) ? styles.selected : ''}`}
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
              {errors.categories && <span className={styles.errorMessage}>{errors.categories}</span>}
            </div>
          </div>
        </div>

        {/* Кнопка создания */}
        <button 
          type="submit" 
          className={styles.createButton}
          disabled={isLoading}
          style={{
            opacity: isLoading ? 0.7 : 1,
            cursor: isLoading ? 'not-allowed' : 'pointer'
          }}
        >
          {isLoading ? 'Сохранение...' : 'Сохранить изменения'}
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