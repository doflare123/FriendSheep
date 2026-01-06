import React, { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import SocialContactsModal from './SocialContactsModal';
import DeleteConfirmModal from './DeleteConfirmModal';
import ImageCropModal from '@/components/ImageCropModal';
import SocialIcon from '@/components/SocialIcon';
import { GroupData } from '../../types/Groups';
import styles from '../../styles/Groups/CreateGroupModal.module.css';
import { getCategoryIcon } from '@/Constants';
import {filterProfanity} from '@/utils'

interface CreateGroupFormProps {
  onSubmit: (groupData: any) => void;
  onDelete?: () => void;
  initialData?: Partial<GroupData & { shortDescription?: string; isPrivate?: boolean; imagePreview?: string; socialContacts?: { name: string; link: string }[] }>;
  showTitle?: boolean;
  isLoading?: boolean;
  isEditMode?: boolean;
  groupName?: string;
}

const CreateGroupForm: React.FC<CreateGroupFormProps> = ({ 
  onSubmit, 
  onDelete,
  initialData,
  showTitle = true,
  isLoading = false,
  isEditMode = false,
  groupName = ''
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
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [deleteStep, setDeleteStep] = useState<1 | 2>(1);
  const [showCropModal, setShowCropModal] = useState(false);
  const [tempImageFile, setTempImageFile] = useState<File | null>(null);
  const [errors, setErrors] = useState({
    name: '',
    shortDescription: '',
    description: '',
    categories: ''
  });
  const fileInputRef = useRef<HTMLInputElement>(null);

  const VALIDATION_RULES = {
    name: { min: 5, max: 40 },
    shortDescription: { min: 5, max: 50 },
    description: { min: 5, max: 300 }
  };

  const categories = [
    { id: 'movies', name: 'Медиа', icon: getCategoryIcon("movies") },
    { id: 'games', name: 'Видеоигры', icon: getCategoryIcon("games") },
    { id: 'board', name: 'Настолки', icon: getCategoryIcon("boards") },
    { id: 'other', name: 'Другое', icon: getCategoryIcon("other") }
  ];

  useEffect(() => {
    console.log('initialData changed:', {
      imagePreview: initialData?.imagePreview,
      selectedImage: selectedImage,
      selectedImageFile: selectedImageFile
    });
    if (initialData) {
      const newCategories = Array.isArray(initialData.categories) ? [...initialData.categories] : [];
      
      setFormData({
        name: initialData.name || '',
        shortDescription: initialData.shortDescription || initialData.small_description || '',
        description: initialData.description || '',
        city: initialData.city || '',
        isPrivate: initialData.isPrivate || initialData.private || false,
        categories: newCategories,
        socialContacts: Array.isArray(initialData.socialContacts) ? [...initialData.socialContacts] : []
      });
      
      if (initialData.imagePreview) {
        setSelectedImage(initialData.imagePreview);
      }
    }
  }, [
    initialData?.name,
    initialData?.shortDescription,
    initialData?.small_description,
    initialData?.description,
    initialData?.city,
    initialData?.isPrivate,
    initialData?.private,
    initialData?.imagePreview,
    JSON.stringify(initialData?.categories),
    JSON.stringify(initialData?.socialContacts)
  ]);

  const validateForm = () => {
    const newErrors = {
      name: '',
      shortDescription: '',
      description: '',
      categories: ''
    };

    if (!formData.name.trim()) {
      newErrors.name = 'Название группы обязательно';
    } else if (formData.name.trim().length < VALIDATION_RULES.name.min) {
      newErrors.name = `Название должно содержать минимум ${VALIDATION_RULES.name.min} символов`;
    } else if (formData.name.trim().length > VALIDATION_RULES.name.max) {
      newErrors.name = `Название не должно превышать ${VALIDATION_RULES.name.max} символов`;
    }

    if (!formData.shortDescription.trim()) {
      newErrors.shortDescription = 'Краткое описание обязательно';
    } else if (formData.shortDescription.trim().length < VALIDATION_RULES.shortDescription.min) {
      newErrors.shortDescription = `Минимум ${VALIDATION_RULES.shortDescription.min} символов`;
    } else if (formData.shortDescription.trim().length > VALIDATION_RULES.shortDescription.max) {
      newErrors.shortDescription = `Максимум ${VALIDATION_RULES.shortDescription.max} символов`;
    }

    if (!formData.description.trim()) {
      newErrors.description = 'Описание группы обязательно';
    } else if (formData.description.trim().length < VALIDATION_RULES.description.min) {
      newErrors.description = `Описание должно содержать минимум ${VALIDATION_RULES.description.min} символов`;
    } else if (formData.description.trim().length > VALIDATION_RULES.description.max) {
      newErrors.description = `Описание не должно превышать ${VALIDATION_RULES.description.max} символов`;
    }

    if (formData.categories.length === 0) {
      newErrors.categories = 'Выберите хотя бы одну категорию';
    }

    setErrors(newErrors);
    return !newErrors.name && !newErrors.shortDescription && !newErrors.description && !newErrors.categories;
  };

  const handleInputChange = (field: string, value: any) => {
    let processedValue = value;
    
    if (field === 'name' && typeof value === 'string') {
      processedValue = value.slice(0, VALIDATION_RULES.name.max);
    } else if (field === 'shortDescription' && typeof value === 'string') {
      processedValue = value.slice(0, VALIDATION_RULES.shortDescription.max);
    } else if (field === 'description' && typeof value === 'string') {
      processedValue = value.slice(0, VALIDATION_RULES.description.max);
    }

    setFormData(prev => ({
      ...prev,
      [field]: processedValue
    }));

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

    if (errors.categories) {
      setErrors(prev => ({ ...prev, categories: '' }));
    }
  };

  const handleImageUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setTempImageFile(file);
      setShowCropModal(true);
    }
  };

  const handleCropSave = (blob: Blob) => {
    const file = new File([blob], 'group-image.jpg', { type: 'image/jpeg' });
    setSelectedImageFile(file);
    
    const reader = new FileReader();
    reader.onload = (e) => {
      setSelectedImage(e.target?.result as string);
    };
    reader.readAsDataURL(file);
    
    setShowCropModal(false);
    setTempImageFile(null);
  };

  const handleCropCancel = () => {
    setShowCropModal(false);
    setTempImageFile(null);
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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    const filteredSocialContacts = formData.socialContacts.map(contact => ({
      name: filterProfanity(contact.name),
      link: contact.link
    }));

    onSubmit({
      name: filterProfanity(formData.name),
      shortDescription: filterProfanity(formData.shortDescription),
      description: filterProfanity(formData.description),
      city: filterProfanity(formData.city),
      isPrivate: formData.isPrivate,
      categories: formData.categories,
      socialContacts: filteredSocialContacts,
      image: selectedImageFile,
      imagePreview: selectedImage
    });
  };

  const handleDeleteClick = () => {
    setDeleteStep(1);
    setIsDeleteModalOpen(true);
  };

  const handleDeleteConfirm = () => {
    if (deleteStep === 1) {
      setDeleteStep(2);
    } else {
      setIsDeleteModalOpen(false);
      if (onDelete) {
        onDelete();
      }
    }
  };

  const handleDeleteModalClose = () => {
    setIsDeleteModalOpen(false);
    setDeleteStep(1);
  };

  const getCharacterCount = (field: 'name' | 'shortDescription' | 'description') => {
    const current = formData[field].length;
    const max = VALIDATION_RULES[field].max;
    return `${current}/${max}`;
  };

  return (
  <>
    {showTitle && (
      <div className={styles.formTitleSection}>
        <h2 className={styles.formTitle}>
          Основная информация
        </h2>
      </div>
    )}
    
    <form onSubmit={handleSubmit} className={styles.createGroupForm}>
        <div className={styles.formRow}>
          <div className={styles.leftColumn}>
            <div className={styles.formGroup}>
              <input
                type="text"
                className={`${styles.nameInput} ${errors.name ? styles.error : ''}`}
                placeholder="Название *"
                value={formData.name}
                onChange={(e) => handleInputChange('name', e.target.value)}
                maxLength={VALIDATION_RULES.name.max}
              />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                {errors.name && <span className={styles.errorMessage}>{errors.name}</span>}
                <span style={{ 
                  fontSize: '12px', 
                  color: 'var(--color-text-muted)',
                  marginLeft: 'auto',
                  marginTop: '4px'
                }}>
                  {getCharacterCount('name')}
                </span>
              </div>
            </div>

            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Краткое описание *</label>
              <input
                type="text"
                className={`${styles.formInput} ${errors.shortDescription ? styles.error : ''}`}
                value={formData.shortDescription}
                onChange={(e) => handleInputChange('shortDescription', e.target.value)}
                maxLength={VALIDATION_RULES.shortDescription.max}
              />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                {errors.shortDescription && <span className={styles.errorMessage}>{errors.shortDescription}</span>}
                <span style={{ 
                  fontSize: '12px', 
                  color: 'var(--color-text-muted)',
                  marginLeft: 'auto',
                  marginTop: '4px'
                }}>
                  {getCharacterCount('shortDescription')}
                </span>
              </div>
            </div>

            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Описание *</label>
              <textarea
                className={`${styles.formTextarea} ${errors.description ? styles.error : ''}`}
                value={formData.description}
                onChange={(e) => handleInputChange('description', e.target.value)}
                rows={4}
                maxLength={VALIDATION_RULES.description.max}
              />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                {errors.description && <span className={styles.errorMessage}>{errors.description}</span>}
                <span style={{ 
                  fontSize: '12px', 
                  color: 'var(--color-text-muted)',
                  marginLeft: 'auto',
                  marginTop: '4px'
                }}>
                  {getCharacterCount('description')}
                </span>
              </div>
            </div>

            <div className={styles.formGroup}>
              <label className={styles.formLabel}>Контакты</label>
              <div className={styles.socialIcons}>
                {formData.socialContacts.map((contact, index) => (
                  <div key={index} className={styles.socialIconWrapper}>
                    <SocialIcon
                      href={contact.link}
                      alt={contact.name}
                      size={50}
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

            <div className={styles.formGroup}>
              <input
                type="text"
                className={styles.cityInput}
                placeholder="Город"
                value={formData.city}
                onChange={(e) => handleInputChange('city', e.target.value)}
              />
            </div>

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

        <div className={styles.actionButtons}>
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
          
          {isEditMode && onDelete && (
            <button 
              type="button"
              className={styles.deleteButton}
              onClick={handleDeleteClick}
              disabled={isLoading}
            >
              Удалить
            </button>
          )}
        </div>
      </form>

      <SocialContactsModal
        isOpen={isSocialModalOpen}
        onClose={() => setIsSocialModalOpen(false)}
        onSave={handleSocialContactsSave}
        initialContacts={formData.socialContacts}
      />

      <DeleteConfirmModal
        isOpen={isDeleteModalOpen}
        onClose={handleDeleteModalClose}
        onConfirm={handleDeleteConfirm}
        step={deleteStep}
        groupName={groupName || formData.name}
      />

      {showCropModal && tempImageFile && (
        <ImageCropModal
          imageFile={tempImageFile}
          onSave={handleCropSave}
          onCancel={handleCropCancel}
          title="Настройте изображение группы"
          cropShape="square"
          finalSize={500}
        />
      )}
    </>
  );
};

export default CreateGroupForm;