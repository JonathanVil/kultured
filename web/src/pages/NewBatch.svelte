<script>
  import { push } from 'svelte-spa-router'
  import { Button } from '$lib/components/ui/button'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'
  import * as Card from '$lib/components/ui/card'

  let name = $state('')
  let teaType = $state('Black')
  let teaG = $state('')
  let steepMin = $state('')
  let startedAt = $state(new Date().toISOString().slice(0, 10))
  let teaVolumeML = $state('')
  let scobyVolumeML = $state('')
  let sugarG = $state('')

  let submitting = $state(false)
  let error = $state(null)

  const totalVolumeML = $derived(
    (parseFloat(teaVolumeML) || 0) + (parseFloat(scobyVolumeML) || 0)
  )
  const backslopPct = $derived(
    totalVolumeML > 0 ? (parseFloat(scobyVolumeML) || 0) / totalVolumeML * 100 : null
  )
  const sugarPct = $derived(
    totalVolumeML > 0 ? (parseFloat(sugarG) || 0) / totalVolumeML * 100 : null
  )
  const teaGPerL = $derived(
    (parseFloat(teaVolumeML) || 0) > 0 ? (parseFloat(teaG) || 0) / (parseFloat(teaVolumeML) || 0) * 1000 : null
  )
  const showPreview = $derived(totalVolumeML > 0 || teaGPerL !== null)

  async function submit(e) {
    e.preventDefault()
    submitting = true
    error = null
    try {
      const res = await fetch('/api/batches', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name,
          tea_type: teaType,
          tea_g: parseFloat(teaG),
          steep_min: parseFloat(steepMin),
          started_at: startedAt,
          tea_volume_ml: parseFloat(teaVolumeML),
          scoby_volume_ml: parseFloat(scobyVolumeML),
          sugar_g: parseFloat(sugarG),
        }),
      })
      if (!res.ok) throw new Error((await res.text()) || `${res.status}`)
      const batch = await res.json()
      push(`/batches/${batch.id}`)
    } catch (err) {
      error = err.message
      submitting = false
    }
  }
</script>

<div class="max-w-lg">
  <div class="flex items-center gap-4 mb-6">
    <Button variant="ghost" size="sm" href="#/">← Back</Button>
    <h1 class="text-2xl font-semibold">New batch</h1>
  </div>

  {#if error}
    <p class="text-destructive mb-4 text-sm">{error}</p>
  {/if}

  <form onsubmit={submit} class="space-y-4">
    <div class="space-y-1">
      <Label for="name">Name</Label>
      <Input id="name" bind:value={name} required placeholder="e.g. Summer brew #3" />
    </div>

    <div class="space-y-1">
        <Label for="tea_type">Tea type</Label>
        <Input
            bind:value={teaType}
            placeholder="e.g. Rooibos blend"
            required
            class="mt-1"
        />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-1">
        <Label for="tea_g">Tea (g)</Label>
        <Input id="tea_g" type="number" min="0" step="0.5" bind:value={teaG} required />
      </div>
      <div class="space-y-1">
        <Label for="steep_min">Steep (min)</Label>
        <Input id="steep_min" type="number" min="0" step="0.5" bind:value={steepMin} required />
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-1">
        <Label for="tea_volume_ml">Tea volume (ml)</Label>
        <Input id="tea_volume_ml" type="number" min="0" step="0.1" bind:value={teaVolumeML} required />
      </div>
      <div class="space-y-1">
        <Label for="scoby_volume_ml">SCOBY volume (ml)</Label>
        <Input id="scoby_volume_ml" type="number" min="0" step="1" bind:value={scobyVolumeML} required />
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="space-y-1">
        <Label for="sugar_g">Sugar (g)</Label>
        <Input id="sugar_g" type="number" min="0" step="1" bind:value={sugarG} required />
      </div>
      <div class="space-y-1">
        <Label for="started_at">Start date</Label>
        <Input id="started_at" type="date" bind:value={startedAt} required />
      </div>
    </div>

    {#if showPreview}
      <Card.Root class="bg-muted/50">
        <Card.Content class="pt-4 pb-4 grid grid-cols-3 gap-4 text-sm text-center">
          <div>
            <p class="text-muted-foreground text-xs mb-1">Total volume</p>
            <p class="font-medium">{totalVolumeML} ml</p>
          </div>
          <div>
            <p class="text-muted-foreground text-xs mb-1">Backslop</p>
            <p class="font-medium">{backslopPct !== null ? backslopPct.toFixed(1) + '%' : '—'}</p>
          </div>
          <div>
            <p class="text-muted-foreground text-xs mb-1">Sugar (Brix)</p>
            <p class="font-medium">{sugarPct !== null ? sugarPct.toFixed(1) + '%' : '—'}</p>
          </div>
          <div class="col-span-3">
            <p class="text-muted-foreground text-xs mb-1">Tea concentration</p>
            <p class="font-medium">{teaGPerL !== null ? teaGPerL.toFixed(1) + ' g/L' : '—'}</p>
          </div>
        </Card.Content>
      </Card.Root>
    {/if}

    <div class="flex gap-2 pt-2">
      <Button type="submit" disabled={submitting}>
        {submitting ? 'Creating…' : 'Start brewing'}
      </Button>
      <Button type="button" variant="outline" href="#/">Cancel</Button>
    </div>
  </form>
</div>
