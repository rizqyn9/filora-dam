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
import { useRenameAsset } from "@/features/assets/api";

interface Props {
  spaceId: string;
  assetId: string;
  currentName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RenameAssetDialog({ spaceId, assetId, currentName, open, onOpenChange }: Props) {
  const [name, setName] = useState(currentName);
  const rename = useRenameAsset(spaceId);

  useEffect(() => { setName(currentName); }, [currentName]);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    rename.mutate(
      { assetId, name: name.trim() },
      {
        onSuccess: () => {
          toast.success("Asset renamed");
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
          <DialogTitle>Rename</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="rename-asset">Name</Label>
            <Input
              id="rename-asset"
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
            <Button type="submit" disabled={!name.trim() || rename.isPending}>
              {rename.isPending ? "Renaming..." : "Rename"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
