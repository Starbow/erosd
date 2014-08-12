module.exports = function (grunt) {

  // Load NPM tasks
  require('load-grunt-tasks')(grunt, ['grunt-*']);
  // var _ = require('underscore');

  // Displays the elapsed execution time of grunt tasks
  require('time-grunt')(grunt);

  // Project configuration.
  grunt.initConfig({
    distdir: 'dist',
    tmpdir: '.tmp',
    pkg: grunt.file.readJSON('package.json'),
    version: '<%= pkg.version %>',
    banner:
    '/*! <%= pkg.title || pkg.name %> - v<%= pkg.version %> - <%= grunt.template.today("yyyy-mm-dd") %>\n' +
    '<%= pkg.homepage ? " * " + pkg.homepage + "\\n" : "" %>' +
    ' * Copyright (c) <%= grunt.template.today("yyyy") %> <%= pkg.author %>;\n' +
    ' * Licensed <%= _.pluck(pkg.licenses, "type").join(", ") %>\n */\n', 
    src: {
      js: ['src/app/**/*.js', '!src/app/eros.proto.js'],
      // specs: ['test/**/*.spec.js'],
      // specsConfig: 'test/config/unit.js',
      html: ['src/*.html'],
      tpl: {
        app: ['src/app/**/*.tpl.html'],
        compiled: ['<%= tmpdir %>/templates/**/*.js']
      },
      css: [
        'src/css/<%= pkg.name %>.css',
        'components/bootstrap/css/bootstrap.css', // min.js is throwing exception
        'components/bootstrap/css/bootstrap-theme.min.css'
      ], 
      cssWatch: ['src/css/*.css'],
      // locales: {
      //   es: ['src/**/locale-es.json' ],
      //   en: ['src/**/locale-en.json' ],
      //   compiled: ['<%= tmpdir %>/locales/**/*.js', 'src/app/locale-default.tpl']
      // }
    },
    clean: ['<%= distdir %>/*', '<%= tmpdir %>/*'],
    copy: {
      assets: {
        files: [{ dest: '<%= distdir %>', src : ['**/*'], expand: true, cwd: 'src/assets/' }]
      },
      boilerplate: {
        files: [{ dest: '<%= distdir %>/html5-boilerplate', src : ['normalize.css', 'main.css'], expand: true, cwd: 'components/html5-boilerplate/css' }]
      },
      glyphicons: {
        files: [{ dest: '<%= distdir %>/fonts', src : '**', expand: true, cwd: 'components/bootstrap/fonts'}]
      },
      cssMap: {
        files: [{ dest: '<%= distdir %>/css', src: "*map", expand: true, cwd: 'src/css'}]
      }
    },
    // karma: {
    //   unit: { configFile: '<%= src.specsConfig %>' },
    //   watch: { configFile: '<%= src.specsConfig %>', singleRun:false, autoWatch: true}
    // },
    html2js: {
      app: {
        options: {
          base: 'src/app'
        },
        src: ['<%= src.tpl.app %>'],
        dest: '<%= tmpdir %>/templates/app.js',
        module: 'templates.app'
      }
    },
    concat:{
       // i18n files
      // langEN: {
      //   src:['src/app/locale-start-en.tpl', '<%= src.locales.en %>', 'src/app/locale-end.tpl'],
      //   dest: '<%= tmpdir %>/locales/localeEN.js'
      // },
      // langES: {
      //   src:['src/app/locale-start-es.tpl', '<%= src.locales.es %>', 'src/app/locale-end.tpl'],
      //   dest: '<%= tmpdir %>/locales/localeES.js'
      // },
      // app source + i18n files + templates
      js:{
        options: {
          banner: "<%= banner %>"
        },
        src:['<%= src.js %>'
          // ,'<%= src.locales.compiled %>'
          ,'<%= src.tpl.compiled %>'
        ],
        dest:'<%= distdir %>/<%= pkg.name %>.js'
      },
      // Pages
      index: {
        src: ['src/index.html'],
        dest: '<%= distdir %>/index.html',
        options: {
          process: true
        }
      },
      404: {
        src: ['src/404.html'],
        dest: '<%= distdir %>/404.html',
        options: {
          process: true
        }
      },
      500: {
        src: ['src/500.html'],
        dest: '<%= distdir %>/500.html',
        options: {
          process: true
        }
      },
      // Third party projects
      jquery: {
        // Not minimized, use only for dev
        src:['components/jquery/dist/jquery.js'],
        dest: '<%= distdir %>/jquery.js'
      },
      angular: {
        // Not minimized, use only for dev
        src:['components/angular/angular.js'],
        dest: '<%= distdir %>/angular.js'

      },
      protobuf: {
        // Not minimized, use only for dev
        src:['src/app/eros.proto.js'],
        dest: '<%= distdir %>/eros.proto.js'

      },
      dev: {
        src:[
          'components/angularjs-scroll-glue/src/scrollglue.js',
          'components/bootstrap/js/bootstrap.min.js',
          'components/html5-boilerplate/js/vendor/modernizr-2.6.2.min.js',

          // Outdated jquery that may back-support more IE versions. 
          // Consider using last version (with migrate?)
          // 'components/html5-boilerplate/js/vendor/jquery-1.10.2.min.js',

          'components/long/dist/Long.min.js',
          'components/bytebuffer/dist/ByteBufferAB.min.js',
          'components/protobuf/dist/ProtoBuf.min.js',

          'components/web-socket-js/swfobject.js',
          'components/web-socket-js/web_socket.js',
          'components/angular-animate/angular-animate.min.js',
          'components/angular-route/angular-route.min.js'
        ],
        dest: '<%= distdir %>/components.js'
      }
    },
    recess: {
      // TODO: Change for scss compiler
      build: {
        files: {
          '<%= distdir %>/css/<%= pkg.name %>.css': ['<%= src.css %>'] },
        options: {
          compile: true
        }
      }
    },
    sass: {
      dist: {
          options:{
            'outputStyle': 'compressed'
          },                             // target
          files: {                        // dictionary of files
              '<%= distdir %>/css/<%= pkg.name %>.css': 'src/css/scss/app.scss'     // 'destination': 'source'
          }
      },
      dev: {                              // another target
          options: {                      // dictionary of render options
              sourceMap: true
          },
          files: {
              'src/css/<%= pkg.name %>.css': 'src/css/scss/app.scss'
          }
      }
    },
    watch:{
      css: {
        files:['<%= src.cssWatch %>'],
        tasks: ['recess:build', 'timestamp']
      }
      // ,assets: {
      //   files:['<%= copy.assets.files[0].cwd %>/<%= copy.assets.files[0].src %>'],
      //   tasks: ['copy:assets', 'timestamp']
      // }
      ,html2jsApp: {
        files:['<%= src.tpl.app %>'],
        tasks: ['html2js:app', 'concat:js', 'timestamp']
      }
      ,karma: {
        files:['<%= src.specs %>', '<%= src.specsConfig %>'],
        tasks: ['karma:unit', 'timestamp']
      }
      ,html: {
        files:['<%= src.html %>'],
        tasks: ['concat:index', 'concat:404', 'concat:500', 'timestamp']
      }
      // ,locales: {
      //   files:['<%= src.locales.es %>', '<%= src.locales.en %>'],
      //   tasks: ['concat:langEN', 'concat:langES', 'concat:js', 'timestamp']
      // }
      ,js: {
        files:['<%= src.js %>'],
        tasks: ['concat:js', 'timestamp']

      }
    },
    jshint: {
      all: {
        src: ['gruntFile.js', '<%= src.js %>', '<%= src.tpl.compiled %>', '<%= src.specs %>']
      },
      options:{
        curly:true,
        eqeqeq:true,
        immed:true,
        latedef:true,
        newcap:true,
        noarg:true,
        sub:true,
        boss:true,
        eqnull:true,
        laxcomma: true,
        "-W099": true, // Mix tabs and spaces
        "-W033": true, // Missing semi colon
        globals:{}
      }
    }
  });

  // The build
  grunt.registerTask('build', ['clean', 'html2js', 'concat', 'sass:dev', 'recess:build', 'copy']);
  grunt.registerTask('live', ['clean', 'html2js', 'concat', 'recess:build', 'copy', 'sass:dist'])

  // Default task.
  grunt.registerTask('default', ['jshint', 'config:dev', 'sbuild']);

  // Print a timestamp (useful for when watching)
  grunt.registerTask('timestamp', function() {
    grunt.log.subhead(Date());
  });

  // Making grunt default to force in order not to break the project.
  grunt.option('force', true);

};
