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
  
  useEffect(() => {
    if (!isOpen) return;

    // Инициализируем карту при открытии модального окна
    const initMap = async () => {
      try {
        await yandexMapsAPI.loadMaps();
        
        // Небольшая задержка для корректного рендера DOM
        setTimeout(() => {
          yandexMapsAPI.openPlacePicker('yandex-map', onPlaceSelected);
        }, 100);
      } catch (error) {
        console.error('Ошибка загрузки карты:', error);
        alert('Ошибка загрузки карты. Проверьте настройки API.');
      }
    };
    
    initMap();
  }, [isOpen, onPlaceSelected]);

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

  const handlePlaceClick = (place: PlaceInfo) => {
    onPlaceSelected(place);
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
        
        <div className={styles.mapSearch}>
          <input
            type="text"
            placeholder="Поиск места..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className={styles.searchInput}
            onKeyPress={handleKeyPress}
          />
          <button 
            onClick={handleSearch} 
            disabled={isSearching}
            className={styles.searchButton}
          >
            {isSearching ? '...' : 'Найти'}
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
        
        <div id="yandex-map" className={styles.mapContainer}></div>
        
        <div className={styles.mapModalFooter}>
          <p>Нажмите на карту или перетащите маркер для выбора места</p>
        </div>
      </div>
    </div>
  );
}