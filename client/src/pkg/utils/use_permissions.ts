import { useAuthStore } from "../stores/auth_store.ts";

export function usePermissions() {
    const auth = useAuthStore();

    const hasRole = (role: string) => {
        const name = auth.user?.role?.name;
        return name === "super-admin" || name === role;
    };

    const hasPermission = (perms: string | string[]) => {
        const need = Array.isArray(perms) ? perms : [perms];
        const granted = auth.user?.role?.permissions?.map((p) => p.name) ?? [];
        if (granted.includes("root_access")) return true;
        return need.every((n) => granted.includes(n));
    };

    return { hasRole, hasPermission };
}
