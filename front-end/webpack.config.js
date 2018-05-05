var path = require('path');

const DIST_DIR = path.resolve(__dirname, './dist/js/');
const SRC_DIR = path.resolve(__dirname, './src/');
const NODE_MODULES = path.resolve(__dirname, './node_modules/');

module.exports ={
  mode: 'development',
  entry: {
      index: SRC_DIR + '/index.jsx',
      page1: SRC_DIR + '/page1.jsx',
  },
  output: {
    path: DIST_DIR,
    filename: '[name].bundle.js',
  },
  module : {
     rules : [
        {
            loader : 'babel-loader',
            test : /\.jsx?/,
            include : SRC_DIR,
            exclude: /node_modules/,
            options: {
                presets: ["es2015",]
            },
        },
     ],
  },
};

