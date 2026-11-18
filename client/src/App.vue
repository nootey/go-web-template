<script setup lang="ts">
import { computed, onMounted } from "vue";
import { storeToRefs } from "pinia";
import { useRoute } from "vue-router";
import { useAuthStore } from "./pkg/stores/auth_store.ts";
import { useThemeStore } from "./pkg/stores/theme_store.ts";

const authStore = useAuthStore();
const themeStore = useThemeStore();
const route = useRoute();

themeStore.initializeTheme();

const { isAuthenticated, isInitialized } = storeToRefs(authStore);

const requiresAuthView = computed<boolean>(() => route.matched.some((r) => r.meta.requiresAuth));

onMounted(async () => {
    if (isAuthenticated.value) {
        await authStore.init();
    }
});
</script>

<template>
    <Toast position="bottom-center" group="bc" />
    <Toast position="bottom-right" group="br" />
    <ConfirmDialog unstyled>
        <template #container="{ message, acceptCallback, rejectCallback }">
            <div class="p-overlay-mask p-overlay-mask-enter flex items-center justify-center">
                <div
                    class="flex w-[min(420px,calc(100vw-2rem))] flex-col gap-4 rounded-2xl border border-edge bg-panel p-6 shadow-2xl"
                >
                    <div class="text-lg font-semibold tracking-tight text-fg">
                        {{ message.header }}
                    </div>
                    <div class="whitespace-pre-line text-sm leading-relaxed text-fg-muted">
                        {{ message.message }}
                    </div>
                    <div class="flex justify-end gap-2 pt-2">
                        <Button
                            :label="message.rejectProps?.label || 'Cancel'"
                            severity="secondary"
                            variant="outlined"
                            @click="rejectCallback"
                        />
                        <Button
                            :label="message.acceptProps?.label || 'Confirm'"
                            :severity="message.acceptProps?.severity"
                            @click="acceptCallback"
                        />
                    </div>
                </div>
            </div>
        </template>
    </ConfirmDialog>

    <div class="flex min-h-screen flex-col">
        <main class="flex flex-1 flex-col">
            <div v-if="requiresAuthView && !isInitialized" class="flex flex-1 items-center justify-center">
                <i class="pi pi-spin pi-spinner text-2xl text-fg-muted" />
            </div>
            <router-view v-else />
        </main>
    </div>
</template>
