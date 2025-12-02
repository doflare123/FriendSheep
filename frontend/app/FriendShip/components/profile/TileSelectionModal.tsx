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

  const sortedTileOptions = [
    ...localSelectedTiles.map(id => tileOptions.find(t => t.id === id)!).filter(Boolean),
    ...tileOptions.filter(t => !localSelectedTiles.includes(t.id))
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
                <View style={styles.row}>
                  {sortedTileOptions.slice(0, 2).map((tile) => {
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

                <View style={styles.row}>
                  {sortedTileOptions.slice(2, 4).map((tile) => {
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

                <View style={styles.row}>
                  {sortedTileOptions.slice(4, 6).map((tile) => {
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
    padding: 16,
    maxWidth: 340,
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
    marginHorizontal: 8,
  },
  row: {
    flexDirection: 'row',
    gap: 8,
  },
  tile: {
    flex: 1,
    backgroundColor: Colors.lightBlue2,
    borderRadius: 40,
    paddingVertical: 8,
    paddingHorizontal: 8,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  tileIcon: {
    width: 22,
    height: 22,
    resizeMode: 'contain',
  },
  tileText: {
    fontFamily: Montserrat.regular,
    fontSize: 14,
    color: Colors.black,
    flex: 1,
  },
  selectionBadge: {
    position: 'absolute',
    top: -6,
    right: -2,
    backgroundColor: Colors.white,
    borderRadius: 20,
    width: 20,
    height: 20,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: Colors.lightBlue2,
  },
  selectionBadgeText: {
    fontFamily: Montserrat.bold,
    fontSize: 10,
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
  confirmButtonDisabled: {
    backgroundColor: Colors.lightBlack,
  },
});

export default TileSelectionModal;