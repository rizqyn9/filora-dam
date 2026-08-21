import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useDeleteFolder } from "@/features/folders/api";

interface Props {
  spaceId: string;
  folderId: string;
  folderName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDeleted?: () => void;
}

export function DeleteFolderDialog({ spaceId, folderId, folderName, open, onOpenChange, onDeleted }: Props) {
  const deleteFolder = useDeleteFolder(spaceId);

  function handleDelete() {
    deleteFolder.mutate(folderId, {
      onSuccess: () => {
        toast.success("Folder moved to trash");
        onOpenChange(false);
        onDeleted?.();
      },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Move to trash?</DialogTitle>
          <DialogDescription>
            "{folderName}" and all files inside will be moved to trash.
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteFolder.isPending}
          >
            {deleteFolder.isPending ? "Deleting..." : "Move to trash"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
