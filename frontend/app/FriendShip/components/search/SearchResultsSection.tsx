import PageHeader from '@/components/PageHeader';
import { Montserrat } from '@/constants/Montserrat';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import { useThemedColors } from '@/hooks/useThemedColors';
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
  const colors = useThemedColors();

  return (
    <>
      <PageHeader title={title} showWave={showWave} />
      
      <View style={styles.resultsHeader}>
        <Text style={[styles.resultsText, { color: colors.blue }]}>
          Результаты поиска:
        </Text>
        <Image
          source={require('@/assets/images/line.png')}
          style={{ resizeMode: 'none' }}
        />
      </View>

      {hasResults ? (
        children
      ) : (
        <Text style={[styles.noResultsText, { color: colors.black }]}>
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
    marginBottom: 4,
  },
  noResultsText: {
    fontFamily: Montserrat.regular,
    fontSize: 18,
    textAlign: 'center',
  },
});

export default SearchResultsSection;