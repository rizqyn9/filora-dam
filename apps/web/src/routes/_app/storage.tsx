import { createFileRoute } from "@tanstack/react-router";
import { HardDrive, Layers, Plus } from "lucide-react";

import { PageHeader } from "@/components/page-header";
import { StatCard } from "@/components/stat-card";
import { Button } from "@/components/ui/button";
import { ProviderCard } from "@/features/storage/components/provider-card";
import { storageAccounts } from "@/features/storage/fixtures";
import { formatBytes } from "@/lib/format";

export const Route = createFileRoute("/_app/storage")({
  component: StoragePage,
});

function StoragePage() {
  const totalUsed = storageAccounts.reduce((acc, a) => acc + a.used, 0);
  const activeCount = storageAccounts.filter((a) => a.isActive).length;
  const totalStored = storageAccounts.reduce(
    (acc, a) => acc + a.storedCount,
    0,
  );

  return (
    <>
      <PageHeader
        title="Storage"
        description="Manage storage providers and monitor usage across accounts."
        actions={
          <Button size="sm">
            <Plus className="size-4" />
            Add provider
          </Button>
        }
      />

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard
          label="Total used"
          value={formatBytes(totalUsed)}
          icon={HardDrive}
        />
        <StatCard
          label="Active accounts"
          value={`${activeCount} / ${storageAccounts.length}`}
          icon={Layers}
        />
        <StatCard
          label="Objects stored"
          value={totalStored.toLocaleString()}
          icon={Layers}
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {storageAccounts.map((account) => (
          <ProviderCard key={account.id} account={account} />
        ))}
      </div>
    </>
  );
}
