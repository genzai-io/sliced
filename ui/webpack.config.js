var webpack = require('webpack');
var path = require('path');

// variables
var isProduction = process.argv.indexOf('-p') >= 0;
var sourcePath = path.join(__dirname, './src');
var outPath = path.join(__dirname, './dist');

// plugins
var HtmlWebpackPlugin = require('html-webpack-plugin');
var ExtractTextPlugin = require('extract-text-webpack-plugin');
var WebpackCleanupPlugin = require('webpack-cleanup-plugin');
const UglifyJsPlugin = require('uglifyjs-webpack-plugin');
const CompressionPlugin = require('compression-webpack-plugin');

module.exports = {
    // devtool: 'source-map',
    context: sourcePath,
    entry: {
        main: './main.tsx'
    },
    output: {
        path: outPath,
        filename: 'bundle.js',
        chunkFilename: '[chunkhash].js'
    },
    target: 'web',
    resolve: {
        extensions: ['.js', '.ts', '.tsx'],
        // Fix webpack's default behavior to not load packages with jsnext:main module
        // (jsnext:main directs not usually distributable es6 format, but es6 sources)
        mainFields: ['module', 'browser', 'main'],
        alias: {
            'app': path.resolve(__dirname, 'src/app/'),
            'proto': path.resolve(__dirname, 'src/proto/'),
        }
    },
    module: {
        rules: [
            // .ts, .tsx
            {
                test: /\.tsx?$/,
                use: isProduction
                    ? 'ts-loader'
                    : ['babel-loader?plugins=react-hot-loader/babel', 'ts-loader']
            },
            // "postcss" loader applies autoprefixer to our CSS.
            // "css" loader resolves paths in CSS and adds assets as dependencies.
            // "style" loader turns CSS into JS modules that inject <style> tags.
            // In production, we use a plugin to extract that CSS to a file, but
            // in development "style" loader enables hot editing of CSS.
            {
                test: /\.css$/,
                use: [
                    require.resolve('style-loader'),
                    {
                        loader: require.resolve('css-loader'),
                        options: {
                            importLoaders: 1,
                        },
                    },
                    require.resolve('postcss-loader'),
                ],
            },
            // css
            // {
            //     test: /\.css$/,
            //     use: ExtractTextPlugin.extract({
            //         fallback: 'style-loader',
            //         use: [
            //             // require.resolve('style-loader'),
            //             {
            //                 loader: 'css-loader',
            //                 query: {
            //                     // modules: true,
            //                     // sourceMap: !isProduction,
            //                     // importLoaders: 2,
            //                     // localIdentName: '[local]__[hash:base64:5]',
            //                 }
            //             },
            //             // require.resolve('postcss-loader'),
            //             {
            //                 loader: 'postcss-loader',
            //                 options: {
            //                     ident: 'postcss',
            //                     plugins: [
            //                         require('postcss-import')({addDependencyTo: webpack}),
            //                         require('postcss-url')(),
            //                         require('postcss-cssnext')(),
            //                         require('postcss-reporter')(),
            //                         require('postcss-browser-reporter')({
            //                             // disabled: isProduction
            //                         })
            //                     ]
            //                 }
            //             }
            //         ]
            //     })
            // },
            {
                test: /\.scss$/,
                use: [
                    require.resolve('style-loader'),
                    {
                        loader: require.resolve('css-loader'),
                        options: {
                            importLoaders: 2,
                        },
                    },
                    require.resolve('postcss-loader'),
                    require.resolve('sass-loader'),
                ],
            },
            // static assets
            {test: /\.html$/, use: 'html-loader'},
            {test: /\.png$/, use: 'file-loader'},
            {test: /\.jpg$/, use: 'file-loader'},
            {
                test: /\.(ttf|eot|woff|woff2)$/,
                loader: "url-loader",
                options: {
                    name: "fonts/[name].[ext]",
                },
            },
        ]
    },
    // optimization: {
    //     nodeEnv: 'development',
    //     // cache: true,
    //     flagIncludedChunks: true,
    //     runtimeChunk: true,
    //     namedModules: true,
    //     namedChunks: true,
    //     minimize: true,
    // },
    // optimization: {
    //     splitChunks: {
    //         name: true,
    //         cacheGroups: {
    //             commons: {
    //                 chunks: 'initial',
    //                 minChunks: 2
    //             },
    //             vendors: {
    //                 test: /[\\/]node_modules[\\/]/,
    //                 chunks: 'all',
    //                 priority: -10
    //             }
    //         }
    //     },
    //     runtimeChunk: false
    // },
    plugins: [
        new WebpackCleanupPlugin(),
        new ExtractTextPlugin({
            filename: 'styles.css',
            disable: !isProduction
        }),
        new HtmlWebpackPlugin({
            inject: true,
            template: 'assets/index.html',
            // minify: {
            //     removeComments: true,
            //     collapseWhitespace: true,
            //     removeRedundantAttributes: true,
            //     useShortDoctype: true,
            //     removeEmptyAttributes: true,
            //     removeStyleLinkTypeAttributes: true,
            //     keepClosingSlash: true,
            //     minifyJS: true,
            //     minifyCSS: true,
            //     minifyURLs: false,
            // }
        }),
        new CompressionPlugin({
            asset: "[path].gz[query]",
            algorithm: "gzip",
            test: /\.js$|\.css$|\.html$/,
            threshold: 10240,
            minRatio: 0.8
        }),
        // new UglifyJsPlugin({
        //
        // }),
    ],
    devServer: {
        contentBase: sourcePath,
        hot: true,
        inline: true,
        historyApiFallback: {
            disableDotRule: true
        },
        stats: 'minimal'
    },
    // devtool: 'cheap-module-eval-source-map',
    node: {
        // workaround for webpack-dev-server issue
        // https://github.com/webpack/webpack-dev-server/issues/60#issuecomment-103411179
        dgram: 'empty',
        fs: 'empty',
        net: 'empty',
        tls: 'empty',
        child_process: 'empty',
    }
};
