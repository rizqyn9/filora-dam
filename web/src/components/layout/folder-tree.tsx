import { Link, useParams } from "@tanstack/react-router";
import { ChevronRight, Folder as FolderIcon } from "lucide-react";
import { useMemo, useState } from "react";

import { useFolders } from "@/features/folders/api";
import { buildFolderTree, type FolderNode } from "@/features/folders/schemas";
import { cn } from "@/lib/utils";

export function FolderTree({ spaceId }: { spaceId: string }) {
  const { data: folders } = useFolders(spaceId);
  const tree = useMemo(() => buildFolderTree(folders ?? []), [folders]);

  if (!folders?.length) {
    return (
      <p className="px-3 py-2 text-xs text-muted-foreground">No folders</p>
    );
  }

  return (
    <ul className="space-y-0.5 px-1">
      {tree.map((node) => (
        <TreeItem key={node.id} node={node} spaceId={spaceId} depth={0} />
      ))}
    </ul>
  );
}

function TreeItem({
  node,
  spaceId,
  depth,
}: {
  node: FolderNode;
  spaceId: string;
  depth: number;
}) {
  const params = useParams({ strict: false }) as { folderId?: string };
  const isActive = params.folderId === node.id;
  const hasChildren = node.children.length > 0;

  // Auto-expand if this node is an ancestor of the active folder
  const isAncestor = useMemo(() => {
    if (!params.folderId) return false;
    return containsDescendant(node, params.folderId);
  }, [node, params.folderId]);

  const [expanded, setExpanded] = useState(isAncestor || isActive);

  // Keep expanded in sync with navigation
  if ((isAncestor || isActive) && !expanded) {
    setExpanded(true);
  }

  return (
    <li>
      <div
        className={cn(
          "group flex items-center gap-0.5 rounded-md px-1 py-1 text-sm",
          isActive
            ? "bg-accent text-accent-foreground"
            : "hover:bg-accent/50",
        )}
        style={{ paddingLeft: `${depth * 12 + 4}px` }}
      >
        {/* Expand/collapse arrow */}
        <button
          type="button"
          className={cn(
            "flex size-5 shrink-0 items-center justify-center rounded-sm hover:bg-accent",
            !hasChildren && "invisible",
          )}
          onClick={() => setExpanded(!expanded)}
          aria-label={expanded ? "Collapse" : "Expand"}
        >
          <ChevronRight
            className={cn(
              "size-3.5 transition-transform",
              expanded && "rotate-90",
            )}
          />
        </button>

        {/* Folder name — navigates */}
        <Link
          to="/spaces/$spaceId/folders/$folderId"
          params={{ spaceId, folderId: node.id }}
          className="flex min-w-0 flex-1 items-center gap-1.5 truncate"
        >
          <FolderIcon className="size-4 shrink-0 text-muted-foreground" />
          <span className="truncate">{node.name}</span>
        </Link>
      </div>

      {/* Children */}
      {expanded && hasChildren && (
        <ul className="space-y-0.5">
          {node.children.map((child) => (
            <TreeItem
              key={child.id}
              node={child}
              spaceId={spaceId}
              depth={depth + 1}
            />
          ))}
        </ul>
      )}
    </li>
  );
}

/** Check if a node or any of its descendants has the given ID. */
function containsDescendant(node: FolderNode, targetId: string): boolean {
  for (const child of node.children) {
    if (child.id === targetId || containsDescendant(child, targetId)) {
      return true;
    }
  }
  return false;
}
