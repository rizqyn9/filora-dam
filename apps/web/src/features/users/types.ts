export type UserStatus = "active" | "invited" | "suspended";

export interface ManagedUser {
  id: number;
  name: string;
  email: string;
  role: string;
  status: UserStatus;
  lastSeen: string | null;
  createdAt: string;
}
