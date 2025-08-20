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
    onPlaceSelected: (place: PlaceInfo) => void
  ): Promise<void> {
    await this.loadMaps();

    const map = new window.ymaps.Map(containerId, {
      center: [55.76, 37.64], // Москва по умолчанию
      zoom: 10,
    });

    const placemark = new window.ymaps.Placemark([55.76, 37.64], {}, {
      draggable: true,
    });

    map.geoObjects.add(placemark);

    placemark.events.add('dragend', () => {
      const coords = placemark.geometry.getCoordinates();
      
      // Получаем адрес по координатам
      window.ymaps.geocode(coords).then((res: any) => {
        const firstGeoObject = res.geoObjects.get(0);
        const address = firstGeoObject.getAddressLine();
        
        onPlaceSelected({
          address,
          coordinates: coords,
          name: firstGeoObject.getMetaDataProperty('name'),
        });
      });
    });

    // Поиск при клике на карту
    map.events.add('click', (e: any) => {
      const coords = e.get('coords');
      placemark.geometry.setCoordinates(coords);
      
      window.ymaps.geocode(coords).then((res: any) => {
        const firstGeoObject = res.geoObjects.get(0);
        const address = firstGeoObject.getAddressLine();
        
        onPlaceSelected({
          address,
          coordinates: coords,
          name: firstGeoObject.getMetaDataProperty('name'),
        });
      });
    });
  }

  async searchPlaces(query: string): Promise<PlaceInfo[]> {
    await this.loadMaps();

    return new Promise((resolve, reject) => {
      window.ymaps.geocode(query, {
        results: 10,
      }).then((res: any) => {
        const places: PlaceInfo[] = [];
        
        res.geoObjects.each((geoObject: any) => {
          places.push({
            address: geoObject.getAddressLine(),
            coordinates: geoObject.geometry.getCoordinates(),
            name: geoObject.getMetaDataProperty('name'),
          });
        });
        
        resolve(places);
      }).catch(reject);
    });
  }
}