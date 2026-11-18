import { defineStore } from "pinia";
import apiClient from "../api/api_client.ts";
import type { User } from "../../domains/user/models.ts";
import type { AuthForm, SignupStatus } from "../../domains/auth/models.ts";
import router from "../router/router.ts";

export const useAuthStore = defineStore("auth", {
    state: () => ({
        apiPrefix: "auth",
        authenticated: localStorage.getItem("authenticated") == "true",
        user: null as User | null,
        initialized: false,
    }),
    getters: {
        isAuthenticated: (s) => s.authenticated,
        isInitialized: (s) => s.initialized,
        isValidated: (s) => !!s.user?.email_confirmed,
        isAdmin: (s) => s.user?.role?.name == "super-admin" || s.user?.role?.name == "admin",
        isSuperAdmin: (s) => s.user?.role?.name == "super-admin",
    },
    actions: {
        async login(form: AuthForm) {
            const response = await apiClient.post(`${this.apiPrefix}/login`, form);
            await this.init();
            return response;
        },

        async signupAllowed(): Promise<boolean> {
            const response = await apiClient.get<SignupStatus>(`${this.apiPrefix}/signup`);
            return response.data.allow_signup;
        },

        async signUp(form: AuthForm, invitation_id: number | null = null) {
            return await apiClient.post(`${this.apiPrefix}/register`, {
                ...form,
                ...(invitation_id && { invitation_id }),
            });
        },

        async resendConfirmationEmail(email?: string) {
            return await apiClient.post(`${this.apiPrefix}/resend-confirmation`, { email: email });
        },

        async requestPasswordReset(email?: string) {
            return await apiClient.post(`${this.apiPrefix}/request-password-reset`, {
                email: email,
            });
        },

        async resetPassword(payload: { token: string; password: string; password_confirmation: string }) {
            return await apiClient.post(`${this.apiPrefix}/reset-password`, payload);
        },

        async getAuthUser(set = true) {
            const response = await apiClient.get(`${this.apiPrefix}/me`, {
                params: { withSecrets: true },
            });

            if (set) {
                if (!response.data) {
                    await this.logoutUser();
                } else {
                    this.setUser(response.data);
                }
            }

            return response.data;
        },

        setUser(userData: User) {
            this.user = userData;
        },

        async logoutUser() {
            await apiClient.post(`${this.apiPrefix}/logout`, null);
            this.logout();
        },

        logout() {
            this.user = null;
            this.setAuthenticated(false);
            this.setInitialized(null);

            sessionStorage.clear();

            router.push("/login").then();
        },

        setAuthenticated(status: boolean) {
            this.authenticated = status;
            localStorage.setItem("authenticated", status.toString());
        },

        setInitialized(user: User | null) {
            this.initialized = user !== null;
        },

        async init() {
            try {
                this.setAuthenticated(true);
                const user = await this.getAuthUser(true);
                if (user) {
                    this.setInitialized(user);
                } else {
                    this.setAuthenticated(false);
                    this.setInitialized(null);
                }
            } catch {
                this.setAuthenticated(false);
                this.setInitialized(null);
            }
        },
    },
});
