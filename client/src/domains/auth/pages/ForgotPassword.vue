<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../../../pkg/stores/auth_store.ts";
import { useToastStore } from "../../../pkg/stores/toast_store.ts";
import { useRegle } from "@regle/core";
import AuthSkeleton from "../../../pkg/components/AuthSkeleton.vue";
import ValidationError from "../../../pkg/components/ValidationError.vue";
import { email, required } from "@regle/rules";

const authStore = useAuthStore();
const toastStore = useToastStore();

const router = useRouter();

const loading = ref<boolean>(false);

const form = ref({
    email: "",
});

const { r$ } = useRegle(form, {
    email: {
        required,
        email,
    },
});

async function requestPasswordReset() {
    await r$.$validate();
    if (r$.$invalid) return;

    loading.value = true;
    try {
        const response = await authStore.requestPasswordReset(form.value.email);
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
                            @keydown.enter="requestPasswordReset"
                        />
                    </div>
                </div>

                <Button
                    :label="loading ? 'Sending...' : 'Request password reset'"
                    :icon="loading ? 'pi pi-spin pi-spinner mr-2' : ''"
                    class="w-full"
                    :disabled="loading"
                    @click="requestPasswordReset"
                />
            </div>

            <div class="flex items-center justify-center gap-2 mt-4 pt-3 border-t border-edge">
                <span class="text-sm text-fg-muted">Remember your password?</span>
                <span
                    class="text-sm font-medium text-accent hover:underline underline-offset-4 cursor-pointer"
                    @click="router.push({ name: 'login' })"
                >
                    Log in
                </span>
            </div>
        </div>
    </AuthSkeleton>
</template>

<style scoped></style>
