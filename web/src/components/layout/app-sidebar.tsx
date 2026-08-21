import { LogOut, Menu } from "lucide-react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { useLogout, useMe } from "@/features/auth/api";
import { getInitials } from "@/lib/utils";
import { useUiStore } from "@/stores/ui-store";

function SidebarContent() {
  const { data: user } = useMe();
  const logout = useLogout();

  return (
    <div className="flex h-full flex-col">
      {/* Space switcher placeholder — wired in ticket 04 */}
      <div className="border-b p-4">
        <p className="text-sm font-medium text-muted-foreground">Spaces</p>
      </div>

      {/* Folder tree placeholder — wired in ticket 05 */}
      <ScrollArea className="flex-1 p-2">
        <p className="p-2 text-xs text-muted-foreground">Folders</p>
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
