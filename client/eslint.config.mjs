import eslint from "@eslint/js";
import { defineConfig } from "eslint/config";
import tseslint from "typescript-eslint";
import pluginVue from "eslint-plugin-vue";
import vueParser from "vue-eslint-parser";
import prettierConfig from "eslint-config-prettier";

export default defineConfig(
    {
        ignores: ["**/dist/**", "**/dist-ssr/**", "**/coverage/**", "**/node_modules/**", "**/components.d.ts"],
    },
    eslint.configs.recommended,
    ...tseslint.configs.recommended,
    ...pluginVue.configs["flat/recommended"],
    prettierConfig,
    {
        files: ["**/*.vue", "**/*.ts", "**/*.js"],
        languageOptions: {
            parser: vueParser,
            parserOptions: {
                parser: tseslint.parser,
                sourceType: "module",
            },
        },
        rules: {
            "@typescript-eslint/no-explicit-any": "off",
            "@typescript-eslint/no-unused-vars": "warn",
            "vue/multi-word-component-names": "off",
            "vue/no-v-html": "off",
            "@typescript-eslint/ban-ts-comment": "off",
            "no-undef": "off",

            // Compact formatting
            "vue/max-attributes-per-line": "off",
            "vue/singleline-html-element-content-newline": "off",
            "vue/multiline-html-element-content-newline": "off",
        },
    },
);
