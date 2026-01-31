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
    <Toast position="top-center" group="bc" />
    <ConfirmDialog unstyled>
        <template #container="{ message, acceptCallback, rejectCallback }">
            <div class="flex justify-center items-center p-overlay-mask p-overlay-mask-enter">
                <div class="flex flex-col p-5 gap-4 rounded-lg bg-white dark:bg-gray-800 shadow-lg">
                    <div class="font-bold text-xl text-gray-900 dark:text-gray-100">
                        {{ message.header }}
                    </div>
                    <div class="text-gray-900 dark:text-gray-100" v-html="message.message.replace(/\n/g, '<br>')" />
                    <div class="flex justify-end gap-2">
                        <Button
                            class="p-2 rounded-lg text-gray-900 dark:text-gray-100 border-gray-900 dark:border-gray-100"
                            :label="message.rejectProps?.label || 'Cancel'"
                            variant="outlined"
                            @click="rejectCallback"
                        />
                        <Button
                            class="p-2 rounded-lg text-white"
                            :label="message.acceptProps?.label || 'Confirm'"
                            :severity="message.acceptProps?.severity"
                            @click="acceptCallback"
                        />
                    </div>
                </div>
            </div>
        </template>
    </ConfirmDialog>

    <div id="app" class="flex flex-col min-h-screen">
        <div class="flex-1 app-content">
            <div v-if="requiresAuthView && !isInitialized" class="w-full h-full flex items-center justify-center">
                <i class="pi pi-spin pi-spinner text-2xl" />
            </div>
            <div v-else>
                <router-view />
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
#app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
}
</style>
