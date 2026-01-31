import { defineStore } from "pinia";

export const useThemeStore = defineStore("theme", {
    state: () => ({
        theme: "system" as "system" | "dark" | "light",
    }),
    getters: {
        isDark(): boolean {
            if (this.theme === "system") {
                return true;
            }
            return this.theme === "dark";
        },
    },
    actions: {
        initializeTheme() {
            // Set initial theme (system default)
            this.applyTheme();

            // Listen for system theme changes
            window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => {
                if (this.theme === "system") {
                    this.applyTheme();
                }
            });
        },

        setTheme(theme: "system" | "dark" | "light") {
            this.theme = theme;
            this.applyTheme();
        },

        applyTheme() {
            const rootEl = document.documentElement;
            if (this.isDark) {
                rootEl.classList.add("my-app-dark");
            } else {
                rootEl.classList.remove("my-app-dark");
            }
        },
    },
});
