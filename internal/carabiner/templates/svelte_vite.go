package templates

func getSvelteViteTemplate() *Template {
	packageJSON := `{
  "name": "svelte-vite-app",
  "version": "0.0.1",
  "private": true,
  "scripts": ` + SharedPackageScriptsVite + `,
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^4.0.0",
    "svelte": "^5.0.0",
    "svelte-check": "^4.0.0",
    "typescript": "^5.0.0",
    "vite": "^6.0.0",
    "eslint": "^9.0.0",
    "prettier": "^3.0.0",
    "prettier-plugin-svelte": "^3.0.0",
    "@typescript-eslint/eslint-plugin": "^8.0.0",
    "@typescript-eslint/parser": "^8.0.0",
    "eslint-plugin-svelte": "^2.0.0",
    "vitest": "^2.0.0",
    "jsdom": "^25.0.0"
  },
  "type": "module"
}
`

	svelteConfig := `import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
  preprocess: vitePreprocess(),
};
`

	viteConfig := `import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
});
`

	indexHTML := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <link rel="icon" href="/favicon.png" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Svelte App</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
`

	appCSS := `:root {
  font-family: system-ui, -apple-system, sans-serif;
  line-height: 1.5;
}

body {
  margin: 0;
  padding: 1rem;
}
`

	appSvelte := `<script lang="ts">
  import Counter from './lib/components/Counter.svelte';
</script>

<Counter />
`

	mainTS := `import App from './App.svelte';
import './app.css';

const app = new App({
  target: document.getElementById('app')!,
});

export default app;
`

	counterComponent := `<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);

  function increment() {
    count++;
  }
</script>

<button onclick={increment}>
  Count: {count}, Doubled: {doubled}
</button>
`

	formatUtils := `// Utility functions
export function formatDate(date: Date): string {
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}
`

	return &Template{
		Name:        "svelte-vite",
		EnforceYAML: SharedEnforceYAML,
		ConfigFiles: map[string]string{
			"package.json":                      packageJSON,
			"svelte.config.js":                  svelteConfig,
			"vite.config.ts":                    viteConfig,
			"tsconfig.json":                     SharedTsconfigVite,
			".eslintrc.json":                    SharedEslintrc,
			".prettierrc":                       SharedPrettierrc,
			".prettierignore":                   SharedPrettierIgnore,
			"vitest.config.ts":                  SharedVitestConfig,
			"index.html":                        indexHTML,
			"src/app.css":                       appCSS,
			"src/App.svelte":                    appSvelte,
			"src/main.ts":                       mainTS,
			"src/lib/components/Counter.svelte": counterComponent,
			"src/lib/utils/format.ts":           formatUtils,
		},
	}
}
