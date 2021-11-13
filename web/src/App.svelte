<script>
  import { Router, Route } from "svelte-routing";
  import Activities from "./Activities.svelte";
  import Projects from "./Projects.svelte";
  import Home from "./Home.svelte";
  import Backup from "./Backup.svelte";
  import Login from "./Login.svelte";
</script>

<style>

</style>

<Router>
  <Route path="/activities/*" component={Activities} />
  <Route path="/projects/*" component={Projects} />
  <Route path="/backup" component="{Backup}"/>
  <Route path="/" component="{Home}"/>
  <Route path="/login" component="{Login}"/>
</Router>
