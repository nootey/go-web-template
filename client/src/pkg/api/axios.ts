import axios from "axios";
import { useAuthStore } from "../stores/auth_store.ts";
import type { AxiosError, AxiosResponse } from "axios";

const apiClient = axios.create({
    baseURL: "/api",
    withCredentials: true,
});

// A 401 from these is an expected outcome the caller handles itself (bad
// credentials, an already-dead session), not a signal to tear down auth state.
const authEndpointPattern = /\/auth\/(login|register|logout|me)/;

apiClient.interceptors.response.use(
    (response: AxiosResponse): AxiosResponse => response,
    async (error: AxiosError): Promise<never> => {
        const url: string = error.config?.url ?? "";

        if (error.response?.status === 401 && !authEndpointPattern.test(url)) {
            const auth = useAuthStore();
            if (auth.isAuthenticated) {
                await auth.logoutUser();
            }
        }

        return Promise.reject(error);
    },
);

export default apiClient;
