<script setup lang="ts"></script>

<template>
    <div
        class="relative flex min-h-screen w-full flex-col items-center justify-center overflow-hidden bg-page px-4 py-12"
    >
        <Transition appear name="auth-animation">
            <div class="relative w-full max-w-md rounded-2xl border border-edge bg-panel p-6 shadow-sm sm:p-8">
                <div class="mb-6 flex flex-col items-center gap-2 text-center">
                    <h1 class="text-2xl font-semibold tracking-tight text-fg">MyApp</h1>
                    <p class="text-sm text-fg-muted">Welcome back</p>
                </div>

                <slot />
            </div>
        </Transition>
    </div>
</template>

<style scoped>
.auth-animation-enter-active {
    transition:
        opacity 0.4s ease,
        transform 0.4s ease;
}

.auth-animation-enter-from {
    opacity: 0;
    transform: translateY(8px);
}
</style>
