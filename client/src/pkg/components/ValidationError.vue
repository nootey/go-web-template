<script setup lang="ts">
import { computed, type Ref, unref } from "vue";

const props = defineProps<{
    message?: string | Ref<string>;
    isRequired?: boolean;
}>();

const isDisplayed = computed(() => Boolean(unref(props.message)));

const displayMessage = computed(() => {
    const msg = unref(props.message);
    return msg?.replace("Value", ": field") ?? "";
});
</script>

<template>
    <div class="flex flex-row align-items-center gap-1" :class="[isDisplayed ? 'invalid' : '']">
        <div class="flex flex-column label align-items-center">
            <slot />
        </div>
        <small v-show="!isDisplayed && props.isRequired" class="invalid disclaimer">*</small>
        <Transition name="slide-fade">
            <span v-if="isDisplayed" class="text-xs">
                {{ displayMessage }}
            </span>
        </Transition>
    </div>
</template>

<style scoped>
.invalid {
    color: #dd0025;
}

.disclaimer {
    font-size: 0.8rem;
    font-style: italic;
}

.slide-fade-enter-active {
    transition:
        opacity 0.4s ease,
        transform 0.4s ease;
}
.slide-fade-enter-from {
    opacity: 0;
    transform: translateX(20px);
}
.slide-fade-enter-to {
    opacity: 1;
    transform: translateX(0);
}

.slide-fade-leave-active {
    transition:
        opacity 0.4s ease,
        transform 0.4s ease;
}
.slide-fade-leave-from {
    opacity: 1;
    transform: translateX(0);
}
.slide-fade-leave-to {
    opacity: 0;
    transform: translateX(20px);
}
</style>
