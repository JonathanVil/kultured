<script>
  import Router from 'svelte-spa-router'
  import { routes } from './routes.js'
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="container mx-auto max-w-4xl px-4 py-8">
    <Router {routes} />
  </div>
</div>
