import { useNavigate } from "@tanstack/react-router";
import { Bell, LogOut, Search, Settings, User } from "lucide-react";

import { ModeToggle } from "@/components/mode-toggle";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { getInitials } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

// Presentational placeholder — swap for `useMe()` once auth is wired.
const currentUser = { name: "Rizqy Nugroho", email: "rizqy@filora.app" };

export function AppHeader() {
  const navigate = useNavigate();
  const clear = useAuthStore((s) => s.clear);

  const logout = () => {
    clear();
    void navigate({ to: "/login" });
  };

  return (
    <header className="sticky top-0 z-10 flex h-14 items-center gap-2 border-b bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <SidebarTrigger className="-ml-1" />
      <Separator orientation="vertical" className="mr-1 h-4" />

      <div className="relative hidden max-w-sm flex-1 md:block">
        <Search className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input placeholder="Search assets, galleries..." className="h-9 pl-8" />
      </div>

      <div className="ml-auto flex items-center gap-1">
        <ModeToggle />
        <Button variant="ghost" size="icon" aria-label="Notifications">
          <Bell className="size-4" />
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="ml-1 h-9 gap-2 px-1.5 md:px-2">
              <Avatar className="size-7">
                <AvatarFallback className="text-xs">
                  {getInitials(currentUser.name)}
                </AvatarFallback>
              </Avatar>
              <span className="hidden text-sm font-medium md:inline">
                {currentUser.name}
              </span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel className="flex flex-col">
              <span>{currentUser.name}</span>
              <span className="text-xs font-normal text-muted-foreground">
                {currentUser.email}
              </span>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <User className="size-4" />
              Profile
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Settings className="size-4" />
              Settings
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem variant="destructive" onClick={logout}>
              <LogOut className="size-4" />
              Log out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
