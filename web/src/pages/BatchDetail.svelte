<script>
  import { push } from 'svelte-spa-router'
  import { Button } from '$lib/components/ui/button'
  import { Badge } from '$lib/components/ui/badge'
  import { Textarea } from '$lib/components/ui/textarea'
  import * as Card from '$lib/components/ui/card'
  import { Separator } from '$lib/components/ui/separator'

  let { params = {} } = $props()
  const id = params.id

  let batch = $state(null)
  let loading = $state(true)
  let error = $state(null)
  let noteText = $state('')
  let submittingNote = $state(false)
  let advancingStage = $state(false)
  let deleting = $state(false)

  $effect(() => {
    fetch(`/api/batches/${id}`)
      .then(r => { if (!r.ok) throw new Error(`${r.status}`); return r.json() })
      .then(data => { batch = data; loading = false })
      .catch(e => { error = e.message; loading = false })
  })

  const stageMap = { f1: 'f2', f2: 'bottled', bottled: 'done' }
  const stageLabelMap = { f1: 'F1', f2: 'F2', bottled: 'Bottled', done: 'Done' }
  const stageVariantMap = { f1: 'secondary', f2: 'secondary', bottled: 'outline', done: 'default' }
  const advanceLabelMap = { f2: 'Move to F2', bottled: 'Move to bottled', done: 'Mark as done' }

  const nextStageKey = $derived(batch ? (stageMap[batch.stage] ?? null) : null)
  const nextStageLabel = $derived(nextStageKey ? advanceLabelMap[nextStageKey] : null)

  async function advanceStage() {
    advancingStage = true
    error = null
    try {
      const res = await fetch(`/api/batches/${id}/stage`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ stage: nextStageKey }),
      })
      if (!res.ok) throw new Error(`${res.status}`)
      const updated = await res.json()
      batch = { ...updated, notes: batch.notes }
    } catch (e) {
      error = e.message
    } finally {
      advancingStage = false
    }
  }

  async function deleteBatch() {
    if (!confirm(`Delete "${batch.name}"? This cannot be undone.`)) return
    deleting = true
    try {
      const res = await fetch(`/api/batches/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error(`${res.status}`)
      push('/')
    } catch (e) {
      error = e.message
      deleting = false
    }
  }

  async function addNote(e) {
    e.preventDefault()
    if (!noteText.trim()) return
    submittingNote = true
    error = null
    try {
      const res = await fetch(`/api/batches/${id}/notes`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ note: noteText.trim() }),
      })
      if (!res.ok) throw new Error(`${res.status}`)
      const note = await res.json()
      batch.notes = [note, ...batch.notes]
      noteText = ''
    } catch (e) {
      error = e.message
    } finally {
      submittingNote = false
    }
  }

  async function deleteNote(noteId) {
    try {
      const res = await fetch(`/api/notes/${noteId}`, { method: 'DELETE' })
      if (!res.ok) throw new Error(`${res.status}`)
      batch.notes = batch.notes.filter(n => n.id !== noteId)
    } catch (e) {
      error = e.message
    }
  }

  function fmtDate(ts) {
    if (!ts) return null
    // Handle both "YYYY-MM-DD" and ISO timestamps
    return new Date(ts.includes('T') ? ts : ts + 'T00:00:00').toLocaleDateString()
  }

  function fmtTs(ts) {
    if (!ts) return null
    return new Date(ts).toLocaleString()
  }
</script>

{#if loading}
  <p class="text-muted-foreground">Loading…</p>
{:else if error && !batch}
  <p class="text-destructive">{error}</p>
{:else if batch}
  <!-- Header -->
  <div class="flex items-start justify-between mb-6">
    <div class="flex items-center gap-3">
      <Button variant="ghost" size="sm" href="#/">← Back</Button>
      <div>
        <h1 class="text-2xl font-semibold leading-tight">{batch.name}</h1>
        <Badge class="mt-1" variant={stageVariantMap[batch.stage]}>
          {stageLabelMap[batch.stage]}
        </Badge>
      </div>
    </div>
    <Button variant="destructive" size="sm" onclick={deleteBatch} disabled={deleting}>
      {deleting ? 'Deleting…' : 'Delete batch'}
    </Button>
  </div>

  {#if error}
    <p class="text-destructive text-sm mb-4">{error}</p>
  {/if}

  <!-- Batch details -->
  <Card.Root class="mb-4">
    <Card.Content class="pt-6 grid grid-cols-2 gap-x-8 gap-y-2 text-sm">
      <div class="flex justify-between">
        <span class="text-muted-foreground">Tea type</span>
        <span class="font-medium capitalize">{batch.tea_type}</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">Tea</span>
        <span class="font-medium">{batch.tea_g} g</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">Steep</span>
        <span class="font-medium">{batch.steep_min} min</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">Sugar</span>
        <span class="font-medium">{batch.sugar_g} g</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">Tea volume</span>
        <span class="font-medium">{batch.tea_volume_l} L</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">SCOBY volume</span>
        <span class="font-medium">{batch.scoby_volume_ml} mL</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">Total volume</span>
        <span class="font-medium">{batch.total_volume_l.toFixed(2)} L</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">Started</span>
        <span class="font-medium">{fmtDate(batch.started_at)}</span>
      </div>
      {#if batch.start_f2}
        <div class="flex justify-between">
          <span class="text-muted-foreground">F2 started</span>
          <span class="font-medium">{fmtTs(batch.start_f2)}</span>
        </div>
      {/if}
      {#if batch.done_at}
        <div class="flex justify-between">
          <span class="text-muted-foreground">Done</span>
          <span class="font-medium">{fmtTs(batch.done_at)}</span>
        </div>
      {/if}
    </Card.Content>
  </Card.Root>

  <!-- Stats -->
  <div class="grid grid-cols-5 gap-2 mb-4 text-center">
    {#each [
      { label: 'F1 days', value: batch.f1_days },
      { label: 'F2 days', value: batch.f2_days > 0 ? batch.f2_days : '—' },
      { label: 'Backslop', value: batch.backslop_pct.toFixed(1) + '%' },
      { label: 'Sugar Brix', value: batch.sugar_pct.toFixed(1) + '%' },
      { label: 'Tea g/L', value: batch.tea_g_per_l.toFixed(1) },
    ] as stat}
      <Card.Root>
        <Card.Content class="pt-4 pb-4">
          <p class="text-xs text-muted-foreground mb-1">{stat.label}</p>
          <p class="text-lg font-semibold">{stat.value}</p>
        </Card.Content>
      </Card.Root>
    {/each}
  </div>

  <!-- Stage advancement -->
  {#if nextStageLabel}
    <Card.Root class="mb-4">
      <Card.Content class="pt-4 pb-4 flex items-center justify-between">
        <p class="text-sm text-muted-foreground">Ready for the next stage?</p>
        <Button onclick={advanceStage} disabled={advancingStage}>
          {advancingStage ? 'Updating…' : nextStageLabel}
        </Button>
      </Card.Content>
    </Card.Root>
  {/if}

  <Separator class="my-6" />

  <!-- Notes -->
  <div>
    <h2 class="text-lg font-semibold mb-3">Notes</h2>

    <form onsubmit={addNote} class="flex gap-2 mb-4">
      <Textarea
        bind:value={noteText}
        placeholder="Add a note…"
        rows={2}
        class="resize-none"
        required
      />
      <Button type="submit" disabled={submittingNote} class="self-end">
        {submittingNote ? '…' : 'Add'}
      </Button>
    </form>

    {#if batch.notes.length === 0}
      <p class="text-sm text-muted-foreground">No notes yet.</p>
    {:else}
      <div class="space-y-2">
        {#each batch.notes as note (note.id)}
          <div class="flex items-start justify-between gap-3 rounded-md border p-3 text-sm">
            <div class="min-w-0">
              <p class="whitespace-pre-wrap">{note.note}</p>
              <p class="text-xs text-muted-foreground mt-1">{fmtTs(note.created_at)}</p>
            </div>
            <Button
              variant="ghost"
              size="sm"
              class="shrink-0 text-muted-foreground hover:text-destructive"
              onclick={() => deleteNote(note.id)}
            >✕</Button>
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}
