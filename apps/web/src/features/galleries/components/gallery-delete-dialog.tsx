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
import { useDeleteGallery } from "@/features/galleries/api";
import type { Gallery } from "@/features/galleries/schemas";
import { ApiError } from "@/lib/api-client";

interface GalleryDeleteDialogProps {
  gallery: Gallery;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * Confirmation dialog for deleting a gallery. Wired to the delete mutation.
 */
export function GalleryDeleteDialog({
  gallery,
  open,
  onOpenChange,
}: GalleryDeleteDialogProps) {
  const deleteGallery = useDeleteGallery();

  const confirm = () => {
    deleteGallery.mutate(gallery.id, {
      onSuccess: () => {
        toast.success(`Gallery "${gallery.name}" deleted`);
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
          <DialogTitle>Delete gallery</DialogTitle>
          <DialogDescription>
            This permanently deletes <strong>{gallery.name}</strong> and all of
            its assets. This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={deleteGallery.isPending}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={confirm}
            disabled={deleteGallery.isPending}
          >
            {deleteGallery.isPending ? "Deleting..." : "Delete"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
