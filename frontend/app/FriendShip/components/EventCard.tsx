import { Colors } from '@/constants/Colors';
import { inter } from '@/constants/Inter';
import React from 'react';
import { Image, StyleSheet, Text, TouchableOpacity, View } from 'react-native';

const EventCard = () => {
  return (
    <View style={styles.card}>
        <View style={styles.imageWrapper}>
        <Image
          source={{ uri: 'https://kinopoisk-ru.clstorage.net/a14b6z215/1f19f0Jqc5/9ZhA4jPVlVhw5sFx59N-AcGjfEwji1Vh35MbrGUeQPUjQgM4-PX89v759hAXERPSxkKEc_Dm8NekKl9SBz3pa67NVrR0i4INcqTCbLSPIgCMuSQcb8EHK5iOQL0_RADFFXApV87Xz9jcn4lXV1YOgJoVxGb96bGPvTwdIG4p6iLitGHKJ8_aQ0Ldq0zF7X8z1eNYZj3tWUH87DyNRzJ-zA1rIHb6wcziy6yjT1dpbKqFCOpY08SQj_IcORhlZfAO7bhWziDe1Vlk4a5M6dBoCOzeaGI18EJH3ssas1kUV_skRQNS2dLr3P6egHdBbm2RkAvmUPXaob7cFiIzZH7fANe-Q-4l8eJbfdKHTu3JagP16X9YLv1GZsnWA45pMFTdKWEeesTdxeTMkINWY0djk4MW2XDZ6pWj6jgNPndd5D_xnmDSDPHCfkPankPV7XEU4d9QSBTdZF3_xAKIRT5U-ShTGkbGyOT33pumYXVvT5iyOu9S6u6zvMAFBjp6c-c1xJB_9TTNwnN-3LhK4vdYNvTYc0EkznB808sWikUkQ-wTZCJf_dTpwd6Uv2NxbHShvSvhTsX5jJPgLCkxVUvdGdmwde8ozeN_Q-yrXc_cTDPs6FVBPtNqQ_rvHKNMJn7uKFI8d9Pz7eLak6hRR3VTh5gC-Ebr-K2hyyUOL1xg7R3QnnPrCvfBX0DZrFr_-VsT9up5WgTKWWLn5wOVXyZd1D5XK0P_7-j6_YKwdkJga423IcJPwMKKi_khMBRJXv8L6q588g3X2nh-7a14w_p-Icnmekoi_mto78sYkW0JWd0iYjBf9frn8PeCrXZydG-ltDrsU97hj5beDSkQS1znANC2Q8Am9P1gZt2cRcHndz3D4kNsH8F4S_7_ArJIJ2XgIFQFQvfR1OLyvaVibHBJpaItwFvU-qe3xC8rC0Zs0hzLjmL1L-jNW1HCt3Li_m4o1-9zTzzIYFLx6geOZhRJyyJIP13z08rd9Lm_Z2FVR7O2OeBk3vKIj9QZJDF6bt8z27JyxCbc92xewItDxctCEtj6ZV4WwVRx3-4-lkQIU-0ScyJT1sLY1fqQtFZqS1CzhxjzUNnUt6HnOQovc0beCvWOUcIu0dx2XcGsR8T5WSzA6n1BHvJkdfnxGrNiCXLeGksaf9_Mzdz6mK9SQltCjbUM0lP-9pSU_DY7B15e_jnJvUbbCMrpdnDAv2PCzkM8yPliYwbPRHTl9j2kTgBI0x1dHnHT9_L17JG0aXpKRZikGdNR5fSPvd8fOQtGUN0w7LZM1Tnf6GJx24x84epMAdvrYUMwwH52zvsDn2Ukf-svVDFNwMjT4dyVgkR4cWqCpjrEbvzPpJ7rBSsHXFv4E_ucZNsPxdVsfMClWMfxdQ7fxH9mMsxUQO3sOpVuHXHJIXobV-Df3vf5mJBQanJhu5wd6kLxxq6X1Ac7D2hY4xTCmmP2F9zXR3L7mmTi-1Y0zdpXTRDAS0r93wKsWRhT3QNBHWT11fXZ2p2hVE5JQbujE9tR-MyosNY-JRZ8SuYe7bZNzQ7wzFJ4-7dszdFABcjHSEQj7WF0-Mg4h1wrWvAEci1m5NbO2dyfk2xZSHaelxHuWcHGkJLHIQ0sUEPwNNS1dtAg98ljUu2NR-PrajzP91dONf1haszbKK1sHUfuPXgCX_DX9OXPkY1pYXh6pLIe6lLqzJi1_xQbGlRw4QLStGfoEdHyfmTwl3z1_Egw08NFfB_xZ0r76jijZTtI9yNyLn3J8-fJzK6mY1hAWrKeAv5Z2uW1hdkeIDx_ffIrwalE5BPv-Xt24a5p48RRBs3senAS3GVe_O0zvWUfdskCQhtb1tXM-f2CiXBCZHGIkBzdcNPQjYDnOCYbXGLgPOKLbskqytx8Y9-6QNP3dRHX21l5IOVgYuLfAKFxFXbbG3gHRenq1t3VlaBOYVdkqJQ0w0fi8qax9TQmM2Fk-T3tq0TvCMH-VlvMi2PP_lcyzMFxTSXdZ0vj8BGCUiNW-wh3EF_Z0_j6_riKTXlPQa6LH89P08q2ocsWABNcRskJ4JhR7wrc6Gxd0JlgyNtQEMjPc2Eb9XxuwvUSgUwGVswdfARQy8_a2P69pXVLcHq5hS3jYtDCiJLRFiMlcmDHIsG2bPA59ehQUe25UcrebSj45GBxON1Eaf38Aa9II13WP2sqb_jhyNfPgpFwYnV5kZYo2WTk4aCU4RgSMFNo5C3rlkHZBODQXVLdrU3O-3AG_-VfUBnQfHzB1RuKXjdf3h5JPlPjyNXQ0KWLa1ZHbYiUPOd6yO2pguUsKyZGWsMMyJFy9BDH6WFA-oh-zvh3MfTGVGs-9XRgz_A0oVkwXfQfSAZQ9Pbs6MiEnW1pUkSKlzTxe8j1qY_QOTIFVFzOE8uPRPUs3_ZkXtyNcsf5TR309GtYEe1iZ9rYN71iJXvCPX8aWNL39dP4sIxYRE9hqJUI-3__xqar2QgeO3h68A3XrkP7COnsW3PEv0rOyGIh2v9RTS_9Q1bZ6wqXfDBj7BtZBXnmwNTS1rWfeExNSIyxKeZu5uGOlOs5HBxTfPkv8IBf8w3T6Vpuw4RZ8vdTFNXBU2E4z2FS2c0fllwGQuwEVDFD-tfx692vj05qc0K4ix76fdDwianGLw0UQ1DODu-URNEE6MFHX9-OYsX5WCXX1GBDIt1na8L4FYlkM2TAK2MmY_v5x_rXlLdqV3N5m5UX2kzGzIiJ-ggJKXhrygv2vHz1NsD2ZkTsh2vS1k4q1sp6XgLoa13e7DGsaDl--AJUEXrg4NX_zK-0VG5JVJ-mN81Xw-aEkdU6AhRjUdIa-rRfzCXf3nlH2r5iwPRgIM3cXFsl8HFv4-kYtGgeV_U4YRh9_Or15fO-smFKR3WMvCnmf-r0pZD_Ag07e1nmCMCxcu444PZyYO2BZ_XzYBPN5HFOE_xDV-jSNod_P37bP1IQRPDV7f_DuKxGQ1Fak6sU01Dr64-D0BY' }}
          style={styles.image}
        />
        <View style={styles.dateBadge}>
          <Text style={styles.dateText}>03.12.2024</Text>
        </View>
        <View style={styles.iconOverlay}>
          <Image source={require("../assets/images/event_card/movie.png")} style={{resizeMode: 'contain', width:20, height: 20}} />
        </View>
      </View>

      <View style={styles.content}>
        <Text style={styles.title}>Крестный отец</Text>
        <View style={styles.genres}>
          <Text style={[styles.metaText, {alignSelf: 'center'}]}>Жанры:</Text>
          {['Драма', 'Криминал'].map((genre) => (
            <View key={genre} style={styles.genreBadge}>
              <Text style={styles.genreText}>{genre}</Text>
            </View>
          ))}
        </View>

      <View style={styles.metaContainer}>
        <View style={styles.metaRow}>
          <Text style={styles.metaText}>Участников: 52/52</Text>
          <Image source={require("../assets/images/event_card/person.png")} style={styles.metaIcon} />
        </View>

        <View style={styles.metaRow}>
          <Text style={styles.metaText}>175 минут</Text>
          <Image source={require("../assets/images/event_card/duration.png")} style={styles.metaIcon} />
        </View>
      </View>

        <View style={styles.iconOverlay}>
          <Image source={require("../assets/images/event_card/online.png")} style={{resizeMode: 'contain', width:20, height: 20}} />
        </View>

        <TouchableOpacity style={styles.button}>
          <Text style={styles.buttonText}>Узнать подробнее</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  card: {
  width: '90%',
  alignSelf: 'center',
  borderRadius: 16,
  backgroundColor: Colors.white,
  borderColor: Colors.blue,
  borderWidth: 4,
  overflow: 'hidden',
  elevation: 4,
  minHeight: 200,
  },
  rowContainer: {
    flexDirection: 'row',
    minHeight: 100,
  },
  imageWrapper: {
    minHeight: 100,
    height: 'auto',
  },
  image: {
    width: '100%',
    height: 120,
    resizeMode: 'cover',
  },
  content: {
    minHeight: 120,
    padding: 12,
    justifyContent: 'space-between',
  },
  dateBadge: {
    position: 'absolute',
    top: 15,
    left: 15,
    backgroundColor: Colors.white,
    borderRadius: 12,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  dateText: {
    fontFamily: inter.regular,
    fontSize: 12,
    margin: -4,
    color: Colors.black,
  },
  iconOverlay: {
    position: 'absolute',
    top: 15,
    right: 15,
    backgroundColor: Colors.white,
    borderRadius: 20,
    padding: 6,
    borderWidth: 2,
    borderColor: Colors.lightBlue,
  },
  title: {
    fontFamily: inter.black,
    fontSize: 18,
    marginTop: -4,
    marginBottom: 8,
  },
  genres: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 6,
    marginBottom: 8,
  },
  genreBadge: {
    backgroundColor: Colors.lightBlue,
    borderRadius: 20,
    paddingHorizontal: 8,
    paddingVertical: 2,
  },
  genreText: {
    color: Colors.black,
    fontSize: 10,
  },
  metaContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  metaRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  metaIcon: {
    resizeMode: 'contain',
    width: 16,
    height: 16,
  },
  metaText: {
    fontSize: 14,
    color: Colors.black
  },
  button: {
    marginTop: 4,
    backgroundColor: Colors.lightBlue,
    borderRadius: 20,
    paddingVertical: 4,
    alignItems: 'center',
  },
  buttonText: {
    color: Colors.white,
    fontSize: 16,
  },
});

export default EventCard;
