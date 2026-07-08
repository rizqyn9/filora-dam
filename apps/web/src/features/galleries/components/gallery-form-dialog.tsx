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
import { useCreateGallery, useUpdateGallery } from "@/features/galleries/api";
import type { Gallery } from "@/features/galleries/schemas";
import { ApiError } from "@/lib/api-client";

interface GalleryFormDialogProps {
  /** Optional trigger. Omit when driving `open` externally (e.g. row menu). */
  trigger?: ReactNode;
  /** Present = edit mode; absent = create mode. */
  gallery?: Gallery;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

/**
 * Create / edit a gallery. Works both trigger-based and fully controlled.
 * Wired to the gallery create/update mutations.
 */
export function GalleryFormDialog({
  trigger,
  gallery,
  open,
  onOpenChange,
}: GalleryFormDialogProps) {
  const isEdit = gallery !== undefined;
  const [name, setName] = useState(gallery?.name ?? "");
  const [description, setDescription] = useState(gallery?.description ?? "");

  const createGallery = useCreateGallery();
  const updateGallery = useUpdateGallery(gallery?.id ?? 0);
  const mutation = isEdit ? updateGallery : createGallery;

  useEffect(() => {
    if (open) {
      setName(gallery?.name ?? "");
      setDescription(gallery?.description ?? "");
    }
  }, [open, gallery]);

  const submit = () => {
    const trimmed = name.trim();
    if (!trimmed) return;

    const input = {
      name: trimmed,
      description: description.trim() || null,
    };

    mutation.mutate(input, {
      onSuccess: () => {
        toast.success(isEdit ? "Gallery updated" : `Gallery "${trimmed}" created`);
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
          <DialogTitle>{isEdit ? "Edit gallery" : "Create gallery"}</DialogTitle>
          <DialogDescription>
            Galleries group your assets, members, and storage together.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="gallery-name">Name</Label>
            <Input
              id="gallery-name"
              value={name}
              maxLength={255}
              placeholder="e.g. Marketing 2026"
              onChange={(e) => setName(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && submit()}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="gallery-description">Description</Label>
            <Textarea
              id="gallery-description"
              value={description}
              maxLength={1000}
              placeholder="What's this gallery for? (optional)"
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
