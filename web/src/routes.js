import BatchList from './pages/BatchList.svelte'
import BatchDetail from './pages/BatchDetail.svelte'
import NewBatch from './pages/NewBatch.svelte'

export const routes = {
  '/': BatchList,
  '/batches/new': NewBatch,
  '/batches/:id': BatchDetail,
}
