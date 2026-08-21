import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useRef } from "react";
import { toast } from "sonner";

import { useAuthStore } from "@/features/auth/auth-store";
import { ApiRequestError } from "@/lib/api";

import { AssetSchema } from "./schemas";

const CONCURRENCY = 2;

interface UploadContext {
  spaceId: string;
  folderId?: string;
}

/**
 * Hook that uploads multiple files with a concurrency cap
 * and shows a persistent aggregate progress toast.
 */
export function useUploadFiles({ spaceId, folderId }: UploadContext) {
  const queryClient = useQueryClient();
  const toastId = useRef<string | number | undefined>(undefined);

  const upload = useCallback(
    async (files: File[]) => {
      const total = files.length;
      let completed = 0;
      let failed = 0;

      toastId.current = toast.loading(`Uploading 0/${total}...`, {
        duration: Infinity,
      });

      // Process files with concurrency limit
      const queue = [...files];
      const workers = Array.from({ length: CONCURRENCY }, async () => {
        while (queue.length) {
          const file = queue.shift()!;
          try {
            await uploadSingleFile(spaceId, folderId, file);
          } catch {
            failed++;
          }
          completed++;
          toast.loading(`Uploading ${completed}/${total}...`, {
            id: toastId.current,
          });
        }
      });

      await Promise.all(workers);

      // Final toast
      if (failed === 0) {
        toast.success(`${total} file${total > 1 ? "s" : ""} uploaded`, {
          id: toastId.current,
          duration: 5000,
        });
      } else {
        toast.warning(
          `${completed - failed} uploaded, ${failed} failed`,
          { id: toastId.current, duration: undefined },
        );
      }

      // Invalidate asset list
      queryClient.invalidateQueries({ queryKey: ["assets", spaceId] });
    },
    [spaceId, folderId, queryClient],
  );

  return upload;
}

async function uploadSingleFile(
  spaceId: string,
  folderId: string | undefined,
  file: File,
) {
  const token = useAuthStore.getState().token;
  const form = new FormData();
  form.append("space_id", spaceId);
  if (folderId) form.append("folder_id", folderId);
  form.append("file", file);

  const res = await fetch("/api/v1/assets/upload", {
    method: "POST",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  });

  const body = await res.json();
  if (!res.ok || !body.success) {
    throw new ApiRequestError(res.status, body.error);
  }
  return AssetSchema.parse(body.data);
}
