import "@fontsource-variable/figtree";
import "primeicons/primeicons.css";
import "./pkg/style/global.css";

import { createApp } from "vue";
import App from "./App.vue";
import router from "./pkg/router/router.ts";
import { createPinia } from "pinia";

// PrimeVue core + theme
import PrimeVue from "primevue/config";
import { AppPreset } from "./pkg/style/theme.ts";

// PrimeVue services & directives
import ConfirmationService from "primevue/confirmationservice";
import Tooltip from "primevue/tooltip";
import Ripple from "primevue/ripple";
import ToastService from "primevue/toastservice";

const app = createApp(App);

// Plugins
app.use(createPinia());
app.use(router);
app.use(PrimeVue, {
    theme: {
        preset: AppPreset,
        options: {
            prefix: "p",
            darkModeSelector: ".app-dark",
            // Must match the layer order declared in global.css.
            cssLayer: { name: "primevue", order: "theme, base, primevue" },
        },
    },
    ripple: true,
    pt: {
        panel: {
            root: { class: "rounded-2xl" },
            title: { class: "text-lg tracking-tight" },
            contentWrapper: { class: "min-w-0" },
        },
        dialog: {
            title: { class: "tracking-tight" },
        },
    },
});

app.use(ToastService);
app.use(ConfirmationService);

// Directives
app.directive("tooltip", Tooltip);
app.directive("ripple", Ripple);

app.mount("#app");
