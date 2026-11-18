import { defineStore } from "pinia";

type ThemeMode = "light" | "dark";

// index.html reads the same key to set the theme before the app loads.
const STORAGE_KEY = "app_theme";

function readStoredTheme(): ThemeMode {
    const value = localStorage.getItem(STORAGE_KEY);
    if (value === "light" || value === "dark") return value;
    return "dark";
}

export const useThemeStore = defineStore("theme", {
    state: () => ({
        theme: readStoredTheme(),
    }),
    getters: {
        isDark(): boolean {
            return this.theme === "dark";
        },
    },
    actions: {
        initializeTheme(): void {
            this.applyTheme();
        },

        setTheme(theme: ThemeMode): void {
            this.theme = theme;
            localStorage.setItem(STORAGE_KEY, theme);
            this.applyTheme();
        },

        applyTheme(): void {
            document.documentElement.classList.toggle("app-dark", this.isDark);
        },
    },
});
