import type { LucideIcon } from "lucide-react";
import {
  Database,
  HardDrive,
  Images,
  LayoutDashboard,
  Shield,
  Tags,
  Users,
} from "lucide-react";

export interface NavItem {
  title: string;
  to: string;
  icon: LucideIcon;
}

export interface NavGroup {
  label: string;
  items: NavItem[];
}

/**
 * Single source of truth for sidebar navigation. Add a route once here and it
 * shows up in the sidebar (and can drive breadcrumbs later).
 */
export const navGroups: NavGroup[] = [
  {
    label: "Workspace",
    items: [
      { title: "Dashboard", to: "/", icon: LayoutDashboard },
      { title: "Galleries", to: "/galleries", icon: Images },
      { title: "Tags", to: "/tags", icon: Tags },
    ],
  },
  {
    label: "Administration",
    items: [
      { title: "Admin", to: "/admin", icon: Shield },
      { title: "Users", to: "/admin/users", icon: Users },
      { title: "Roles", to: "/admin/roles", icon: Database },
      { title: "Storage", to: "/storage", icon: HardDrive },
    ],
  },
];
