export interface Role {
  id: number;
  slug: string;
  name: string;
  description: string | null;
  isSystem: boolean;
  memberCount: number;
  permissions: string[];
}
