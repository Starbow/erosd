// Karma configuration
// Generated on Tue Sep 16 2014 14:30:26 GMT+0200 (CEST)

module.exports = function(config) {
  config.set({

    // base path that will be used to resolve all patterns (eg. files, exclude)
    basePath: '..',


    // frameworks to use
    // available frameworks: https://npmjs.org/browse/keyword/karma-adapter
    frameworks: ['jasmine'],


    // list of files / patterns to load in the browser
    files: [
    	'components/angular/angular.js',
    	'components/jquery/dist/jquery.js',

        'components/angularjs-scroll-glue/src/scrollglue.js',
		'src/bootstrap/js/bootstrap.min.js',
		'components/html5-boilerplate/js/vendor/modernizr-2.6.2.min.js',

		'components/long/dist/Long.min.js',
		'components/bytebuffer/dist/ByteBufferAB.min.js',
		'components/protobuf/dist/ProtoBuf.min.js',

		'components/web-socket-js/swfobject.js',
		'components/web-socket-js/web_socket.js',
		'components/angular-animate/angular-animate.min.js',
		'components/angular-route/angular-route.min.js',

		'src/**/*.js',
    	'test/**/*.spec.js',
    ],


    // list of files to exclude
    exclude: [
    	'src/app/eros.proto.js'
    ],


    // preprocess matching files before serving them to the browser
    // available preprocessors: https://npmjs.org/browse/keyword/karma-preprocessor
    preprocessors: {
    },


    // test results reporter to use
    // possible values: 'dots', 'progress'
    // available reporters: https://npmjs.org/browse/keyword/karma-reporter
    reporters: ['spec'],


    // web server port
    port: 9876,


    // enable / disable colors in the output (reporters and logs)
    colors: true,


    // level of logging
    // possible values: config.LOG_DISABLE || config.LOG_ERROR || config.LOG_WARN || config.LOG_INFO || config.LOG_DEBUG
    logLevel: config.LOG_INFO,


    // enable / disable watching file and executing tests whenever any file changes
    autoWatch: true,


    // start these browsers
    // available browser launchers: https://npmjs.org/browse/keyword/karma-launcher
    // browsers: ['Chrome', 'Firefox', 'IE', 'Safari'],
    browsers: ['Chrome'],


    // Continuous Integration mode
    // if true, Karma captures browsers, runs the tests and exits
    singleRun: true
  });
};
