const path = require('path');

// App Router does not use next.config.js i18n. Keep default lang in layout and
// handle locale routing in the app/ tree when needed. Also set tracing root for monorepo.
const nextConfig = {
	reactStrictMode: true,
	outputFileTracingRoot: path.join(__dirname, '..'),
};

module.exports = nextConfig;
