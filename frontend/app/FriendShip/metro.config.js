const path = require('path');
const { getDefaultConfig } = require('expo/metro-config');

const projectRoot = path.resolve(__dirname);
const defaultConfig = getDefaultConfig(projectRoot);

defaultConfig.resolver.assetExts.push('png', 'jpg', 'jpeg', 'svg', 'gif', 'webp');

module.exports = {
  ...defaultConfig,
  projectRoot,
  watchFolders: [projectRoot],
  server: {
    port: 8081,
  },
  resolver: {
    ...defaultConfig.resolver,
    sourceExts: [...defaultConfig.resolver.sourceExts, 'jsx', 'js', 'ts', 'tsx', 'json'],
  },
};