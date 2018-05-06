const path = require('path');
const HtmlWebpackPlugin = require("html-webpack-plugin");

const DIST_DIR = path.resolve(__dirname, './dist/');
const SRC_DIR = path.resolve(__dirname, './src/');
const NODE_MODULES = path.resolve(__dirname, './node_modules/');

module.exports ={
  mode: 'development',
  entry: {

  }
  output: {
    path: DIST_DIR,
    filename: '[name].bundle.js',
  },
  module : {
     rules : [
        {
            loader : 'babel-loader',
            test : /\.jsx?$/,
            include : SRC_DIR,
            exclude: /node_modules/,
            options: {
                presets: ["es2015",]
            },
        },
        {
            loader : 'html-loader',
            test : /\.html$/,
        }
     ],
  },
  plugins : [
    new HtmlWebpackPlugin({
        template : SRC_DIR + "/public/index.html",
        filename : DIST_DIR + "/public/index.html"
    }),
    new HtmlWebpackPlugin({
        template : SRC_DIR + "/public/failed_login.html",
        filename : DIST_DIR + "/public/failed_login.html"
    }),
    new HtmlWebpackPlugin({
        template : SRC_DIR + "/public/page1.html",
        filename : DIST_DIR + "/public/page1.html"
    }),
    new HtmlWebpackPlugin({
        template : SRC_DIR + "/private/login_success.html",
        filename : DIST_DIR + "/private/login_success.html"
    }),
    new HtmlWebpackPlugin({
        template : SRC_DIR + "/private/page1.html",
        filename : DIST_DIR + "/private/page1.html"
    }),
  ],
};

