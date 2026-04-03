package templates

func getReactTypescriptTemplate() *Template {
	enforceYAML := `version: 1
tools:
  eslint:
    enabled: true
    command: "npx"
    args: ["eslint", "--max-warnings=0", "."]
    files: ["src/**/*.{ts,tsx}", "tests/**/*.{ts,tsx}"]

  typescript:
    enabled: true
    command: "npx"
    args: ["tsc", "--noEmit"]
    files: ["src/**/*.{ts,tsx}"]

  prettier:
    enabled: true
    command: "npx"
    args: ["prettier", "--check", "."]
    files: ["src/**/*.{ts,tsx,css,md}"]

  vitest:
    enabled: true
    command: "npx"
    args: ["vitest", "run"]
    files: ["src/**/*.test.{ts,tsx}"]

behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: false
`

	eslintrc := `{
  "//": "Requires eslint-plugin-sonarjs to be installed: npm install eslint-plugin-sonarjs --save-dev",
  "extends": ["eslint:recommended", "plugin:@typescript-eslint/recommended", "plugin:sonarjs/recommended"],
  "parser": "@typescript-eslint/parser",
  "plugins": ["@typescript-eslint", "sonarjs"],
  "rules": {
    "no-unused-vars": "error",
    "no-console": "warn"
  }
}
`

	tsconfig := `{
  "compilerOptions": {
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true
  }
}
`

	prettierrc := `{
  "semi": true,
  "singleQuote": true,
  "trailingComma": "es5"
}
`

	return &Template{
		Name:        "react-typescript",
		EnforceYAML: enforceYAML,
		ConfigFiles: map[string]string{
			".eslintrc.json": eslintrc,
			"tsconfig.json":  tsconfig,
			".prettierrc":    prettierrc,
		},
	}
}
