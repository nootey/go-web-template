<script setup lang="ts">
import { useAuthStore } from "../../pkg/stores/auth_store.ts";

const authStore = useAuthStore();
</script>

<template>
    <span>Hello there ...</span>
    <i @click="authStore.logoutUser()">Log out</i>
</template>

<style scoped></style>
