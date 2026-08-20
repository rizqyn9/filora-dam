import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useDeleteSpace } from "@/features/spaces/api";
import type { Space } from "@/features/spaces/schemas";
import { ApiError } from "@/lib/api-client";

interface SpaceDeleteDialogProps {
  space: Space;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * Confirmation dialog for deleting a space.
 */
export function SpaceDeleteDialog({
  space,
  open,
  onOpenChange,
}: SpaceDeleteDialogProps) {
  const deleteSpace = useDeleteSpace();

  const confirm = () => {
    deleteSpace.mutate(space.id, {
      onSuccess: () => {
        toast.success(`Space "${space.name}" deleted`);
        onOpenChange(false);
      },
      onError: (error) => {
        toast.error(
          error instanceof ApiError ? error.message : "Something went wrong",
        );
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete space</DialogTitle>
          <DialogDescription>
            This permanently deletes <strong>{space.name}</strong> and all of
            its files and folders. This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={deleteSpace.isPending}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={confirm}
            disabled={deleteSpace.isPending}
          >
            {deleteSpace.isPending ? "Deleting..." : "Delete"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
