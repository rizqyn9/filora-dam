import { type FormEvent, useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useRenameFolder } from "@/features/folders/api";

interface Props {
  spaceId: string;
  folderId: string;
  currentName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RenameFolderDialog({ spaceId, folderId, currentName, open, onOpenChange }: Props) {
  const [name, setName] = useState(currentName);
  const renameFolder = useRenameFolder(spaceId);

  useEffect(() => { setName(currentName); }, [currentName]);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    renameFolder.mutate(
      { folderId, name: name.trim() },
      {
        onSuccess: () => {
          toast.success("Folder renamed");
          onOpenChange(false);
        },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Rename Folder</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="rename-folder">Name</Label>
            <Input
              id="rename-folder"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              maxLength={255}
              autoFocus
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={!name.trim() || renameFolder.isPending}>
              {renameFolder.isPending ? "Renaming..." : "Rename"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
