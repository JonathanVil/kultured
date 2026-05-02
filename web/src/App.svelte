<script>
  import Router from 'svelte-spa-router'
  import { routes } from './routes.js'
</script>

<div class="min-h-screen bg-background text-foreground">
  <header class="border-b border-border/60 bg-card/80 backdrop-blur-sm sticky top-0 z-10">
    <div class="container mx-auto max-w-4xl px-4 h-12 flex items-center">
      <a href="#/" class="flex items-center gap-2 text-foreground no-underline">
        <span class="text-lg">🫙</span>
        <span class="font-semibold tracking-tight text-primary">kultured</span>
      </a>
    </div>
  </header>
  <main class="container mx-auto max-w-4xl px-4 py-8">
    <Router {routes} />
  </main>
</div>
