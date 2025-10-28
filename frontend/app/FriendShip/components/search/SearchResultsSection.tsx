import PageHeader from '@/components/PageHeader';
import { Colors } from '@/constants/Colors';
import { Montserrat } from '@/constants/Montserrat';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import React, { ReactNode } from 'react';
import { Image, StyleSheet, Text, View } from 'react-native';

interface SearchResultsSectionProps {
  title: string;
  searchQuery: string;
  hasResults: boolean;
  children?: ReactNode;
  showWave?: boolean;
}

const SearchResultsSection: React.FC<SearchResultsSectionProps> = ({
  title,
  searchQuery,
  hasResults,
  children,
  showWave = false,
}) => {
  return (
    <>
      <PageHeader title={title} showWave={showWave} />
      
      <View style={styles.resultsHeader}>
        <Text style={styles.resultsText}>Результаты поиска:</Text>
        <Image
          source={require('@/assets/images/line.png')}
          style={{ resizeMode: 'none' }}
        />
      </View>

      {hasResults ? (
        children
      ) : (
        <Text style={styles.noResultsText}>
          Ничего не найдено по запросу - {searchQuery}
        </Text>
      )}
    </>
  );
};

const styles = StyleSheet.create({
  resultsHeader: {
    paddingHorizontal: 16,
  },
  resultsText: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 20,
    color: Colors.blue2,
    marginBottom: 4,
  },
  noResultsText: {
    fontFamily: Montserrat.regular,
    fontSize: 18,
    color: Colors.black,
    textAlign: 'center',
  },
});

export default SearchResultsSection;