package templates

func getSvelteKitTemplate() *Template {
	packageJSON := `{
  "name": "svelte-kit-app",
  "version": "0.0.1",
  "private": true,
  "scripts": ` + SharedPackageScripts + `,
  "devDependencies": {
    "@sveltejs/adapter-auto": "^3.0.0",
    "@sveltejs/kit": "^2.0.0",
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

	svelteConfig := `import adapter from '@sveltejs/adapter-auto';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter(),
  },
};

export default config;
`

	viteConfig := `import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
});
`

	appHTML := `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <link rel="icon" href="%sveltekit.assets%/favicon.png" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    %sveltekit.head%
  </head>
  <body data-sveltekit-preload-data="hover">
    <div style="display: contents">%sveltekit.body%</div>
  </body>
</html>
`

	appCSS := `/* Global styles */
:root {
  font-family: system-ui, -apple-system, sans-serif;
  line-height: 1.5;
}

body {
  margin: 0;
  padding: 1rem;
}
`

	routesPageSvelte := `<script lang="ts">
  import Counter from '$lib/components/Counter.svelte';
</script>

<h1>Welcome to SvelteKit</h1>
<Counter />
`

	routesLayoutSvelte := `<script lang="ts">
  import '../app.css';
</script>

<slot />
`

	routesPageServer := `import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
  return {
    timestamp: Date.now(),
  };
};
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

	dbServer := `// Server-only database utilities
// Never import this file from client-side code

export async function query<T>(sql: string, params?: unknown[]): Promise<T[]> {
  // TODO: implement database query
  return [] as T[];
}
`

	return &Template{
		Name:        "svelte-kit",
		EnforceYAML: SharedEnforceYAML,
		ConfigFiles: map[string]string{
			"package.json":                      packageJSON,
			"svelte.config.js":                  svelteConfig,
			"vite.config.ts":                    viteConfig,
			"tsconfig.json":                     SharedTsconfig,
			".eslintrc.json":                    SharedEslintrc,
			".prettierrc":                       SharedPrettierrc,
			".prettierignore":                   SharedPrettierIgnore,
			"vitest.config.ts":                  SharedVitestConfig,
			"src/app.html":                      appHTML,
			"src/app.css":                       appCSS,
			"src/routes/+page.svelte":           routesPageSvelte,
			"src/routes/+layout.svelte":         routesLayoutSvelte,
			"src/routes/+page.server.ts":        routesPageServer,
			"src/lib/components/Counter.svelte": counterComponent,
			"src/lib/server/db.ts":              dbServer,
		},
	}
}
