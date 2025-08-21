export interface YandexMapsConfig {
  apiKey: string;
  lang?: string;
}

export interface PlaceInfo {
  address: string;
  coordinates: [number, number];
  name?: string;
}

declare global {
  interface Window {
    ymaps: any;
  }
}

export class YandexMapsAPI {
  private apiKey: string;
  private isLoaded = false;
  private currentMap: any = null;
  private currentPlacemark: any = null;

  constructor(config: YandexMapsConfig) {
    this.apiKey = config.apiKey;
  }

  async loadMaps(): Promise<void> {
    if (this.isLoaded || window.ymaps) {
      return;
    }

    return new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = `https://api-maps.yandex.ru/2.1/?apikey=${this.apiKey}&lang=ru_RU`;
      script.onload = () => {
        window.ymaps.ready(() => {
          this.isLoaded = true;
          resolve();
        });
      };
      script.onerror = reject;
      document.head.appendChild(script);
    });
  }

  async openPlacePicker(
    containerId: string,
    onPlaceSelected: (place: PlaceInfo) => void,
    initialCoords?: [number, number]
  ): Promise<void> {
    await this.loadMaps();

    // Очищаем предыдущую карту если она есть
    this.destroyMap();

    const defaultCoords: [number, number] = initialCoords || [55.76, 37.64];
    
    // Создаем новую карту
    this.currentMap = new window.ymaps.Map(containerId, {
      center: defaultCoords,
      zoom: 13,
      controls: ['zoomControl', 'searchControl', 'typeSelector', 'fullscreenControl']
    });

    // Создаем метку
    this.currentPlacemark = new window.ymaps.Placemark(defaultCoords, {
      hintContent: 'Перетащите метку или кликните на карту',
      balloonContent: 'Выбранное место'
    }, {
      draggable: true,
      iconColor: '#37A2E6'
    });

    this.currentMap.geoObjects.add(this.currentPlacemark);

    // Обработчик перетаскивания метки
    this.currentPlacemark.events.add('dragend', () => {
      const coords = this.currentPlacemark.geometry.getCoordinates();
      this.getAddressFromCoords(coords, onPlaceSelected);
    });

    // Обработчик клика по карте
    this.currentMap.events.add('click', (e: any) => {
      const coords = e.get('coords');
      this.currentPlacemark.geometry.setCoordinates(coords);
      this.getAddressFromCoords(coords, onPlaceSelected);
    });

    // Получаем адрес для начальных координат
    this.getAddressFromCoords(defaultCoords, onPlaceSelected);
  }

  private async getAddressFromCoords(
    coords: [number, number], 
    callback: (place: PlaceInfo) => void
  ): Promise<void> {
    try {
      const res = await window.ymaps.geocode(coords);
      const firstGeoObject = res.geoObjects.get(0);
      
      if (firstGeoObject) {
        // Правильный способ получения названия места
        const addressLine = firstGeoObject.getAddressLine();
        const thoroughfare = firstGeoObject.getThoroughfare();
        const premise = firstGeoObject.getPremise();
        
        // Пытаемся получить название места разными способами
        let placeName = '';
        try {
          // Пробуем получить название организации или POI
          const metaData = firstGeoObject.getMetaData();
          if (metaData && metaData.GeocoderMetaData) {
            const geocoderData = metaData.GeocoderMetaData;
            if (geocoderData.AddressDetails) {
              // Извлекаем название из компонентов адреса
              const components = geocoderData.Address?.Components || [];
              const nameComponent = components.find((comp: any) => 
                comp.kind === 'house' || comp.kind === 'building'
              );
              if (nameComponent) {
                placeName = nameComponent.name;
              }
            }
          }
        } catch (e) {
          // Игнорируем ошибки получения метаданных
        }
        
        // Если не удалось получить конкретное название, формируем его из адреса
        if (!placeName) {
          if (thoroughfare && premise) {
            placeName = `${thoroughfare}, ${premise}`;
          } else if (thoroughfare) {
            placeName = thoroughfare;
          } else {
            placeName = 'Выбранное место';
          }
        }
        
        const place: PlaceInfo = {
          address: addressLine,
          coordinates: coords,
          name: placeName
        };
        
        callback(place);
        
        // Обновляем содержимое балуна
        this.currentPlacemark.properties.set({
          balloonContent: `<strong>${place.name}</strong><br>${place.address}`
        });
      }
    } catch (error) {
      console.error('Ошибка получения адреса:', error);
    }
  }

  async searchPlaces(query: string): Promise<PlaceInfo[]> {
    await this.loadMaps();

    return new Promise((resolve, reject) => {
      window.ymaps.geocode(query, {
        results: 10,
        boundedBy: this.currentMap?.getBounds() // Ограничиваем поиск текущей областью карты
      }).then((res: any) => {
        const places: PlaceInfo[] = [];
        
        res.geoObjects.each((geoObject: any) => {
          // Правильный способ получения названия места
          const addressLine = geoObject.getAddressLine();
          const thoroughfare = geoObject.getThoroughfare();
          const premise = geoObject.getPremise();
          
          let placeName = '';
          try {
            // Пытаемся получить название из метаданных
            const metaData = geoObject.getMetaData();
            if (metaData && metaData.GeocoderMetaData) {
              const geocoderData = metaData.GeocoderMetaData;
              if (geocoderData.AddressDetails) {
                const components = geocoderData.Address?.Components || [];
                const nameComponent = components.find((comp: any) => 
                  comp.kind === 'house' || comp.kind === 'building'
                );
                if (nameComponent) {
                  placeName = nameComponent.name;
                }
              }
            }
          } catch (e) {
            // Игнорируем ошибки
          }
          
          // Если не удалось получить название, формируем из адреса
          if (!placeName) {
            if (thoroughfare && premise) {
              placeName = `${thoroughfare}, ${premise}`;
            } else if (thoroughfare) {
              placeName = thoroughfare;
            } else {
              placeName = addressLine.split(',')[0] || 'Найденное место';
            }
          }
          
          places.push({
            address: addressLine,
            coordinates: geoObject.geometry.getCoordinates(),
            name: placeName
          });
        });
        
        resolve(places);
      }).catch(reject);
    });
  }

  // Метод для перехода к выбранному месту из поиска
  async goToPlace(place: PlaceInfo): Promise<void> {
    if (!this.currentMap || !this.currentPlacemark) return;

    // Перемещаем карту к выбранному месту
    this.currentMap.setCenter(place.coordinates, 15, {
      checkZoomRange: true
    });

    // Перемещаем метку
    this.currentPlacemark.geometry.setCoordinates(place.coordinates);
    
    // Обновляем содержимое балуна
    this.currentPlacemark.properties.set({
      balloonContent: `<strong>${place.name}</strong><br>${place.address}`
    });

    // Открываем балун
    this.currentPlacemark.balloon.open();
  }

  // Метод для очистки ресурсов карты
  destroyMap(): void {
    if (this.currentMap) {
      this.currentMap.destroy();
      this.currentMap = null;
      this.currentPlacemark = null;
    }
  }

  // Метод для получения текущего состояния карты
  getCurrentPlace(): PlaceInfo | null {
    if (!this.currentPlacemark) return null;

    const coords = this.currentPlacemark.geometry.getCoordinates();
    return {
      address: '',
      coordinates: coords,
      name: ''
    };
  }
}