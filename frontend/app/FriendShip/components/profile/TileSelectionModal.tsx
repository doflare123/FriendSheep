import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React, { useEffect, useState } from 'react';
import {
  Image,
  Modal,
  StyleSheet,
  Text,
  TouchableOpacity,
  TouchableWithoutFeedback,
  View,
} from 'react-native';

export type TileType = 'movies' | 'video_games' | 'board_games' | 'other' | 'all' | 'time';

interface TileOption {
  id: TileType;
  title: string;
  icon: any;
}

interface TileSelectionModalProps {
  visible: boolean;
  onClose: () => void;
  selectedTiles: TileType[];
  onTilesChange: (tiles: TileType[]) => void;
  stats: {
    all: number;
    movies: number;
    video_games: number;
    board_games: number;
    other: number;
    time: number;
  };
}

const TileSelectionModal: React.FC<TileSelectionModalProps> = ({
  visible,
  onClose,
  selectedTiles,
  onTilesChange,
  stats,
}) => {
  const [localSelectedTiles, setLocalSelectedTiles] = useState<TileType[]>(selectedTiles);

  useEffect(() => {
    if (visible) {
      setLocalSelectedTiles(selectedTiles);
    }
  }, [visible, selectedTiles]);

  const tileOptions: TileOption[] = [
    { id: 'movies', title: `Медиа - ${stats.movies}`, icon: require('@/assets/images/profile/movies.png') },
    { id: 'video_games', title: `Игры - ${stats.video_games}`, icon: require('@/assets/images/profile/games.png') },
    { id: 'board_games', title: `Настолки - ${stats.board_games}`, icon: require('@/assets/images/profile/table_games.png') },
    { id: 'other', title: `Другое - ${stats.other}`, icon: require('@/assets/images/profile/others.png') },
    { id: 'all', title: `Сессий - ${stats.all}`, icon: require('@/assets/images/profile/sessions.png') },
    { id: 'time', title: `Часов - ${stats.time}`, icon: require('@/assets/images/profile/hours.png') },
  ];

  const handleTilePress = (tileId: TileType) => {
    setLocalSelectedTiles((prev) => {
      const index = prev.indexOf(tileId);
      
      if (index > -1) {
        return prev.filter(id => id !== tileId);
      } else {
        if (prev.length >= 4) {
          return prev;
        }
        return [...prev, tileId];
      }
    });
  };

  const handleConfirm = () => {
    if (localSelectedTiles.length === 4) {
      onTilesChange(localSelectedTiles);
      onClose();
    }
  };

  const getTileSelectionNumber = (tileId: TileType): number | null => {
    const index = localSelectedTiles.indexOf(tileId);
    return index > -1 ? index + 1 : null;
  };

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
    >
      <TouchableWithoutFeedback onPress={onClose}>
        <View style={styles.overlay}>
          <TouchableWithoutFeedback>
            <View style={styles.container}>
              <Text style={styles.title}>
                Выберите плитки для отображения
              </Text>

              <View style={styles.tilesGrid}>
                {tileOptions.map((tile) => {
                  const selectionNumber = getTileSelectionNumber(tile.id);
                  const isSelected = selectionNumber !== null;

                  return (
                    <TouchableOpacity
                      key={tile.id}
                      style={styles.tile}
                      onPress={() => handleTilePress(tile.id)}
                      activeOpacity={0.7}
                    >
                      <Image source={tile.icon} style={styles.tileIcon} />
                      <Text style={styles.tileText}>{tile.title}</Text>
                      
                      {isSelected && (
                        <View style={styles.selectionBadge}>
                          <Text style={styles.selectionBadgeText}>
                            {selectionNumber}
                          </Text>
                        </View>
                      )}
                    </TouchableOpacity>
                  );
                })}
              </View>

              <TouchableOpacity 
                style={[
                  styles.confirmButton,
                  localSelectedTiles.length !== 4 && styles.confirmButtonDisabled
                ]}
                onPress={handleConfirm}
                disabled={localSelectedTiles.length !== 4}
              >
                <Text style={styles.confirmButtonText}>
                  Применить {localSelectedTiles.length !== 4 && `(${localSelectedTiles.length}/4)`}
                </Text>
              </TouchableOpacity>
            </View>
          </TouchableWithoutFeedback>
        </View>
      </TouchableWithoutFeedback>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  container: {
    backgroundColor: Colors.white,
    borderRadius: 24,
    padding: 24,
    maxWidth: 300,
  },
  title: {
    fontFamily: Montserrat.bold,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
    marginBottom: 16,
  },
  tilesGrid: {
    flexDirection: 'column',
    gap: 8,
    marginBottom: 24,
    marginHorizontal: 8
  },
  tile: {
    backgroundColor: Colors.lightBlue2,
    borderRadius: 40,
    paddingVertical: 8,
    paddingHorizontal: 16,
    flexDirection: 'row',
    alignItems: 'center',
    alignSelf: 'center',
    gap: 6,
    width: '100%',
  },
  tileIcon: {
    width: 30,
    height: 30,
    resizeMode: 'contain'
  },
  tileText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
  },
  selectionBadge: {
    position: 'absolute',
    top: 6,
    right: 12,
    backgroundColor: Colors.white,
    borderRadius: 40,
    width: 35,
    height: 35,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: Colors.lightBlue2,
  },
  selectionBadgeText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.lightBlue2,
  },
  confirmButton: {
    backgroundColor: Colors.blue3,
    borderRadius: 40,
    paddingVertical: 8,
    alignItems: 'center',
  },
  confirmButtonText: {
    fontFamily: Montserrat.bold,
    fontSize: 16,
    color: Colors.white,
  },
  confirmButtonDisabled:{
    backgroundColor: Colors.lightBlack
  }
});

export default TileSelectionModal;