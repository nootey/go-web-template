<script setup lang="ts"></script>

<template>
    <div class="min-h-screen w-full flex items-center justify-center p-4 bg-gray-100 dark:bg-gray-900">
        <Transition appear name="auth-animation">
            <div
                class="w-full max-w-md bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl shadow-lg p-8"
            >
                <div class="mb-8 text-center">
                    <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">Client</h1>
                    <p class="text-gray-600 dark:text-gray-400">Welcome back</p>
                </div>

                <slot />
            </div>
        </Transition>
    </div>
</template>

<style scoped>
.auth-animation-enter-active,
.auth-animation-leave-active {
    transition: all 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.auth-animation-enter-from {
    opacity: 0;
    transform: translateY(30px) scale(0.95);
}

.auth-animation-leave-to {
    opacity: 0;
    transform: translateY(-30px) scale(0.95);
}
</style>
