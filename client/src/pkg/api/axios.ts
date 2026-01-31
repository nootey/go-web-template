import axios from "axios";
import { useAuthStore } from "../stores/auth_store.ts";

const apiClient = axios.create({
    baseURL: "/api",
    withCredentials: true,
});

let isRefreshing = false;
let failedQueue: Array<{
    resolve: (value?: unknown) => void;
    reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: unknown) => {
    failedQueue.forEach((prom) => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve();
        }
    });
    failedQueue = [];
};

apiClient.interceptors.response.use(
    (response) => response,
    async (error) => {
        const { config, response } = error;
        const url: string = error?.config?.url ?? "";

        if (!response) return Promise.reject(error);

        if (response.status === 401 && !config._retry) {
            if (isRefreshing) {
                // Queue this request
                return new Promise((resolve, reject) => {
                    failedQueue.push({ resolve, reject });
                })
                    .then(() => apiClient(config))
                    .catch((err) => Promise.reject(err));
            }

            config._retry = true;
            isRefreshing = true;

            try {
                // First request triggers refresh, others wait
                await apiClient(config);
                processQueue(null);
                isRefreshing = false;
                return await apiClient(config);
            } catch (retryError) {
                processQueue(retryError);
                isRefreshing = false;
                throw retryError;
            }
        }

        if (response.status === 401) {
            const auth = useAuthStore();
            const isAuthEndpoint = /\/auth\/(current|logout|login)/.test(url);

            if (!isAuthEndpoint && auth.isAuthenticated) {
                await auth.logoutUser();
            }
        }

        return Promise.reject(error);
    },
);

export default apiClient;
