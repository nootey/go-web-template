<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "../../../pkg/stores/auth_store.ts";
import { useToastStore } from "../../../pkg/stores/toast_store.ts";
import type { AuthForm } from "../models.ts";
import { useRegle } from "@regle/core";
import AuthSkeleton from "../../../pkg/components/AuthSkeleton.vue";
import ValidationError from "../../../pkg/components/ValidationError.vue";
import { email, required } from "@regle/rules";

const authStore = useAuthStore();
const toastStore = useToastStore();

const router = useRouter();
const route = useRoute();

const loading = ref<boolean>(false);

const form = ref<AuthForm>({
    email: "",
    password: "",
    remember_me: false,
});

const { r$ } = useRegle(form, {
    email: {
        required,
        email,
    },
    password: {
        required,
    },
});

function resolveRedirect(): string {
    const q = route.query.redirect as string | string[] | undefined;
    const redirect = Array.isArray(q) ? q[0] : q;

    if (typeof redirect !== "string") return "/";

    // Disallow absolute URLs or protocol-relative
    if (/^https?:\/\//i.test(redirect) || redirect.startsWith("//")) return "/";

    // Allow only root-relative paths
    if (!redirect.startsWith("/")) return "/";

    // Avoid looping back to login
    if (redirect === "/login") return "/";

    return redirect;
}

async function login() {
    await r$.$validate();
    if (r$.$invalid) return;

    loading.value = true;
    try {
        await authStore.login(form.value);

        if (authStore.authenticated) {
            const target = resolveRedirect();
            await router.replace(target);
        }
    } catch (error) {
        toastStore.errorResponseToast(error);
    } finally {
        loading.value = false;
    }
}

function signUp() {
    console.log("signUp");
    // router.push({ name: "sign.up" });
}

function forgotPassword() {
    console.log("forgotPassword");
    // router.push({ name: "forgot.password" });
}
</script>

<template>
    <AuthSkeleton>
        <div class="w-full max-w-md mx-auto px-3 sm:px-0">
            <div class="flex flex-col gap-3">
                <div class="flex flex-row w-full">
                    <div class="flex flex-col gap-1 w-full">
                        <ValidationError :is-required="true" :message="r$.email.$errors[0]">
                            <label>Email</label>
                        </ValidationError>
                        <InputText
                            id="email"
                            v-model="form.email"
                            type="email"
                            :placeholder="'Email'"
                            class="w-full rounded-xl"
                        />
                    </div>
                </div>

                <div class="flex flex-row w-full">
                    <div class="flex flex-col gap-1 w-full">
                        <ValidationError :is-required="true" :message="r$.password.$errors[0]">
                            <label>Password</label>
                        </ValidationError>
                        <InputText
                            id="password"
                            v-model="form.password"
                            type="password"
                            :placeholder="'Password'"
                            class="w-full rounded-xl"
                            @keydown.enter="login"
                        />
                    </div>
                </div>

                <div class="flex flex-row w-full justify-between">
                    <div class="flex flex-row items-center gap-2">
                        <Checkbox v-model="form.remember_me" input-id="rememberMe" :binary="true" class="scale-90" />
                        <label for="rememberMe" class="text-sm cursor-pointer text-gray-600 dark:text-gray-400">
                            Remember me
                        </label>
                    </div>

                    <span
                        class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 cursor-pointer"
                        @click="forgotPassword"
                    >
                        Forgot password?
                    </span>
                </div>

                <Button
                    :label="loading ? 'Signing in...' : 'Sign in'"
                    :icon="loading ? 'pi pi-spin pi-spinner mr-2' : ''"
                    class="w-full auth-accent-button"
                    :disabled="loading"
                    @click="login"
                />
            </div>

            <div class="flex items-center justify-center gap-2 mt-4 pt-3 border-t border-gray-200 dark:border-gray-700">
                <span class="text-sm text-gray-600 dark:text-gray-400">Don't have an account?</span>
                <span
                    class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 cursor-pointer"
                    @click="signUp"
                >
                    Create account
                </span>
            </div>
        </div>
    </AuthSkeleton>
</template>

<style scoped></style>
