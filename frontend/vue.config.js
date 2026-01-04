module.exports = {
  // Disable features that cause rendering issues in Wails on HiDPI displays
  css: {
    extract: true,
    sourceMap: false,
  },
  
  configureWebpack: {
    optimization: {
      splitChunks: {
        cacheGroups: {
          styles: {
            name: 'styles',
            test: /\.css$/,
            chunks: 'all',
            enforce: true,
          },
        },
      },
    },
  },
  
  chainWebpack: config => {
    // Disable prefetch and preload for better Wails compatibility
    config.plugins.delete('prefetch');
    config.plugins.delete('preload');
    
    // Optimize for Wails environment
    config.performance
      .maxEntrypointSize(512000)
      .maxAssetSize(512000);
  },
  
  // Production source maps can cause issues in Wails
  productionSourceMap: false,
  
  // Disable file name hashing for easier debugging in Wails
  filenameHashing: false,
}
