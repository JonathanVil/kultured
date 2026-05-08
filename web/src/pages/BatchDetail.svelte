<script>
    import {push} from 'svelte-spa-router'
    import {Badge} from '$lib/components/ui/badge'
    import {Textarea} from '$lib/components/ui/textarea'
    import * as Card from '$lib/components/ui/card'
    import {Separator} from '$lib/components/ui/separator'
    import {Input} from '$lib/components/ui/input';
    import {Button} from "$lib/components/ui/button/index.js";
    import {fmtDate, fmtTs} from '$lib/utils.js';

    let {params = {}} = $props()
    const id = params.id

    let batch = $state(null)
    let loading = $state(true)
    let error = $state(null)
    let noteText = $state('')
    let submittingNote = $state(false)
    let advancingStage = $state(false)
    let deleting = $state(false)
    let editing = $state(false)
    let editingBatch = $state(null)
    let ntfyEnabled = $state(false)

    // Reminder UI state
    let reminderMode = $state('idle') // 'idle' | 'editing'
    let reminderDraft = $state({reminder_interval_days: 1, reminder_time: '08:00', reminder_day_of_week: null})
    let savingReminder = $state(false)

    const DOW_LABELS = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']

    const reminderFreqKey = $derived.by(() => {
        const d = reminderDraft.reminder_interval_days
        if (d === 1) return 'daily'
        if (d === 7) return 'weekly'
        return 'custom'
    })

    function reminderSummary(b) {
        const d = b.reminder_interval_days
        let freq
        if (d === 1) {
            freq = 'Daily'
        } else if (d === 7 && b.reminder_day_of_week != null) {
            freq = `Weekly on ${DOW_LABELS[b.reminder_day_of_week]}`
        } else {
            freq = `Every ${d} days`
        }
        return `${freq} at ${b.reminder_time}`
    }

    function openReminderForm() {
        reminderDraft = {
            reminder_interval_days: batch.reminder_interval_days ?? 1,
            reminder_time: batch.reminder_time ?? '08:00',
            reminder_day_of_week: batch.reminder_day_of_week ?? (new Date().getDay() + 6) % 7,
        }
        reminderMode = 'editing'
    }

    function setReminderFreq(freq) {
        if (freq === 'daily') {
            reminderDraft.reminder_interval_days = 1
            reminderDraft.reminder_day_of_week = null
        } else if (freq === 'weekly') {
            reminderDraft.reminder_interval_days = 7
            if (reminderDraft.reminder_day_of_week == null) {
                reminderDraft.reminder_day_of_week = (new Date().getDay() + 6) % 7
            }
        } else if (freq === 'custom') {
            if (reminderDraft.reminder_interval_days === 1 || reminderDraft.reminder_interval_days === 7) {
                reminderDraft.reminder_interval_days = 2
            }
            reminderDraft.reminder_day_of_week = null
        }
    }

    async function saveReminder() {
        savingReminder = true
        error = null
        try {
            const res = await fetch(`/api/batches/${id}`, {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({batch: {
                    ...batch,
                    reminder_enabled: true,
                    reminder_interval_days: reminderDraft.reminder_interval_days,
                    reminder_time: reminderDraft.reminder_time,
                    reminder_day_of_week: reminderDraft.reminder_day_of_week,
                }}),
            })
            if (!res.ok) throw new Error(`${res.status}`)
            const updated = await res.json()
            batch = {...updated, notes: batch.notes}
            reminderMode = 'idle'
        } catch (e) {
            error = e.message
        } finally {
            savingReminder = false
        }
    }

    async function removeReminder() {
        error = null
        try {
            const res = await fetch(`/api/batches/${id}`, {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({batch: {...batch, reminder_enabled: false}}),
            })
            if (!res.ok) throw new Error(`${res.status}`)
            const updated = await res.json()
            batch = {...updated, notes: batch.notes}
            reminderMode = 'idle'
        } catch (e) {
            error = e.message
        }
    }

    $effect(() => {
        Promise.all([
            fetch(`/api/batches/${id}`).then(r => {
                if (!r.ok) throw new Error(`${r.status}`)
                return r.json()
            }),
            fetch('/api/config').then(r => r.ok ? r.json() : {ntfy_enabled: false}),
        ]).then(([data, cfg]) => {
            batch = data
            ntfyEnabled = cfg.ntfy_enabled
            loading = false
        }).catch(e => {
            error = e.message
            loading = false
        })
    })

    const stageMap = {f1: 'f2', f2: 'bottled', bottled: 'done'}
    const stageLabelMap = {f1: 'F1', f2: 'F2', bottled: 'Bottled', done: 'Done'}
    const stageBadgeClass = {
        f1: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
        f2: 'bg-lime-100 text-lime-800 dark:bg-lime-900/40 dark:text-lime-300',
        bottled: 'bg-stone-100 text-stone-700 dark:bg-stone-800 dark:text-stone-300',
        done: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300',
    }
    const advanceLabelMap = {f2: 'Move to F2', bottled: 'Move to bottled', done: 'Mark as done'}

    const nextStageKey = $derived(batch ? (stageMap[batch.stage] ?? null) : null)
    const nextStageLabel = $derived(nextStageKey ? advanceLabelMap[nextStageKey] : null)

    async function advanceStage() {
        advancingStage = true
        error = null
        try {
            const res = await fetch(`/api/batches/${id}/stage`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({stage: nextStageKey}),
            })
            if (!res.ok) throw new Error(`${res.status}`)
            const updated = await res.json()
            batch = {...updated, notes: batch.notes}
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
            const res = await fetch(`/api/batches/${id}`, {method: 'DELETE'})
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
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({note: noteText.trim()}),
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
            const res = await fetch(`/api/notes/${noteId}`, {method: 'DELETE'})
            if (!res.ok) throw new Error(`${res.status}`)
            batch.notes = batch.notes.filter(n => n.id !== noteId)
        } catch (e) {
            error = e.message
        }
    }

    function startEditing() {
        editingBatch = $state.snapshot(batch)
        editing = true
    }

    function cancelEditing() {
        editing = false
        editingBatch = null
    }

    async function saveEditing() {
        error = null
        try {
            const res = await fetch(`/api/batches/${id}`, {
                method: 'PUT',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({batch: editingBatch}),
            })
            if (!res.ok) throw new Error(`${res.status}`)
            const updated = await res.json()
            batch = {...updated, notes: batch.notes}
            editing = false
            editingBatch = null
        } catch (e) {
            error = e.message
        }
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
                <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium mt-1 {stageBadgeClass[batch.stage]}">
                    {stageLabelMap[batch.stage]}
                </span>
            </div>
        </div>
        <div class="flex gap-2">
            {#if !editing}
                <Button variant="secondary" size="sm" onclick={startEditing}>Edit</Button>
                <Button variant="destructive" size="sm" onclick={deleteBatch} disabled={deleting}>
                    {deleting ? 'Deleting…' : 'Delete'}
                </Button>
            {:else}
                <Button size="sm" onclick={saveEditing}>Save</Button>
                <Button variant="outline" size="sm" onclick={cancelEditing}>Cancel</Button>
            {/if}
        </div>
    </div>

    {#if error}
        <p class="text-destructive text-sm mb-4">{error}</p>
    {/if}

    <!-- Batch details -->
    {#if !editing}
        <Card.Root class="mb-4">
            <Card.Content class="pt-6 grid grid-cols-2 gap-x-8 gap-y-2 text-sm">
                <div class="flex justify-between">
                    <span class="text-muted-foreground">Tea type</span>
                    <span class="font-medium">{batch.tea_type}</span>
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
                    <span class="font-medium">{batch.tea_volume_ml} ml</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-muted-foreground">SCOBY volume</span>
                    <span class="font-medium">{batch.scoby_volume_ml} ml</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-muted-foreground">Total volume</span>
                    <span class="font-medium">{batch.total_volume_ml} ml</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-muted-foreground">Started</span>
                    <span class="font-medium">{fmtDate(batch.started_at)}</span>
                </div>
                {#if batch.start_f2}
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">F2 started</span>
                        <span class="font-medium">{fmtDate(batch.start_f2)}</span>
                    </div>
                {/if}
                {#if batch.done_at}
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">Done</span>
                        <span class="font-medium">{fmtDate(batch.done_at)}</span>
                    </div>
                {/if}
            </Card.Content>
        </Card.Root>
    {:else}
        <Card.Root class="mb-4">
            <Card.Content class="pt-6 grid grid-cols-2 gap-x-8 gap-y-3 text-sm">
                <div class="space-y-1">
                    <span class="text-muted-foreground">Name</span>
                    <Input bind:value={editingBatch.name} />
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">Stage</span>
                    <select
                        bind:value={editingBatch.stage}
                        class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
                    >
                        <option value="f1">F1</option>
                        <option value="f2">F2</option>
                        <option value="bottled">Bottled</option>
                        <option value="done">Done</option>
                    </select>
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">Tea type</span>
                    <Input bind:value={editingBatch.tea_type} />
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">Started</span>
                    <Input type="date" bind:value={editingBatch.started_at} />
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">Tea (g)</span>
                    <Input type="number" min="0" step="0.5" bind:value={editingBatch.tea_g} />
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">Steep (min)</span>
                    <Input type="number" min="0" step="0.5" bind:value={editingBatch.steep_min} />
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">Sugar (g)</span>
                    <Input type="number" min="0" step="1" bind:value={editingBatch.sugar_g} />
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">Tea volume (ml)</span>
                    <Input type="number" min="0" step="1" bind:value={editingBatch.tea_volume_ml} />
                </div>
                <div class="space-y-1">
                    <span class="text-muted-foreground">SCOBY volume (ml)</span>
                    <Input type="number" min="0" step="1" bind:value={editingBatch.scoby_volume_ml} />
                </div>
            </Card.Content>
        </Card.Root>
    {/if}

    <!-- Stats -->
    <div class="grid grid-cols-5 gap-2 mb-4 text-center">
        {#each [
            {label: 'F1 days', value: batch.f1_days},
            {label: 'F2 days', value: batch.f2_days > 0 ? batch.f2_days : '—'},
            {label: 'Backslop', value: batch.backslop_pct.toFixed(1) + '%'},
            {label: 'Sugar Brix', value: batch.sugar_pct.toFixed(1) + '%'},
            {label: 'Tea g/L', value: batch.tea_g_per_l.toFixed(1)},
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
            <Card.Content class="pt-2 pb-2 flex items-center justify-between">
                <p class="text-sm text-muted-foreground">Ready for the next stage?</p>
                <Button onclick={advanceStage} disabled={advancingStage}>
                    {advancingStage ? 'Updating…' : nextStageLabel}
                </Button>
            </Card.Content>
        </Card.Root>
    {/if}

    <!-- Reminder (only shown when ntfy is configured) -->
    {#if ntfyEnabled && batch.stage !== 'done'}
        <Card.Root class="mb-4">
            {#if reminderMode === 'idle'}
                <Card.Content class="pt-2 pb-2 flex items-center justify-between">
                    {#if batch.reminder_enabled}
                        <div>
                            <p class="text-sm font-medium">Reminder</p>
                            <p class="text-xs text-muted-foreground">{reminderSummary(batch)}</p>
                        </div>
                        <Button variant="ghost" size="sm" class="text-muted-foreground hover:text-destructive" onclick={removeReminder}>
                            Remove reminder
                        </Button>
                    {:else}
                        <p class="text-sm text-muted-foreground">No reminder set.</p>
                        <Button variant="outline" size="sm" onclick={openReminderForm}>Add reminder</Button>
                    {/if}
                </Card.Content>
            {:else}
                <Card.Content class="pt-2 pb-2 space-y-3">
                    <p class="text-sm font-medium">Reminder</p>
                    <div class="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
                        <div class="space-y-1">
                            <span class="text-muted-foreground">Frequency</span>
                            <select
                                value={reminderFreqKey}
                                onchange={e => setReminderFreq(e.target.value)}
                                class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
                            >
                                <option value="daily">Daily</option>
                                <option value="weekly">Weekly</option>
                                <option value="custom">Custom</option>
                            </select>
                        </div>
                        <div class="space-y-1">
                            <span class="text-muted-foreground">Time</span>
                            <Input type="time" bind:value={reminderDraft.reminder_time} />
                        </div>
                        {#if reminderFreqKey === 'weekly'}
                            <div class="space-y-1">
                                <span class="text-muted-foreground">Day</span>
                                <select
                                    bind:value={reminderDraft.reminder_day_of_week}
                                    class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
                                >
                                    {#each DOW_LABELS as label, i}
                                        <option value={i}>{label}</option>
                                    {/each}
                                </select>
                            </div>
                        {:else if reminderFreqKey === 'custom'}
                            <div class="space-y-1">
                                <span class="text-muted-foreground">Days between</span>
                                <Input
                                    type="number"
                                    min="1"
                                    step="1"
                                    bind:value={reminderDraft.reminder_interval_days}
                                />
                            </div>
                        {/if}
                    </div>
                    <div class="flex gap-2 pt-1">
                        <Button size="sm" onclick={saveReminder} disabled={savingReminder}>
                            {savingReminder ? 'Saving…' : 'Save reminder'}
                        </Button>
                        <Button variant="outline" size="sm" onclick={() => reminderMode = 'idle'}>Cancel</Button>
                    </div>
                </Card.Content>
            {/if}
        </Card.Root>
    {/if}

    <Separator class="my-6"/>

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
