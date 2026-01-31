import type { RouteRecordRaw } from "vue-router";
import NotFound from "../../domains/index/NotFound.vue";
import Dashboard from "../../domains/index/Dashboard.vue";
import Login from "../../domains/auth/pages/Login.vue";

declare module "vue-router" {
    interface RouteMeta {
        requiresAuth?: boolean;
        guestOnly?: boolean;
        emailConfirmed?: boolean;
        permsAny?: string[]; // allow if the user has ANY of these
        permsAll?: string[]; // allow only if the user has ALL of these
    }
}

const routes: RouteRecordRaw[] = [
    {
        path: "/",
        name: "dashboard",
        meta: { title: "Dash", requiresAuth: true },
        component: Dashboard,
    },
    {
        path: "/login",
        name: "login",
        meta: { title: "Login", guestOnly: true, hideNavigation: true },
        component: Login,
    },
    {
        path: "/:pathMatch(.*)*",
        name: "NotFound",
        component: NotFound,
        meta: { title: "404" },
    },
];

export default routes;
