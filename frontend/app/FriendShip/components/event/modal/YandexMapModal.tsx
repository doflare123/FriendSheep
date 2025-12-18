import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { useThemedColors } from '@/hooks/useThemedColors';
import React, { useRef, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Modal,
  Platform,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import { WebView } from 'react-native-webview';
// eslint-disable-next-line import/no-unresolved
import { YANDEX_JS_API_KEY } from '@env';

interface YandexMapModalProps {
  visible: boolean;
  onClose: () => void;
  onSelectAddress: (address: string) => void;
  initialAddress?: string;
}

const YandexMapModal: React.FC<YandexMapModalProps> = ({
  visible,
  onClose,
  onSelectAddress,
  initialAddress = '',
}) => {
  const colors = useThemedColors();
  const [searchQuery, setSearchQuery] = useState(initialAddress);
  const [selectedAddress, setSelectedAddress] = useState(initialAddress);
  const [isSearching, setIsSearching] = useState(false);
  const [isMapLoading, setIsMapLoading] = useState(true);
  const webViewRef = useRef<WebView>(null);

  const handleSearch = () => {
    if (!searchQuery.trim()) return;
    
    console.log('[YandexMapModal] Поиск адреса:', searchQuery);
    setIsSearching(true);
    
    const jsCode = `
      if (typeof searchAddress === 'function') {
        searchAddress("${searchQuery.replace(/"/g, '\\"').replace(/\n/g, ' ')}");
      }
      true;
    `;
    
    webViewRef.current?.injectJavaScript(jsCode);
  };

  const handleMessage = (event: any) => {
    try {
      const data = JSON.parse(event.nativeEvent.data);
      
      if (data.type === 'addressSelected') {
        console.log('[YandexMapModal] Адрес выбран:', data.address);
        setSelectedAddress(data.address);
        setIsSearching(false);
      } else if (data.type === 'searchComplete') {
        setIsSearching(false);
      } else if (data.type === 'searchError') {
        setIsSearching(false);
        Alert.alert('Не найдено', 'Адрес не найден. Попробуйте другой запрос.');
      } else if (data.type === 'mapReady') {
        console.log('[YandexMapModal] Карта готова');
      } else if (data.type === 'error') {
        console.error('[YandexMapModal] Ошибка в WebView:', data.message);
        setIsSearching(false);
      }
    } catch (error) {
      console.error('[YandexMapModal] Ошибка обработки сообщения:', error);
    }
  };

  const handleSelect = () => {
    if (!selectedAddress.trim()) {
      Alert.alert('Ошибка', 'Выберите адрес на карте или найдите через поиск');
      return;
    }
    onSelectAddress(selectedAddress);
    handleClose();
  };

  const handleClose = () => {
    setSearchQuery(initialAddress);
    setSelectedAddress(initialAddress);
    onClose();
  };

  const htmlContent = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <script src="https://api-maps.yandex.ru/2.1/?apikey=${YANDEX_JS_API_KEY}&lang=ru_RU" type="text/javascript"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { width: 100%; height: 100%; overflow: hidden; }
        #map { width: 100%; height: 100%; }
        
        /* Скрываем копирайты и ссылки внизу */
        .ymaps-2-1-79-copyright,
        .ymaps-2-1-79-copyright__content,
        .ymaps-2-1-79-copyright__link,
        .ymaps-2-1-79-copyright__wrap,
        [class*="copyright"],
        [class*="legal"],
        a[href*="yandex"] {
            display: none !important;
            opacity: 0 !important;
            visibility: hidden !important;
        }
    </style>
</head>
<body>
    <div id="map"></div>
    
    <script type="text/javascript">
        let myMap, myPlacemark;
        
        function sendMessage(data) {
            try {
                if (window.ReactNativeWebView) {
                    window.ReactNativeWebView.postMessage(JSON.stringify(data));
                }
            } catch (error) {
                console.error('Send error:', error);
            }
        }
        
        ymaps.ready(init);
        
        function init() {
            try {
                myMap = new ymaps.Map("map", {
                    center: [55.751244, 37.618423],
                    zoom: 12,
                    controls: ['zoomControl', 'geolocationControl']
                });
                
                sendMessage({ type: 'mapReady' });
                
                ymaps.geolocation.get({
                    provider: 'auto',
                    mapStateAutoApply: false
                }).then(function(result) {
                    myMap.setCenter(result.geoObjects.position, 15);
                }).catch(function() {

                });

                myMap.events.add('click', function (e) {
                    const coords = e.get('coords');
                    setPlacemark(coords);
                    getAddress(coords);
                });
                
            } catch (error) {
                sendMessage({ type: 'error', message: error.toString() });
            }
        }
        
        function setPlacemark(coords) {
            if (myPlacemark) {
                myPlacemark.geometry.setCoordinates(coords);
            } else {
                myPlacemark = new ymaps.Placemark(coords, {}, {
                    preset: 'islands#redIcon',
                    draggable: true
                });
                myMap.geoObjects.add(myPlacemark);
                
                myPlacemark.events.add('dragend', function() {
                    getAddress(myPlacemark.geometry.getCoordinates());
                });
            }
        }
        
        function getAddress(coords) {
            ymaps.geocode(coords, { kind: 'house' })
                .then(function(res) {
                    const obj = res.geoObjects.get(0);
                    if (obj) {
                        sendMessage({
                            type: 'addressSelected',
                            address: obj.getAddressLine(),
                            coords: coords
                        });
                    } else {
                        sendMessage({ type: 'searchError' });
                    }
                })
                .catch(function(err) {
                    sendMessage({ type: 'error', message: 'Geocode: ' + err.toString() });
                    sendMessage({ type: 'searchError' });
                });
        }
        
        function searchAddress(address) {
            ymaps.geocode(address, { results: 1 })
                .then(function(res) {
                    const obj = res.geoObjects.get(0);
                    if (obj) {
                        const coords = obj.geometry.getCoordinates();
                        myMap.setCenter(coords, 15);
                        setPlacemark(coords);
                        
                        sendMessage({
                            type: 'addressSelected',
                            address: obj.getAddressLine(),
                            coords: coords
                        });
                        sendMessage({ type: 'searchComplete' });
                    } else {
                        sendMessage({ type: 'searchError' });
                    }
                })
                .catch(function(err) {
                    sendMessage({ type: 'error', message: 'Search: ' + err.toString() });
                    sendMessage({ type: 'searchError' });
                });
        }
    </script>
</body>
</html>
  `;

  return (
    <Modal visible={visible} animationType="slide" transparent={false}>
      <View style={[styles.container, {backgroundColor: colors.white}]}>
        <View style={[styles.header, {borderBottomColor: colors.lightGrey,}]}>
          <Text style={[styles.headerTitle, {color: colors.black}]}>Выбор места</Text>
          <TouchableOpacity onPress={handleClose} style={styles.closeButton}>
            <Text style={[styles.closeButtonText, {color: colors.grey}]}>✕</Text>
          </TouchableOpacity>
        </View>

        <View style={styles.searchContainer}>
          <TextInput
            style={[styles.searchInput, {borderColor: colors.lightGrey, color: colors.black}]}
            placeholder="Введите адрес"
            placeholderTextColor={colors.grey}
            value={searchQuery}
            onChangeText={setSearchQuery}
            onSubmitEditing={handleSearch}
            returnKeyType="search"
          />
          <TouchableOpacity
            style={[styles.searchButton, {backgroundColor: colors.lightBlue}]}
            onPress={handleSearch}
            disabled={isSearching || !searchQuery.trim()}
          >
            {isSearching ? (
              <ActivityIndicator size="small" color={colors.white} />
            ) : (
              <Text style={styles.searchButtonText}>Найти</Text>
            )}
          </TouchableOpacity>
        </View>

        {selectedAddress !== '' && (
          <View style={styles.selectedAddressContainer}>
            <Text style={[styles.selectedAddressLabel, {color: colors.grey}]}>Выбранный адрес:</Text>
            <Text style={[styles.selectedAddressText, {color: colors.black}]} numberOfLines={2}>
              {selectedAddress}
            </Text>
          </View>
        )}

        <View style={styles.mapContainer}>
          {isMapLoading && (
            <View style={[styles.loadingOverlay, {backgroundColor: colors.white}]}>
              <ActivityIndicator size="large" color={colors.lightBlue} />
              <Text style={[styles.loadingText, {color: colors.grey}]}>Загрузка карты...</Text>
            </View>
          )}
          <WebView
            ref={webViewRef}
            source={{ html: htmlContent }}
            style={styles.map}
            onMessage={handleMessage}
            onLoadEnd={() => setIsMapLoading(false)}
            javaScriptEnabled={true}
            domStorageEnabled={true}
            originWhitelist={['*']}
            mixedContentMode="always"
          />
        </View>

        <View style={[styles.hintContainer, {backgroundColor: colors.white}]}>
          <Text style={[styles.hintText, {color: colors.grey}]}>
            Нажмите на карту или перетащите метку для выбора адреса
          </Text>
        </View>

        <View style={[styles.footer, {borderTopColor: colors.lightGrey}]}>
          <TouchableOpacity
            style={[
              styles.selectButton, {backgroundColor: colors.lightBlue},
              !selectedAddress.trim() && {backgroundColor: colors.lightGrey},
            ]}
            onPress={handleSelect}
            disabled={!selectedAddress.trim()}
          >
            <Text style={styles.selectButtonText}>Выбрать адрес</Text>
          </TouchableOpacity>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    paddingHorizontal: 20,
    paddingTop: Platform.OS === 'ios' ? 50 : 16,
    borderBottomWidth: 1,
    position: 'relative',
  },
  headerTitle: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
  },
  closeButton: {
    position: 'absolute',
    right: 20,
    top: 4,
    padding: 5,
  },
  closeButtonText: {
    fontSize: 28,
    fontFamily: Montserrat.regular,
  },
  searchContainer: {
    flexDirection: 'row',
    padding: 12,
    gap: 10,
  },
  searchInput: {
    flex: 1,
    borderWidth: 1,
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 8,
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  searchButton: {
    borderRadius: 20,
    paddingHorizontal: 20,
    justifyContent: 'center',
    alignItems: 'center',
    minWidth: 80,
  },
  searchButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 14,
    color: Colors.white,
  },
  selectedAddressContainer: {
    paddingHorizontal: 16,
    paddingBottom: 12,
  },
  selectedAddressLabel: {
    fontFamily: Montserrat.bold,
    fontSize: 12,
    marginBottom: 4,
  },
  selectedAddressText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  mapContainer: {
    flex: 1,
    position: 'relative',
  },
  map: {
    flex: 1,
  },
  loadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 10,
  },
  loadingText: {
    marginTop: 16,
    fontFamily: Montserrat.regular,
    fontSize: 14,
  },
  hintContainer: {
    padding: 12,
  },
  hintText: {
    fontFamily: Montserrat.regular,
    fontSize: 10,
    textAlign: 'center',
  },
  footer: {
    padding: 12,
    borderTopWidth: 1,
  },
  selectButton: {
    borderRadius: 20,
    paddingVertical: 8,
    alignItems: 'center',
  },
  selectButtonDisabled: {

  },
  selectButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
});

export default YandexMapModal;