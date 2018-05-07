const path = require('path');
const HtmlWebpackPlugin = require("html-webpack-plugin");

const DIST_DIR = path.resolve(__dirname, './dist/');
const SRC_DIR = path.resolve(__dirname, './src/');
const NODE_MODULES = path.resolve(__dirname, './node_modules/');

const config = {
  mode: 'development',
  devtool: "source-map",
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
};

const private_config = Object.assign({}, config, {
    name : 'private',
    entry: {
        login_success: path.resolve(SRC_DIR, "private/js/login_success.jsx"),
        page1: path.resolve(SRC_DIR, "private/js/page1.jsx"),
    },
    output : {
        path : DIST_DIR + '/private/js/',
        filename: '[name].bundle.js',
    },
    plugins : [
      new HtmlWebpackPlugin({
          template : SRC_DIR + "/private/login_success.html",
          filename : DIST_DIR + "/private/login_success.html",
          chunks : ['login_success']
      }),
      new HtmlWebpackPlugin({
          template : SRC_DIR + "/private/page1.html",
          filename : DIST_DIR + "/private/page1.html",
          chunks : ['page1'],
      }),
    ],
});

const public_config = Object.assign({}, config, {
    name : 'public',
    entry: {
        index: path.resolve(SRC_DIR, "public/js/index.jsx"),
        failed_login: path.resolve(SRC_DIR, "public/js/failed_login.jsx"),
        page1: path.resolve(SRC_DIR, "public/js/page1.jsx"),
        signup: path.resolve(SRC_DIR, "public/js/signup.jsx"),
    },
    output : {
        path : DIST_DIR + '/public/js/',
        filename: '[name].bundle.js',
    },
    plugins : [
      new HtmlWebpackPlugin({
          template : SRC_DIR + "/public/index.html",
          filename : DIST_DIR + "/public/index.html",
          chunks : ['index'],
      }),
      new HtmlWebpackPlugin({
          template : SRC_DIR + "/public/failed_login.html",
          filename : DIST_DIR + "/public/failed_login.html",
          chunks : ['failed_login'],
      }),
      new HtmlWebpackPlugin({
          template : SRC_DIR + "/public/signup.html",
          filename : DIST_DIR + "/public/signup.html",
          chunks : ['signup'],
      }),
      new HtmlWebpackPlugin({
          template : SRC_DIR + "/public/page1.html",
          filename : DIST_DIR + "/public/page1.html",
          chunks : ['page1'],
      }),
    ],
});

module.exports = [
    private_config, public_config,
];
