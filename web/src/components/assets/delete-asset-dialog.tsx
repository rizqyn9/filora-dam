import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useDeleteAssetRef } from "@/features/assets/api";

interface Props {
  spaceId: string;
  refId: number;
  assetName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DeleteAssetDialog({ spaceId, refId, assetName, open, onOpenChange }: Props) {
  const deleteRef = useDeleteAssetRef(spaceId);

  function handleDelete() {
    deleteRef.mutate(refId, {
      onSuccess: () => {
        toast.success("Moved to trash");
        onOpenChange(false);
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
            "{assetName}" will be moved to trash.
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteRef.isPending}
          >
            {deleteRef.isPending ? "Deleting..." : "Move to trash"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
