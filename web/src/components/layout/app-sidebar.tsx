import { useParams } from "@tanstack/react-router";
import { LogOut, Menu } from "lucide-react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { useLogout, useMe } from "@/features/auth/api";
import { getInitials } from "@/lib/utils";
import { useUiStore } from "@/stores/ui-store";

import { FolderTree } from "./folder-tree";
import { SpaceSwitcher } from "./space-switcher";

function SidebarContent() {
  const { data: user } = useMe();
  const logout = useLogout();
  const params = useParams({ strict: false }) as { spaceId?: string };

  return (
    <div className="flex h-full flex-col">
      {/* Space switcher */}
      <SpaceSwitcher />
      <Separator />

      {/* Folder tree */}
      <ScrollArea className="flex-1 py-2">
        {params.spaceId ? (
          <FolderTree spaceId={params.spaceId} />
        ) : (
          <p className="px-3 py-2 text-xs text-muted-foreground">
            Select a space
          </p>
        )}
      </ScrollArea>

      {/* User footer */}
      <Separator />
      <div className="flex items-center gap-2 p-3">
        <Avatar className="size-8">
          <AvatarFallback className="text-xs">
            {user ? getInitials(user.name) : "?"}
          </AvatarFallback>
        </Avatar>
        <span className="flex-1 truncate text-sm">{user?.name ?? ""}</span>
        <Button
          variant="ghost"
          size="icon"
          className="size-7"
          onClick={() => logout.mutate()}
          aria-label="Sign out"
        >
          <LogOut className="size-4" />
        </Button>
      </div>
    </div>
  );
}

/** Desktop sidebar + mobile Sheet wrapper */
export function AppSidebar() {
  const sidebarOpen = useUiStore((s) => s.sidebarOpen);
  const setSidebarOpen = useUiStore((s) => s.setSidebarOpen);

  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden w-64 shrink-0 border-r md:block">
        <SidebarContent />
      </aside>

      {/* Mobile sheet */}
      <Sheet open={sidebarOpen} onOpenChange={setSidebarOpen}>
        <SheetContent side="left" className="w-64 p-0">
          <SheetTitle className="sr-only">Navigation</SheetTitle>
          <SidebarContent />
        </SheetContent>
      </Sheet>
    </>
  );
}

/** Hamburger button for mobile */
export function SidebarTrigger() {
  const setSidebarOpen = useUiStore((s) => s.setSidebarOpen);

  return (
    <Button
      variant="ghost"
      size="icon"
      className="md:hidden"
      onClick={() => setSidebarOpen(true)}
      aria-label="Open navigation"
    >
      <Menu className="size-5" />
    </Button>
  );
}
