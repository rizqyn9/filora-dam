import { Link } from "@tanstack/react-router";
import { LayoutGrid, List, Upload } from "lucide-react";

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { useBreadcrumbs } from "@/features/folders/api";
import { useUiStore } from "@/stores/ui-store";

interface ContentToolbarProps {
  spaceId: string;
  spaceName?: string;
  folderId?: string;
  onUploadClick?: () => void;
}

export function ContentToolbar({
  spaceId,
  spaceName,
  folderId,
  onUploadClick,
}: ContentToolbarProps) {
  const viewMode = useUiStore((s) => s.viewMode);
  const setViewMode = useUiStore((s) => s.setViewMode);
  const { data: crumbs } = useBreadcrumbs(folderId);

  return (
    <div className="flex items-center gap-2 border-b px-4 py-2">
      {/* Breadcrumbs */}
      <Breadcrumb className="flex-1">
        <BreadcrumbList>
          <BreadcrumbItem>
            {folderId ? (
              <BreadcrumbLink
                render={<Link to="/spaces/$spaceId" params={{ spaceId }} />}
              >
                {spaceName ?? "Root"}
              </BreadcrumbLink>
            ) : (
              <BreadcrumbPage>{spaceName ?? "Root"}</BreadcrumbPage>
            )}
          </BreadcrumbItem>

          {crumbs?.map((crumb, i) => {
            const isLast = i === crumbs.length - 1;
            return (
              <span key={crumb.id} className="contents">
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  {isLast ? (
                    <BreadcrumbPage>{crumb.name}</BreadcrumbPage>
                  ) : (
                    <BreadcrumbLink
                      render={
                        <Link
                          to="/spaces/$spaceId/folders/$folderId"
                          params={{ spaceId, folderId: crumb.id }}
                        />
                      }
                    >
                      {crumb.name}
                    </BreadcrumbLink>
                  )}
                </BreadcrumbItem>
              </span>
            );
          })}
        </BreadcrumbList>
      </Breadcrumb>

      {/* View toggle */}
      <ToggleGroup
        value={[viewMode]}
        onValueChange={(v) => {
          if (v.length) setViewMode(v[0] as "grid" | "list");
        }}
        size="sm"
      >
        <ToggleGroupItem value="grid" aria-label="Grid view">
          <LayoutGrid className="size-4" />
        </ToggleGroupItem>
        <ToggleGroupItem value="list" aria-label="List view">
          <List className="size-4" />
        </ToggleGroupItem>
      </ToggleGroup>

      {/* Upload button */}
      <Button
        size="sm"
        variant="outline"
        onClick={onUploadClick}
        disabled={!onUploadClick}
      >
        <Upload className="mr-1.5 size-4" />
        Upload
      </Button>
    </div>
  );
}
