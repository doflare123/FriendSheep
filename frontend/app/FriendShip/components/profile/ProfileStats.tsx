import StatisticsBar from '@/components/profile/StatisticsBar';
import { TileType } from '@/components/profile/TileSelectionModal';
import React from 'react';
import { StyleSheet, View } from 'react-native';

interface ProfileStatsProps {
  selectedTiles: TileType[];
  stats: {
    movies: number;
    video_games: number;
    board_games: number;
    other: number;
    time: number;
    all: number;
  };
}

interface TileConfig {
  type: TileType;
  title: string;
  icon: string;
}

const tileConfigs: Record<TileType, TileConfig> = {
  movies: { type: 'movies', title: 'Медиа', icon: 'movies' },
  video_games: { type: 'video_games', title: 'Игры', icon: 'games' },
  board_games: { type: 'board_games', title: 'Настолки', icon: 'table_games' },
  other: { type: 'other', title: 'Другое', icon: 'other' },
  time: { type: 'time', title: 'Часов', icon: 'hours' },
  all: { type: 'all', title: 'Сессий', icon: 'sessions' },
};

const ProfileStats: React.FC<ProfileStatsProps> = ({ selectedTiles, stats }) => {
  const displayedTiles = selectedTiles.slice(0, 4);
  
  const renderTile = (tileType: TileType, index: number) => {
    const config = tileConfigs[tileType];
    if (!config) return null;
    
    const count = stats[tileType];
    
    return (
      <StatisticsBar
        key={`${tileType}-${index}`}
        title={config.title}
        count={count}
        icon={config.icon}
        fullWidth
      />
    );
  };

  const firstRow = displayedTiles.slice(0, 2);
  const secondRow = displayedTiles.slice(2, 4);

  return (
    <>
      {firstRow.length > 0 && (
        <View style={styles.statisticsRow}>
          {firstRow.map((tile, idx) => renderTile(tile, idx))}
        </View>
      )}
      
      {secondRow.length > 0 && (
        <View style={[styles.statisticsRow, styles.lastRow]}>
          {secondRow.map((tile, idx) => renderTile(tile, idx + 2))}
        </View>
      )}
    </>
  );
};

const styles = StyleSheet.create({
  statisticsRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 8,
    paddingHorizontal: 16,
    marginVertical: 2,
  },
  lastRow: {
    marginBottom: 14,
  },
});

export default ProfileStats;