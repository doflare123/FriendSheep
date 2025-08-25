import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import styles from '../../styles/Events/MapModal.module.css';
import { yandexMapsAPI, PlaceInfo } from '../../lib/api';

interface MapModalProps {
  isOpen: boolean;
  onClose: () => void;
  onPlaceSelected: (place: PlaceInfo) => void;
}

export default function MapModal({ isOpen, onClose, onPlaceSelected }: MapModalProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<PlaceInfo[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [selectedPlace, setSelectedPlace] = useState<PlaceInfo | null>(null);
  
  useEffect(() => {
    if (!isOpen) return;

    // Инициализируем карту при открытии модального окна
    const initMap = async () => {
      try {
        await yandexMapsAPI.loadMaps();
        
        // Небольшая задержка для корректного рендера DOM
        setTimeout(() => {
          yandexMapsAPI.openPlacePicker('yandex-map', (place) => {
            // Сохраняем выбранное место, но не закрываем модал сразу
            setSelectedPlace(place);
          });
        }, 100);
      } catch (error) {
        console.error('Ошибка загрузки карты:', error);
        alert('Ошибка загрузки карты. Проверьте настройки API.');
      }
    };
    
    initMap();

    // Очищаем состояние при закрытии модала
    return () => {
      if (!isOpen) {
        setSelectedPlace(null);
        setSearchResults([]);
        setSearchQuery('');
      }
    };
  }, [isOpen]);

  if (!isOpen) return null;

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;
    
    setIsSearching(true);
    try {
      const results = await yandexMapsAPI.searchPlaces(searchQuery);
      setSearchResults(results);
    } catch (error) {
      console.error('Ошибка поиска места:', error);
      alert('Ошибка поиска места');
    } finally {
      setIsSearching(false);
    }
  };

  const handlePlaceClick = async (place: PlaceInfo) => {
    // Переходим к месту на карте и сохраняем его как выбранное
    await yandexMapsAPI.goToPlace(place);
    setSelectedPlace(place);
    // Очищаем результаты поиска для лучшего UX
    setSearchResults([]);
    setSearchQuery('');
  };

  const handleConfirmSelection = () => {
    if (selectedPlace) {
      onPlaceSelected(selectedPlace);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  return (
    <div className={styles.mapModalOverlay}>
      <div className={styles.mapModal}>
        <div className={styles.mapModalHeader}>
          <h3>Выберите место на карте</h3>
          <button className={styles.closeButton} onClick={onClose}>
            <Image src="/icons/close.png" alt="Закрыть" width={24} height={24} />
          </button>
        </div>

        {searchResults.length > 0 && (
          <div className={styles.searchResults}>
            {searchResults.map((place, index) => (
              <div
                key={index}
                className={styles.searchResultItem}
                onClick={() => handlePlaceClick(place)}
              >
                <strong>{place.name || 'Место'}</strong>
                <span>{place.address}</span>
              </div>
            ))}
          </div>
        )}

        {selectedPlace && (
          <div className={styles.selectedPlaceInfo}>
            <div className={styles.selectedPlaceContent}>
              <strong>Выбранное место:</strong>
              <div>{selectedPlace.name || 'Выбранное место'}</div>
              <div className={styles.selectedAddress}>{selectedPlace.address}</div>
            </div>
            <button 
              className={styles.confirmButton}
              onClick={handleConfirmSelection}
            >
              Выбрать это место
            </button>
          </div>
        )}
        
        <div id="yandex-map" className={styles.mapContainer}></div>
        
        <div className={styles.mapModalFooter}>
          <p>Нажмите на карту или перетащите маркер для выбора места, затем подтвердите выбор</p>
        </div>
      </div>
    </div>
  );
}