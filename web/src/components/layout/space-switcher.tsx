import { useNavigate, useParams } from "@tanstack/react-router";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useSpaces } from "@/features/spaces/api";

export function SpaceSwitcher() {
  const { data: spaces, isLoading } = useSpaces();
  const params = useParams({ strict: false }) as { spaceId?: string };
  const navigate = useNavigate();

  if (isLoading || !spaces?.length) {
    return (
      <div className="p-4">
        <p className="text-sm text-muted-foreground">
          {isLoading ? "Loading..." : "No spaces"}
        </p>
      </div>
    );
  }

  const activeId = params.spaceId ?? spaces[0].id;

  return (
    <div className="p-3">
      <Select
        value={activeId}
        onValueChange={(id) => {
          if (id) navigate({ to: "/spaces/$spaceId", params: { spaceId: id } });
        }}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder="Select space" />
        </SelectTrigger>
        <SelectContent>
          {spaces.map((space) => (
            <SelectItem key={space.id} value={space.id}>
              {space.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
