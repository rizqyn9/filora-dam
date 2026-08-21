import { type FormEvent, useState } from "react";
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
import { useCreateFolder } from "@/features/folders/api";

interface Props {
  spaceId: string;
  parentId?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateFolderDialog({ spaceId, parentId, open, onOpenChange }: Props) {
  const [name, setName] = useState("");
  const createFolder = useCreateFolder(spaceId);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    createFolder.mutate(
      { name: name.trim(), parentId },
      {
        onSuccess: () => {
          toast.success("Folder created");
          setName("");
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
          <DialogTitle>New Folder</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="folder-name">Name</Label>
            <Input
              id="folder-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Folder name"
              required
              maxLength={255}
              autoFocus
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={!name.trim() || createFolder.isPending}>
              {createFolder.isPending ? "Creating..." : "Create"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
