'use strict';

module.exports = function (grunt) {
	require('time-grunt')(grunt);
	require('load-grunt-tasks')(grunt);

	grunt.initConfig({
		yeoman: {
            scss: 'static/src/scss',
            css: 'static/css',
            coffee: 'static/src/coffee',
            js: 'static/js',
            images: 'static/images'
        },

        watch: {
        	coffeeweb: {
                files: ['<%= yeoman.coffee %>/{,*/}*.coffee'],
                tasks: ['coffee:web']
            },
            compass: {
                files: ['<%= yeoman.scss %>/{,*/}*.{scss,sass}'],
                tasks: ['compass:server']
            },
            livereload: {
                options: {
                    livereload: '<%= connect.options.livereload %>'
                },
                files: [
                    '{,*/}*.html',
                    '<%= yeoman.css %>{,*/}*.css',
                    '<%= yeoman.images %>/{,*/}*.{png,jpg,jpeg,gif,webp,svg}',
                    '<%= yeoman.js %>/{,*/}*.js'
                ]
            }
        },
        connect: {
            options: {
                port: 9000,
                livereload: 35729,
                // change this to '0.0.0.0' to access the server from outside
                hostname: 'localhost'
            },
            livereload: {
                options: {
                    open: true,
                    base: ['.']
                }
            }
        },
        coffee: {
            web: {
                files: [{
                    expand: true,
                    cwd: '<%= yeoman.coffee %>',
                    src: '{,*/}*.coffee',
                    dest: '<%= yeoman.js %>',
                    ext: '.js',
                }],
                options: {
                    sourceMap: true
                }
            }
        },
        compass: {
            options: {
                sassDir: '<%= yeoman.scss %>',
                cssDir: '<%= yeoman.css %>',
                generatedImagesDir: '<%= yeoman.images %>/generated',
                imagesDir: '<%= yeoman.images %>',
                javascriptsDir: '<%= yeoman.js %>',
                fontsDir: 'fonts',
                importPath: '/',
                httpImagesPath: '/<%= yeoman.images %>',
                httpGeneratedImagesPath: '/<%= yeoman.images %>/generated',
                httpFontsPath: '/fonts',
                relativeAssets: false,
                assetCacheBuster: false
            },
            dist: {
                options: {
                    generatedImagesDir: '<%= yeoman.images %>/generated'
                }
            },
            server: {
                options: {
                    debugInfo: false
                }
            }
        },
        uglify: {
            options: {
                mangle: false
            },
            all: {
                src : 'static/js/*.js',
                dest : 'static/build.js'
            }
        },
        open: {
          server: {
            path: 'http://localhost:<%= connect.options.port %>'
          }
        },
        concurrent: {
            server: [
                'compass:server',
                'coffee:web'
            ]
        }
	});
    grunt.registerTask('server', ['concurrent:server', 'connect:livereload', 'watch']);
    grunt.registerTask('build', ['coffee:web', 'compass:dist']);
}