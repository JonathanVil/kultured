<script>
  import { push } from 'svelte-spa-router'
  import { Button } from '$lib/components/ui/button'
  import { Badge } from '$lib/components/ui/badge'
  import * as Table from '$lib/components/ui/table'

  let batches = $state([])
  let loading = $state(true)
  let error = $state(null)

  $effect(() => {
    fetch('/api/batches')
      .then(r => {
        if (!r.ok) throw new Error(`${r.status} ${r.statusText}`)
        return r.json()
      })
      .then(data => { batches = data ?? []; loading = false })
      .catch(e => { error = e.message; loading = false })
  })

  const stageVariant = (stage) => ({
    f1: 'secondary',
    f2: 'secondary',
    bottled: 'outline',
    done: 'default',
  }[stage] ?? 'secondary')

  const stageLabel = (stage) => ({
    f1: 'F1',
    f2: 'F2',
    bottled: 'Bottled',
    done: 'Done',
  }[stage] ?? stage)
</script>

<div class="flex items-center justify-between mb-6">
  <h1 class="text-2xl font-semibold">Batches</h1>
  <Button href="#/batches/new">New batch</Button>
</div>

{#if loading}
  <p class="text-muted-foreground">Loading…</p>
{:else if error}
  <p class="text-destructive">Failed to load batches: {error}</p>
{:else if batches.length === 0}
  <p class="text-muted-foreground">No batches yet. Time to brew something.</p>
{:else}
  <Table.Root>
    <Table.Header>
      <Table.Row>
        <Table.Head>Name</Table.Head>
        <Table.Head>Tea type</Table.Head>
        <Table.Head>Volume</Table.Head>
        <Table.Head>Started</Table.Head>
        <Table.Head class="text-right">F1 days</Table.Head>
        <Table.Head class="text-right">F2 days</Table.Head>
        <Table.Head>Stage</Table.Head>
      </Table.Row>
    </Table.Header>
    <Table.Body>
      {#each batches as batch (batch.id)}
        <Table.Row
          class="cursor-pointer"
          onclick={() => push(`/batches/${batch.id}`)}
        >
          <Table.Cell class="font-medium">{batch.name}</Table.Cell>
          <Table.Cell class="capitalize">{batch.tea_type}</Table.Cell>
          <Table.Cell>{batch.total_volume_ml} ml</Table.Cell>
          <Table.Cell>{batch.started_at}</Table.Cell>
          <Table.Cell class="text-right">{batch.f1_days}</Table.Cell>
          <Table.Cell class="text-right">{batch.f2_days > 0 ? batch.f2_days : '—'}</Table.Cell>
          <Table.Cell>
            <Badge variant={stageVariant(batch.stage)}>{stageLabel(batch.stage)}</Badge>
          </Table.Cell>
        </Table.Row>
      {/each}
    </Table.Body>
  </Table.Root>
{/if}
