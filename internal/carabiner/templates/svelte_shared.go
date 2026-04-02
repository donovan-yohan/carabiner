package templates

// SharedEnforceYAML is the enforce.yaml for all Svelte templates.
const SharedEnforceYAML = "version: 1\n" +
	"tools:\n" +
	"  svelte-check:\n" +
	"    enabled: true\n" +
	"    command: \"npx\"\n" +
	"    args: [\"svelte-check\", \"--fail-on-warnings\"]\n" +
	"    files: [\"src/**/*.{svelte,ts}\"]\n\n" +
	"  eslint:\n" +
	"    enabled: true\n" +
	"    command: \"npx\"\n" +
	"    args: [\"eslint\", \"--max-warnings=0\", \".\"]\n" +
	"    files: [\"src/**/*.{svelte,ts,js}\"]\n\n" +
	"  prettier:\n" +
	"    enabled: true\n" +
	"    command: \"npx\"\n" +
	"    args: [\"prettier\", \"--check\", \".\"]\n" +
	"    files: [\"src/**/*.{svelte,ts,js,css}\"]\n\n" +
	"  vitest:\n" +
	"    enabled: true\n" +
	"    command: \"npx\"\n" +
	"    args: [\"vitest\", \"run\"]\n" +
	"    files: [\"src/**/*.test.{ts,tsx}\"]\n\n" +
	"behavior:\n" +
	"  fail_on_warning: true\n" +
	"  stop_on_first_failure: false\n" +
	"  parallel: false\n"

// SharedEslintrc is the .eslintrc.json for all Svelte templates.
const SharedEslintrc = "{\n" +
	"  \"extends\": [\n" +
	"    \"eslint:recommended\",\n" +
	"    \"plugin:@typescript-eslint/recommended\",\n" +
	"    \"plugin:svelte/recommended\"\n" +
	"  ],\n" +
	"  \"plugins\": [\"@typescript-eslint\", \"svelte\"],\n" +
	"  \"rules\": {\n" +
	"    \"no-console\": \"error\",\n" +
	"    \"no-unused-vars\": \"off\",\n" +
	"    \"@typescript-eslint/no-unused-vars\": [\"error\", { \"argsIgnorePattern\": \"^_\" }],\n" +
	"    \"max-lines-per-function\": [\"error\", 40],\n" +
	"    \"complexity\": [\"error\", 10],\n" +
	"    \"max-depth\": [\"error\", 3],\n" +
	"    \"max-params\": [\"error\", 3],\n" +
	"    \"svelte/valid-compile\": \"error\",\n" +
	"    \"svelte/no-at-html-tags\": \"error\"\n" +
	"  }\n" +
	"}\n"

// SharedPrettierrc is the .prettierrc for all Svelte templates.
const SharedPrettierrc = "{\n" +
	"  \"semi\": true,\n" +
	"  \"singleQuote\": true,\n" +
	"  \"trailingComma\": \"es5\"\n" +
	"}\n"

// SharedPrettierIgnore is the .prettierignore for all Svelte templates.
const SharedPrettierIgnore = "node_modules\nbuild\ndist\n.svelte-kit\npackage\n.vite\n.env\n.env.*\n!.env.example\n"

// SharedTsconfig is the tsconfig.json for all Svelte templates.
const SharedTsconfig = "{\n" +
	"  \"extends\": \"./.svelte-kit/tsconfig.json\",\n" +
	"  \"compilerOptions\": {\n" +
	"    \"allowJs\": true,\n" +
	"    \"checkJs\": true,\n" +
	"    \"esModuleInterop\": true,\n" +
	"    \"forceConsistentCasingInFileNames\": true,\n" +
	"    \"resolveJsonModule\": true,\n" +
	"    \"skipLibCheck\": true,\n" +
	"    \"sourceMap\": true,\n" +
	"    \"strict\": true,\n" +
	"    \"moduleResolution\": \"bundler\"\n" +
	"  }\n" +
	"}\n"

// SharedVitestConfig is the vitest.config.ts for all Svelte templates.
const SharedVitestConfig = "import { defineConfig } from 'vitest/config';\n\n" +
	"export default defineConfig({\n" +
	"  test: {\n" +
	"    environment: 'jsdom',\n" +
	"    globals: true,\n" +
	"    include: ['src/**/*.{test,spec}.{js,ts}'],\n" +
	"  },\n" +
	"});\n"

// SharedPackageScripts is the scripts section for package.json (template uses full package.json).
const SharedPackageScripts = "\"scripts\": {\n" +
	"  \"dev\": \"vite dev\",\n" +
	"  \"build\": \"vite build\",\n" +
	"  \"preview\": \"vite preview\",\n" +
	"  \"check\": \"svelte-kit sync && svelte-check --tsconfig ./tsconfig.json\",\n" +
	"  \"check:watch\": \"svelte-kit sync && svelte-check --tsconfig ./tsconfig.json --watch\",\n" +
	"  \"test\": \"vitest run\",\n" +
	"  \"test:watch\": \"vitest\"\n" +
	"}\n"
