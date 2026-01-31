import { defineStore } from "pinia";
import { useToast } from "primevue/usetoast";
import axios from "axios";

export const useToastStore = defineStore("toast", () => {
    const toast = useToast();
    const isMobile = window.matchMedia("(max-width: 768px)").matches;
    const group = isMobile ? "bc" : "br";

    const errorResponseToast = (error: unknown) => {
        console.error("triggered error", error);

        let summary = "Unexpected Error";
        let detail = "An unknown error occurred.";

        if (axios.isAxiosError(error)) {
            const data = error.response?.data as { title?: string; message?: string } | undefined;

            if (data?.title || data?.message) {
                summary = data.title ?? "Error";
                detail = data.message ?? "Something went wrong.";
            }

            if (!error.response || error.code === "ERR_NETWORK" || error.message === "Network Error") {
                summary = "Server unreachable";
                detail = "The server is currently not reachable.";
            }

            if (
                (!data?.message || detail === "Something went wrong.") &&
                error.message &&
                error.message !== "Request failed with status code 500"
            ) {
                detail = error.message;
            }
        } else if (error instanceof Error) {
            detail = error.message;
        }

        toast.add({
            severity: "error",
            summary,
            detail,
            life: isMobile ? 2500 : 5000,
            group,
        });
    };

    type ToastResponse = {
        data?: {
            title?: string;
            message?: string;
        };
        title?: string;
        message?: string;
    };

    const successResponseToast = (response: ToastResponse) => {
        const data = response?.data || response;
        if (data?.title || data?.message) {
            toast.add({
                severity: "success",
                summary: data.title ?? "Success",
                detail: data.message ?? "",
                life: isMobile ? 1500 : 3000,
                group,
            });
        }
    };

    const infoResponseToast = (response: ToastResponse) => {
        const data = response?.data || response;
        if (data?.title || data?.message) {
            toast.add({
                severity: "info",
                summary: data.title ?? "Info",
                detail: data.message ?? "",
                life: isMobile ? 1000 : 2000,
                group,
            });
        }
    };

    const createInfoToast = (title: string, msg: string) => {
        if (title && msg) {
            toast.add({
                severity: "info",
                summary: title,
                detail: msg,
                life: isMobile ? 1000 : 2000,
                group,
            });
        }
    };

    return {
        errorResponseToast,
        successResponseToast,
        infoResponseToast,
        createInfoToast,
    };
});
