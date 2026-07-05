import { type ReactNode, useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface TagFormDialogProps {
  /** Optional trigger. Omit when driving `open` externally (e.g. row menu). */
  trigger?: ReactNode;
  /** Present = edit mode; absent = create mode. */
  initialName?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

/**
 * Create / edit a tag. Works both trigger-based and fully controlled.
 * Presentational — wire submit to the tag mutations.
 */
export function TagFormDialog({
  trigger,
  initialName,
  open,
  onOpenChange,
}: TagFormDialogProps) {
  const isEdit = initialName !== undefined;
  const [name, setName] = useState(initialName ?? "");

  useEffect(() => {
    if (open) setName(initialName ?? "");
  }, [open, initialName]);

  const submit = () => {
    if (!name.trim()) return;
    toast.success(isEdit ? "Tag updated" : `Tag "${name}" created`);
    onOpenChange?.(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {trigger && <DialogTrigger asChild>{trigger}</DialogTrigger>}
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit tag" : "Create tag"}</DialogTitle>
          <DialogDescription>
            Tags help you filter and group assets within a gallery.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-2 py-2">
          <Label htmlFor="tag-name">Name</Label>
          <Input
            id="tag-name"
            value={name}
            maxLength={64}
            placeholder="e.g. hero, social, approved"
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && submit()}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange?.(false)}>
            Cancel
          </Button>
          <Button onClick={submit}>{isEdit ? "Save" : "Create"}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
