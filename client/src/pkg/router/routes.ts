import type { RouteRecordRaw } from "vue-router";
import NotFound from "../../domains/index/NotFound.vue";
import Dashboard from "../../domains/index/Dashboard.vue";
import Login from "../../domains/auth/pages/Login.vue";
import SignUp from "../../domains/auth/pages/SignUp.vue";
import ForgotPassword from "../../domains/auth/pages/ForgotPassword.vue";
import ResetPassword from "../../domains/auth/pages/ResetPassword.vue";
import ConfirmEmail from "../../domains/auth/pages/ConfirmEmail.vue";

declare module "vue-router" {
    interface RouteMeta {
        title?: string;
        hideNavigation?: boolean;
        requiresAuth?: boolean;
        guestOnly?: boolean;
        requiresSignup?: boolean; // open only when the server allows sign up
        emailConfirmed?: boolean;
        permsAny?: string[]; // allow if the user has ANY of these
        permsAll?: string[]; // allow only if the user has ALL of these
    }
}

const routes: RouteRecordRaw[] = [
    {
        path: "/",
        name: "dashboard",
        meta: { title: "Dashboard", requiresAuth: true },
        component: Dashboard,
    },
    {
        path: "/login",
        name: "login",
        meta: { title: "Login", guestOnly: true, hideNavigation: true },
        component: Login,
    },
    {
        path: "/signup",
        name: "sign.up",
        meta: { title: "Sign up", guestOnly: true, requiresSignup: true, hideNavigation: true },
        component: SignUp,
    },
    {
        path: "/forgot-password",
        name: "forgot.password",
        meta: { title: "Forgot password", guestOnly: true, hideNavigation: true },
        component: ForgotPassword,
    },
    {
        path: "/reset-password",
        name: "reset.password",
        meta: { title: "Reset password", hideNavigation: true },
        component: ResetPassword,
    },
    {
        path: "/confirm-email",
        name: "confirm.email",
        meta: { title: "Confirm email", hideNavigation: true },
        component: ConfirmEmail,
    },
    {
        path: "/:pathMatch(.*)*",
        name: "NotFound",
        component: NotFound,
        meta: { title: "Not found" },
    },
];

export default routes;
