<script>
  import { navigate, Router, Link, Route } from "svelte-routing";
  import { projectStore, addProject } from "./stores.js";
  import DeleteProject from "./DeleteProject.svelte";

  let title = "";

  $: isValidTitle = validateTitle(title);

  function validateTitle(title) {
    return title && title.length >= 3;
  }

  function back() {
    navigate("/", { replace: true });
  }

  function add() {
    let project = {
      title: title,
      active: true,
    };

    addProject(project);

    title = "";
  }
</script>

<style>

</style>

<Router>
  <Route path=":id/delete" let:params>
    <DeleteProject id={params.id} />
  </Route>
  <Route>
    <h1 class="title is-1">Projects</h1>

    <div class="columns is-multiline">
      <div class=" column is-12">

        <div class="field">
          <label class="label">Name</label>

          <div class="field has-addons">
            <div class="control">
              <input
                class="input"
                minlength="3"
                maxlength="200"
                type="text"
                bind:value={title} />
            </div>
            <div class="control">
              <button
                class="button is-success"
                disabled={!isValidTitle}
                on:click={add}>
                <span class="icon is-medium">
                  <i class="fa fa-plus" />
                </span>
              </button>
            </div>
          </div>
        </div>

      </div>

      {#each $projectStore as project}
        <div class=" column is-12">
          <div class="card" title={project.title}>
            <header class="card-header">
              <p class="card-header-title">{project.title}</p>

              {#if project._links.delete}
              <a href="/projects/{project.id}/delete" class="card-header-icon">
                <span class="icon">
                  <i class="fas fa-trash" aria-hidden="true" />
                </span>
              </a>
              {/if}
            </header>

          </div>
        </div>
      {/each}

      <div class=" column is-12">
        <div class="field is-grouped">
          <p class="control">
            <button class="button" on:click={back}>Back</button>
          </p>
        </div>
      </div>

    </div>
  </Route>
</Router>
