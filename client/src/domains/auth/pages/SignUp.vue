<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../../../pkg/stores/auth_store.ts";
import { useToastStore } from "../../../pkg/stores/toast_store.ts";
import type { AuthForm } from "../models.ts";
import { useRegle } from "@regle/core";
import AuthSkeleton from "../../../pkg/components/AuthSkeleton.vue";
import ValidationError from "../../../pkg/components/ValidationError.vue";
import { email, minLength, required, sameAs } from "@regle/rules";

const authStore = useAuthStore();
const toastStore = useToastStore();

const router = useRouter();

const loading = ref<boolean>(false);

const form = ref<AuthForm>({
    display_name: "",
    email: "",
    password: "",
    password_confirmation: "",
});

const { r$ } = useRegle(form, {
    display_name: {
        required,
    },
    email: {
        required,
        email,
    },
    password: {
        required,
        minLength: minLength(8),
    },
    password_confirmation: {
        required,
        sameAs: sameAs(() => form.value.password),
    },
});

async function signUp() {
    await r$.$validate();
    if (r$.$invalid) return;

    loading.value = true;
    try {
        const response = await authStore.signUp(form.value);
        toastStore.successResponseToast(response);
        await router.push({ name: "login" });
    } catch (error) {
        toastStore.errorResponseToast(error);
    } finally {
        loading.value = false;
    }
}
</script>

<template>
    <AuthSkeleton>
        <div class="w-full max-w-md mx-auto px-3 sm:px-0">
            <div class="flex flex-col gap-3">
                <div class="flex flex-row w-full">
                    <div class="flex flex-col gap-1 w-full">
                        <ValidationError :is-required="true" :message="r$.display_name.$errors[0]">
                            <label>Display name</label>
                        </ValidationError>
                        <InputText
                            id="display_name"
                            v-model="form.display_name"
                            type="text"
                            :placeholder="'Display name'"
                            :disabled="loading"
                            class="w-full rounded-xl"
                        />
                    </div>
                </div>

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
                            :disabled="loading"
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
                            :disabled="loading"
                            class="w-full rounded-xl"
                        />
                    </div>
                </div>

                <div class="flex flex-row w-full">
                    <div class="flex flex-col gap-1 w-full">
                        <ValidationError :is-required="true" :message="r$.password_confirmation.$errors[0]">
                            <label>Confirm password</label>
                        </ValidationError>
                        <InputText
                            id="password_confirmation"
                            v-model="form.password_confirmation"
                            type="password"
                            :placeholder="'Confirm password'"
                            :disabled="loading"
                            class="w-full rounded-xl"
                            @keydown.enter="signUp"
                        />
                    </div>
                </div>

                <Button
                    :label="loading ? 'Creating account...' : 'Sign up'"
                    :icon="loading ? 'pi pi-spin pi-spinner mr-2' : ''"
                    class="w-full auth-accent-button"
                    :disabled="loading"
                    @click="signUp"
                />
            </div>

            <div class="flex items-center justify-center gap-2 mt-4 pt-3 border-t border-gray-200 dark:border-gray-700">
                <span class="text-sm text-gray-600 dark:text-gray-400">Already have an account?</span>
                <span
                    class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 cursor-pointer"
                    @click="router.push({ name: 'login' })"
                >
                    Log in
                </span>
            </div>
        </div>
    </AuthSkeleton>
</template>

<style scoped></style>
