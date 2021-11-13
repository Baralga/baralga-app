<script>
  import { Router, Route } from "svelte-routing";
  import EditActivity from "./EditActivity.svelte";
</script>

<style>
</style>

<Router>
  <Route path="/new">
    <EditActivity />
  </Route>
  <Route path="/:id/edit" let:params>
    <EditActivity id={params.id} />
  </Route>
</Router>
