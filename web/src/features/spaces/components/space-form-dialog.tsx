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
import { Textarea } from "@/components/ui/textarea";
import { useCreateSpace, useUpdateSpace } from "@/features/spaces/api";
import type { Space } from "@/features/spaces/schemas";
import { ApiError } from "@/lib/api-client";

interface SpaceFormDialogProps {
  /** Optional trigger. Omit when driving `open` externally (e.g. row menu). */
  trigger?: ReactNode;
  /** Present = edit mode; absent = create mode. */
  space?: Space;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

/**
 * Create / edit a space. Works both trigger-based and fully controlled.
 */
export function SpaceFormDialog({
  trigger,
  space,
  open,
  onOpenChange,
}: SpaceFormDialogProps) {
  const isEdit = space !== undefined;
  const [name, setName] = useState(space?.name ?? "");
  const [description, setDescription] = useState(space?.description ?? "");

  const createSpace = useCreateSpace();
  const updateSpace = useUpdateSpace(space?.id ?? 0);
  const mutation = isEdit ? updateSpace : createSpace;

  useEffect(() => {
    if (open) {
      setName(space?.name ?? "");
      setDescription(space?.description ?? "");
    }
  }, [open, space]);

  const submit = () => {
    const trimmed = name.trim();
    if (!trimmed) return;

    const input = {
      name: trimmed,
      description: description.trim() || null,
    };

    mutation.mutate(input, {
      onSuccess: () => {
        toast.success(isEdit ? "Space updated" : `Space "${trimmed}" created`);
        onOpenChange?.(false);
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
      {trigger && <DialogTrigger asChild>{trigger}</DialogTrigger>}
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit space" : "Create space"}</DialogTitle>
          <DialogDescription>
            Spaces group your files, members, and storage together.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="space-name">Name</Label>
            <Input
              id="space-name"
              value={name}
              maxLength={255}
              placeholder="e.g. Family Photos"
              onChange={(e) => setName(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && submit()}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="space-description">Description</Label>
            <Textarea
              id="space-description"
              value={description}
              maxLength={1000}
              placeholder="What's this space for? (optional)"
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange?.(false)}
            disabled={mutation.isPending}
          >
            Cancel
          </Button>
          <Button onClick={submit} disabled={!name.trim() || mutation.isPending}>
            {mutation.isPending
              ? "Saving..."
              : isEdit
                ? "Save"
                : "Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
