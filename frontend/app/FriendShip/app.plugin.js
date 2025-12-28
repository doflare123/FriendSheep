const { withAndroidManifest } = require('@expo/config-plugins');

const withCustomFirebaseConfig = (config) => {
  return withAndroidManifest(config, (config) => {
    const { manifest } = config.modResults;

    const application = manifest.application[0];

    if (!application['meta-data']) {
      application['meta-data'] = [];
    }

    const colorMetaDataIndex = application['meta-data'].findIndex(
      (meta) =>
        meta.$['android:name'] === 'com.google.firebase.messaging.default_notification_color'
    );

    const colorMetaData = {
      $: {
        'android:name': 'com.google.firebase.messaging.default_notification_color',
        'android:resource': '@color/notification_icon_color',
        'tools:replace': 'android:resource',
      },
    };

    if (colorMetaDataIndex !== -1) {
      application['meta-data'][colorMetaDataIndex] = colorMetaData;
    } else {
      application['meta-data'].push(colorMetaData);
    }

    if (!manifest.$['xmlns:tools']) {
      manifest.$['xmlns:tools'] = 'http://schemas.android.com/tools';
    }

    return config;
  });
};

module.exports = withCustomFirebaseConfig;