import { Colors } from '@/constants/Colors';
import { Montserrat_Alternates } from '@/constants/Montserrat-Alternates';
import React from 'react';
import { Image, StyleSheet, Text, View } from 'react-native';

interface PageHeaderProps {
  title: string;
  showWave?: boolean;
}

const PageHeader: React.FC<PageHeaderProps> = ({ title, showWave = false }) => {
  return (
    <>
      <View style={styles.headerContainer}>
        <Text style={styles.headerTitle}>{title}</Text>
      </View>
      {showWave && (
        <Image
          source={require('@/assets/images/wave.png')}
          style={{ resizeMode: 'none', marginBottom: -20 }}
        />
      )}
    </>
  );
};

const styles = StyleSheet.create({
  headerContainer: {
    backgroundColor: Colors.lightBlue2,
    paddingVertical: 20,
    alignItems: 'center',
    marginBottom: 12,
  },
  headerTitle: {
    fontFamily: Montserrat_Alternates.medium,
    fontSize: 24,
    color: Colors.white,
    textTransform: 'none',
  },
});

export default PageHeader;