import Aura from "@primeuix/themes/aura";
import { definePreset } from "@primeuix/themes";

const STEPS = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950] as const;

function palette(name: string): Record<number, string> {
    return Object.fromEntries(STEPS.map((step) => [step, `{${name}.${step}}`]));
}

const GOPHER_BLUE = {
    50: "#f3fbfe",
    100: "#e3f4fb",
    200: "#c2e8f7",
    300: "#89d6f2",
    400: "#40c0e9",
    500: "#00add8",
    600: "#008db0",
    700: "#03708d",
    800: "#025e77",
    900: "#0c4e62",
    950: "#073240",
};

export const AppPreset = definePreset(Aura, {
    primitive: {
        borderRadius: {
            none: "0",
            xs: "3px",
            sm: "5px",
            md: "8px",
            lg: "10px",
            xl: "14px",
        },
        gopher: GOPHER_BLUE,
    },
    semantic: {
        primary: palette("gopher"),
        focusRing: { width: "2px", style: "solid", offset: "2px" },
        colorScheme: {
            light: {
                // White text fails contrast on light blue, so buttons use dark text.
                primary: { contrastColor: "{surface.950}" },
                surface: { 0: "#ffffff", ...palette("stone") },
                formField: {
                    background: "{surface.0}",
                    borderColor: "{surface.200}",
                    hoverBorderColor: "{surface.400}",
                },
                content: { borderColor: "{surface.200}" },
            },
            dark: {
                surface: { 0: "#ffffff", ...palette("neutral") },
                formField: {
                    background: "{surface.950}",
                    borderColor: "{surface.700}",
                    hoverBorderColor: "{surface.500}",
                },
                content: { background: "{surface.900}", borderColor: "{surface.800}" },
                overlay: {
                    popover: { background: "{surface.900}", borderColor: "{surface.800}" },
                    modal: { background: "{surface.900}", borderColor: "{surface.800}" },
                    select: { background: "{surface.900}", borderColor: "{surface.800}" },
                },
            },
        },
    },
    components: {
        panel: {
            header: { padding: "1.25rem 1.5rem 0.75rem" },
            content: { padding: "0.5rem 1.5rem 1.5rem" },
            title: { fontWeight: "600" },
        },
        datatable: {
            headerCell: { padding: "0.75rem 1rem" },
            bodyCell: { padding: "0.8rem 1rem" },
            columnTitle: { fontWeight: "500" },
            footer: { background: "transparent", borderWidth: "0", padding: "0.75rem 1rem" },
            colorScheme: {
                light: {
                    headerCell: { background: "transparent", color: "{surface.500}" },
                    header: { background: "transparent" },
                    row: { background: "transparent" },
                },
                dark: {
                    headerCell: { background: "transparent", color: "{surface.400}" },
                    header: { background: "transparent" },
                    row: { background: "transparent" },
                },
            },
        },
        paginator: {
            root: { background: "transparent", padding: "0" },
        },
        dialog: {
            header: { padding: "1.5rem 1.5rem 1rem" },
            content: { padding: "0 1.5rem 1.5rem" },
            title: { fontSize: "1.25rem", fontWeight: "600" },
        },
        drawer: {
            colorScheme: {
                dark: { root: { background: "{surface.900}" } },
            },
        },
        tooltip: {
            colorScheme: {
                light: { root: { background: "{surface.900}", color: "{surface.0}" } },
                dark: { root: { background: "{surface.100}", color: "{surface.950}" } },
            },
        },
    },
});
