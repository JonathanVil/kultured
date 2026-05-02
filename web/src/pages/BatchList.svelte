<script>
  import { push } from 'svelte-spa-router'
  import { Button } from '$lib/components/ui/button'
  import * as Table from '$lib/components/ui/table'
  import { fmtDate } from '$lib/utils.js'

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

  const stageBadgeClass = (stage) => ({
    f1: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
    f2: 'bg-lime-100 text-lime-800 dark:bg-lime-900/40 dark:text-lime-300',
    bottled: 'bg-stone-100 text-stone-700 dark:bg-stone-800 dark:text-stone-300',
    done: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300',
  }[stage] ?? 'bg-secondary text-secondary-foreground')

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
          <Table.Cell>{fmtDate(batch.started_at)}</Table.Cell>
          <Table.Cell class="text-right">{batch.f1_days}</Table.Cell>
          <Table.Cell class="text-right">{batch.f2_days > 0 ? batch.f2_days : '—'}</Table.Cell>
          <Table.Cell>
            <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {stageBadgeClass(batch.stage)}">
              {stageLabel(batch.stage)}
            </span>
          </Table.Cell>
        </Table.Row>
      {/each}
    </Table.Body>
  </Table.Root>
{/if}
