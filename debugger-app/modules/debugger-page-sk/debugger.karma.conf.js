const startIdx = process.argv.findIndex((v) => v === 'start');

// The path to the JS bundle is passed by the karma_test rule as the second argument after start.
const jsTestFile = process.argv[startIdx + 2];
// version.js, debugger.js and debugger.wasm files are right after that.
const versionJSFile = process.argv[startIdx + 3];
const debuggerJSFile = process.argv[startIdx + 4];
const debuggerWASMFile = process.argv[startIdx + 5];

// Detect whether we're running as a test (e.g. "bazel test //path/to/my:karma_test").
//
// https://github.com/bazelbuild/rules_nodejs/blob/681c6683dac742f1e375a401c0399ec7783ac8fd/packages/karma/karma_web_test.bzl#L257
const isBazelTest = !process.env.BUILD_WORKSPACE_DIRECTORY; // Set when running via "bazel run".

// Returns the path to the Bazel runfiles directory.
//
// See:
//  - https://docs.bazel.build/versions/master/skylark/rules.html#runfiles-location
//  - https://docs.bazel.build/versions/master/test-encyclopedia.html#initial-conditions
const bazelRunfilesDir = () => process.env.RUNFILES_DIR + '/' + process.env.TEST_WORKSPACE;

// Forces Karma to use the Bazel-downloaded Google Chrome browser.
process.env.CHROME_BIN = bazelRunfilesDir() + '/external/google_chrome/opt/google/chrome/chrome';

module.exports = function(config) {
  config.set({
    plugins: [
      'karma-chrome-launcher',
      'karma-sinon',
      'karma-mocha',
      'karma-chai',
      'karma-chai-dom',
      'karma-spec-reporter',
    ],

    // Frameworks are loaded in reverse order, so chai-dom loads after chai.
    frameworks: ['mocha', 'chai-dom', 'chai', 'sinon'],

    files: [
      // We want the WASM file to be available for loading by the debugger JS file.
      { pattern: debuggerWASMFile, included: false, served: true },
      // We want the version.js file to be run before the tests so SKIA_VERSION is defined.
      { pattern: versionJSFile },
      // We want the debugger.js file to be run before the tests so DebuggerInit is defined.
      { pattern: debuggerJSFile },
      {
        pattern: jsTestFile,
        // Force the test files to be served from disk on each request. Without this,
        // interactive mode with ibazel does not work (e.g. "ibazel run //path/to/my:karma_test").
        nocache: true,
      },
    ],

    proxies: {
      // This lets our tests just try to load /dist/debugger.wasm instead of the
      // actual path (which is deep inside Bazel's output directory)
      '/dist/debugger.wasm': debuggerWASMFile,
    },

    // Only use a headless browser when running as a test (i.e. "bazel test").
    //
    // Do not launch any browsers when running interactively (i.e. with "bazel run"). The developer
    // can open the printed out URL for the Karma server in their browser of choice.
    browsers: isBazelTest ? ['ChromeHeadlessCustom'] : [],

    // Headless Chrome that works on the RBE toolchain container, and any other Docker container.
    //
    // https://github.com/puppeteer/puppeteer/blob/master/docs/troubleshooting.md#running-puppeteer-in-docker
    customLaunchers: {
      ChromeHeadlessCustom: {
        base: 'ChromeHeadless',
        flags: ['--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
      },
    },

    // Report the outcome of each individual test case to stdout.
    //
    // This is consistent with the output format of Puppeteer tests (and any other Node.js tests
    // using Mocha).
    colors: true,
    reporters: ['spec'],

    // Run tests only once when invoked via "bazel test", otherwise Karma will run forever and Bazel
    // will eventually time out and report test failure.
    singleRun: isBazelTest,

    // Autowatch is disabled because it is currently broken in two major ways:
    //
    //   1. It doesn't force a browser refresh when files change.
    //   2. Sometimes it randomly reports changed files as deleted, and stops executing the contents
    //      of those files for the remainder of the session.
    //
    // ibazel can be used to achieve a similar workflow:
    //
    //   - `ibazel test //path/to/my:karma_test`: Reruns the tests on a headless Chrome every time
    //     the tests change, and reports the results on the console.
    //   - `ibazel run //path/to/my:karma_test`: Interactive mode. Prints out a URL that can be
    //     opened in the browser. Tests are rebuilt automatically when the code changes. Reload the
    //     page manually to see the changes.
    autoWatch: false,
  });
};
