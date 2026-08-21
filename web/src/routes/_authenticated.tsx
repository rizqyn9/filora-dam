import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

import { AppSidebar, SidebarTrigger } from "@/components/layout/app-sidebar";
import { useAuthStore } from "@/features/auth/auth-store";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: () => {
    if (!useAuthStore.getState().token) {
      throw redirect({ to: "/login" });
    }
  },
  component: AuthenticatedLayout,
});

function AuthenticatedLayout() {
  return (
    <div className="flex h-screen">
      <AppSidebar />
      <main className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-12 items-center gap-2 border-b px-4 md:hidden">
          <SidebarTrigger />
        </header>
        <div className="flex-1 overflow-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
