<script setup lang="ts">
import { useRouter } from "vue-router";
import { storeToRefs } from "pinia";
import { useAuthStore } from "../../pkg/stores/auth_store.ts";

const router = useRouter();
const authStore = useAuthStore();

const { isAuthenticated } = storeToRefs(authStore);

function handlePrimary() {
    router.push({ name: isAuthenticated.value ? "dashboard" : "login" });
}
</script>

<template>
    <div class="min-h-screen w-full flex flex-col items-center justify-center text-center p-6">
        <div class="flex flex-col gap-3" role="region" aria-label="Page not found">
            <h1 ref="headingRef" tabindex="-1" class="text-3xl font-bold">404 Page not found</h1>
            <p class="text-gray-600 dark:text-gray-400">The page you're looking for doesn't exist or was moved.</p>
            <div class="flex flex-row gap-3 justify-center">
                <Button
                    :label="isAuthenticated ? 'Go Home' : 'Login'"
                    class="main-button w-40"
                    @click="handlePrimary"
                />
            </div>
        </div>
    </div>
</template>
