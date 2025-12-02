import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';

interface Genre {
  name: string;
  count: number;
}

interface FavoriteGenresProps {
  genres: Genre[];
}

const FavoriteGenres: React.FC<FavoriteGenresProps> = ({ genres }) => {
  if (!genres || genres.length === 0) {
    return (
      <Text style={styles.emptyText}>
        Пока нет любимых жанров
      </Text>
    );
  }

  const getFontFamily = (index: number): string => {
    const totalGenres = genres.length;
    
    if (index === 0) return Montserrat.semibold;
    if (index === totalGenres - 1) return Montserrat.light;
    
    return Montserrat.regular;
  };

  const leftGenres = genres.slice(0, Math.ceil(genres.length / 2));
  const rightGenres = genres.slice(Math.ceil(genres.length / 2));

  return (
    <View style={styles.genresContainer}>
      <View style={styles.genresColumn}>
        {leftGenres.map((genre, index) => (
          <Text 
            key={`left-${index}`} 
            style={[
              styles.genreItem,
              { fontFamily: getFontFamily(index) }
            ]}
          >
            {index + 1}. {genre.name} - {genre.count}
          </Text>
        ))}
      </View>
      <View style={styles.genresColumn}>
        {rightGenres.map((genre, index) => {
          const globalIndex = leftGenres.length + index;
          return (
            <Text 
              key={`right-${index}`} 
              style={[
                styles.genreItem,
                { fontFamily: getFontFamily(globalIndex) }
              ]}
            >
              {globalIndex + 1}. {genre.name} - {genre.count}
            </Text>
          );
        })}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  genresContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  genresColumn: {
    flex: 1,
  },
  genreItem: {
    fontSize: 15,
    color: Colors.black,
    marginBottom: 4,
  },
  emptyText: {
    fontFamily: Montserrat.regular,
    fontSize: 16,
    color: Colors.black,
    alignSelf: 'center',
    marginVertical: 8,
  },
});

export default FavoriteGenres;