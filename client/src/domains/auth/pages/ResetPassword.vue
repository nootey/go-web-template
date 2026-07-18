<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "../../../pkg/stores/auth_store.ts";
import { useToastStore } from "../../../pkg/stores/toast_store.ts";
import { useRegle } from "@regle/core";
import AuthSkeleton from "../../../pkg/components/AuthSkeleton.vue";
import ValidationError from "../../../pkg/components/ValidationError.vue";
import { minLength, required, sameAs } from "@regle/rules";

const authStore = useAuthStore();
const toastStore = useToastStore();

const router = useRouter();
const route = useRoute();

const token = ref<string>((route.query.token as string) ?? "");
const loading = ref<boolean>(false);

const form = ref({
    password: "",
    password_confirmation: "",
});

const { r$ } = useRegle(form, {
    password: {
        required,
        minLength: minLength(8),
    },
    password_confirmation: {
        required,
        sameAs: sameAs(() => form.value.password),
    },
});

async function resetPassword() {
    await r$.$validate();
    if (r$.$invalid) return;

    loading.value = true;
    try {
        const response = await authStore.resetPassword({
            token: token.value,
            password: form.value.password,
            password_confirmation: form.value.password_confirmation,
        });
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
            <div v-if="!token" class="text-center text-sm text-gray-600 dark:text-gray-400">
                This password reset link is invalid or incomplete.
            </div>

            <div v-else class="flex flex-col gap-3">
                <div class="flex flex-row w-full">
                    <div class="flex flex-col gap-1 w-full">
                        <ValidationError :is-required="true" :message="r$.password.$errors[0]">
                            <label>New password</label>
                        </ValidationError>
                        <InputText
                            id="password"
                            v-model="form.password"
                            type="password"
                            :placeholder="'New password'"
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
                            @keydown.enter="resetPassword"
                        />
                    </div>
                </div>

                <Button
                    :label="loading ? 'Updating...' : 'Update password'"
                    :icon="loading ? 'pi pi-spin pi-spinner mr-2' : ''"
                    class="w-full auth-accent-button"
                    :disabled="loading"
                    @click="resetPassword"
                />
            </div>

            <div class="flex items-center justify-center gap-2 mt-4 pt-3 border-t border-gray-200 dark:border-gray-700">
                <span
                    class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 cursor-pointer"
                    @click="router.push({ name: 'login' })"
                >
                    Back to log in
                </span>
            </div>
        </div>
    </AuthSkeleton>
</template>

<style scoped></style>
